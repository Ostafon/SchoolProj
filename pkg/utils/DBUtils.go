package utils

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

// addSorting — добавляет ORDER BY в запрос из параметров ?sortBy=field:asc
func AddSorting(r *http.Request, query string) string {
	sortParams := r.URL.Query()["sortBy"]
	if len(sortParams) > 0 {
		query += " ORDER BY"
		for i, param := range sortParams {
			parts := strings.Split(param, ":")
			if len(parts) != 2 {
				continue
			}
			field, order := parts[0], parts[1]
			if !IsValidSortOrder(order) || !IsValidSortField(field) {
				continue
			}
			if i > 0 {
				query += ","
			}
			query += fmt.Sprintf(" %s %s", field, order)
		}
	}
	return query
}

func IsValidSortOrder(sort string) bool {
	return sort == "asc" || sort == "desc"
}

func IsValidSortField(field string) bool {
	validFields := map[string]bool{
		"firstName": true,
		"lastName":  true,
		"email":     true,
		"class":     true,
		"subject":   true,
	}
	return validFields[field]
}

// addFilters — добавляет WHERE фильтры по параметрам
func AddFilters(r *http.Request, query string, args []interface{}) (string, []interface{}) {
	params := map[string]string{
		"firstName": "firstName",
		"lastName":  "lastName",
		"email":     "email",
		"class":     "class",
		"subject":   "subject",
	}
	for k, dbField := range params {
		value := r.URL.Query().Get(k)
		if value != "" {
			query += " AND " + dbField + " = ?"
			args = append(args, value)
		}
	}
	return query, args
}

// getDBFields — получает список db тегов без id (только для insert/update)
func getDBFields(model interface{}) []string {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	var fields []string
	for i := 0; i < t.NumField(); i++ {
		dbTag := t.Field(i).Tag.Get("db")
		if dbTag != "" && dbTag != "id" {
			fields = append(fields, dbTag)
		}
	}
	return fields
}

// GenerateSQL — универсальный генератор SQL запросов по модели и типу запроса
func GenerateSQL(model interface{}, queryType string) string {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	tableName := strings.ToLower(t.Name()) + "s"
	if tableName == "execdtos" {
		tableName = "execs"
	}

	var fields []string
	for i := 0; i < t.NumField(); i++ {
		dbTag := t.Field(i).Tag.Get("db")
		if dbTag != "" && dbTag != "-" {
			fields = append(fields, dbTag)
		}
	}

	if len(fields) == 0 {
		return ""
	}

	switch strings.ToLower(queryType) {
	case "insert":
		placeholders := strings.Repeat("?, ", len(fields))
		placeholders = strings.TrimSuffix(placeholders, ", ")
		return "INSERT " + "INTO " + tableName +
			" (" + strings.Join(fields, ", ") + ")" +
			" VALUES (" + placeholders + ")"

	case "update":
		var setParts []string
		for _, f := range fields {
			if f != "id" {
				setParts = append(setParts, f+" = ?")
			}
		}
		return "UPDATE " + tableName +
			" SET " + strings.Join(setParts, ", ") +
			" WHERE id = ?"

	case "delete":
		fmt.Println("Deleting")
		return "DELETE " + "FROM " + tableName + " WHERE id = ?"

	case "select":
		return "SELECT " + strings.Join(fields, ", ") +
			" FROM " + tableName + " WHERE id = ?"
	default:
		return ""
	}
}

// GetStructFields — получает слайс интерфейсов для Scan или Exec по модели
// includeID — включать ли id поле
// forScan — возвращать адреса для Scan (true) или значения для Exec (false)
func GetStructFields(model interface{}, includeID bool, forScan bool) []interface{} {
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	if forScan {
		if reflect.ValueOf(model).Kind() != reflect.Ptr {
			panic("forScan=true requires pointer to struct")
		}
	}

	var fields []interface{}
	for i := 0; i < t.NumField(); i++ {
		dbTag := t.Field(i).Tag.Get("db")
		if dbTag == "" {
			continue
		}
		if dbTag == "id" && !includeID {
			continue
		}

		fieldVal := v.Field(i)
		if forScan {
			fields = append(fields, fieldVal.Addr().Interface())
		} else {
			fields = append(fields, fieldVal.Interface())
		}
	}
	return fields
}
