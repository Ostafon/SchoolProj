package sqlconnect

import (
	mod "WebProject/internal/models"
	"WebProject/pkg/utils"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strconv"
)

// GetAllTeachers — получаем список учителей с фильтрами и сортировкой
func GetAllTeachers(r *http.Request) ([]mod.Teacher, error) {
	query := "SELECT * FROM teachers WHERE 1=1"
	var args []interface{}

	query, args = utils.AddFilters(r, query, args)
	query = utils.AddSorting(r, query)

	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to DB")
	}
	defer db.Close()

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error querying DB")
	}
	defer rows.Close()

	teacherList := make([]mod.Teacher, 0)

	for rows.Next() {
		var teacher mod.Teacher
		err := rows.Scan(utils.GetStructFields(&teacher, true, true)...)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error scanning DB")
		}
		teacherList = append(teacherList, teacher)
	}
	return teacherList, nil
}

// FindTeacherById — найти учителя по ID
func FindTeacherById(err error, id int, teacher mod.Teacher) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "Error connecting to DB")
	}
	defer db.Close()

	err = db.QueryRow(
		utils.GenerateSQL(mod.Student{}, "select"),
		id,
	).Scan(utils.GetStructFields(teacher, true, true)...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return utils.ErrorHandler(err, "Student not found")
		}
		return utils.ErrorHandler(err, "Error querying DB")
	}

	return nil
}

// SaveTeachers — вставка новых учителей из JSON
func SaveTeachers(r *http.Request) ([]mod.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to DB")
	}
	defer db.Close()

	var newTeachers []mod.Teacher
	err = json.NewDecoder(r.Body).Decode(&newTeachers)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error decoding JSON")
	}

	stmt, err := db.Prepare(utils.GenerateSQL(mod.Teacher{}, "insert"))
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error preparing statement")
	}
	defer stmt.Close()

	addedTeachers := make([]mod.Teacher, len(newTeachers))
	for i, teacher := range newTeachers {
		res, err := stmt.Exec(utils.GetStructFields(teacher, false, false)...)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error inserting teacher")
		}
		lastId, err := res.LastInsertId()
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error getting last insert ID")
		}
		teacher.ID = int(lastId)
		addedTeachers[i] = teacher
	}
	return addedTeachers, nil
}

// UpdateTeacherById — полное обновление учителя по ID
func UpdateTeacherById(err error, id int, updatedTeacher mod.Teacher) (mod.Teacher, error) {
	db, err := ConnectDB()
	if err != nil {
		return mod.Teacher{}, utils.ErrorHandler(err, "Error connecting to DB")
	}
	defer db.Close()

	var existingTeacher mod.Teacher
	err = db.QueryRow(utils.GenerateSQL(mod.Teacher{}, "select"), id).Scan(utils.GetStructFields(&existingTeacher, true, true)...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return mod.Teacher{}, utils.ErrorHandler(err, "Teacher not found")
		}
		return mod.Teacher{}, utils.ErrorHandler(err, "Error querying teacher")
	}

	updatedTeacher.ID = existingTeacher.ID
	fields := utils.GetStructFields(updatedTeacher, false, false)
	fields = append(fields, updatedTeacher.ID) // для WHERE id = ?

	_, err = db.Exec(utils.GenerateSQL(mod.Teacher{}, "update"), fields...)
	if err != nil {
		return mod.Teacher{}, utils.ErrorHandler(err, "Error updating teacher")
	}
	return updatedTeacher, nil
}

// PatchTeacherById — частичное обновление по ID
func PatchTeacherById(err error, id int, updates map[string]interface{}) (mod.Teacher, error) {
	var existingTeacher mod.Teacher

	db, err := ConnectDB()
	if err != nil {
		return mod.Teacher{}, utils.ErrorHandler(err, "Error connecting to DB")
	}
	defer db.Close()

	err = db.QueryRow(utils.GenerateSQL(mod.Teacher{}, "select"), id).Scan(utils.GetStructFields(&existingTeacher, true, true)...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return mod.Teacher{}, utils.ErrorHandler(err, "Teacher not found")
		}
		return mod.Teacher{}, utils.ErrorHandler(err, "Error fetching teacher")
	}

	teacherVal := reflect.ValueOf(&existingTeacher).Elem()
	teacherType := teacherVal.Type()

	for k, v := range updates {
		for i := 0; i < teacherVal.NumField(); i++ {
			field := teacherType.Field(i)
			if field.Tag.Get("json") == k {
				if teacherVal.Field(i).CanSet() {
					val := reflect.ValueOf(v)
					if val.Type().ConvertibleTo(teacherVal.Field(i).Type()) {
						teacherVal.Field(i).Set(val.Convert(teacherVal.Field(i).Type()))
					} else {
						return mod.Teacher{}, utils.ErrorHandler(errors.New("type mismatch"), "Invalid JSON value for field "+k)
					}
				}
			}
		}
	}

	fields := utils.GetStructFields(existingTeacher, false, false)
	fields = append(fields, existingTeacher.ID)

	_, err = db.Exec(utils.GenerateSQL(mod.Teacher{}, "update"), fields...)
	if err != nil {
		return mod.Teacher{}, utils.ErrorHandler(err, "Error updating teacher")
	}
	return existingTeacher, nil
}

// PatchAllTeachers — частичное обновление множества учителей (транзакция)
func PatchAllTeachers(err error, updates []map[string]interface{}) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "Error connecting to DB")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return utils.ErrorHandler(err, "Error starting transaction")
	}

	for _, update := range updates {
		var id int
		switch v := update["id"].(type) {
		case string:
			id, err = strconv.Atoi(v)
			if err != nil {
				tx.Rollback()
				return utils.ErrorHandler(err, "Invalid ID format")
			}
		case float64:
			id = int(v)
		default:
			tx.Rollback()
			return utils.ErrorHandler(errors.New("missing or invalid id"), "Invalid ID")
		}

		var existingTeacher mod.Teacher
		err = tx.QueryRow(utils.GenerateSQL(mod.Teacher{}, "select"), id).Scan(utils.GetStructFields(&existingTeacher, true, true)...)
		if err != nil {
			tx.Rollback()
			return utils.ErrorHandler(err, "Teacher not found or error fetching")
		}

		teacherVal := reflect.ValueOf(&existingTeacher).Elem()
		teacherType := teacherVal.Type()

		for k, v := range update {
			if k == "id" {
				continue
			}
			for i := 0; i < teacherVal.NumField(); i++ {
				field := teacherType.Field(i)
				if field.Tag.Get("json") == k {
					fieldVal := teacherVal.Field(i)
					if fieldVal.CanSet() {
						val := reflect.ValueOf(v)
						if val.Type().ConvertibleTo(fieldVal.Type()) {
							fieldVal.Set(val.Convert(fieldVal.Type()))
						} else {
							tx.Rollback()
							return utils.ErrorHandler(errors.New("type mismatch"), "Invalid JSON value for field "+k)
						}
					}
					break
				}
			}
		}

		fields := utils.GetStructFields(existingTeacher, false, false)
		fields = append(fields, id)

		_, err = tx.Exec(utils.GenerateSQL(mod.Teacher{}, "update"), fields...)
		if err != nil {
			tx.Rollback()
			return utils.ErrorHandler(err, "Error updating teacher with ID "+strconv.Itoa(id))
		}
	}

	err = tx.Commit()
	if err != nil {
		return utils.ErrorHandler(err, "Error committing transaction")
	}
	return nil
}

// DeleteTeacherById — удаление по ID
func DeleteTeacherById(err error, id int) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "Error connecting to DB")
	}
	defer db.Close()

	res, err := db.Exec(utils.GenerateSQL(mod.Teacher{}, "delete"), id)
	if err != nil {
		return utils.ErrorHandler(err, "Error deleting teacher")
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return utils.ErrorHandler(err, "Error checking deletion result")
	}
	if rows == 0 {
		return utils.ErrorHandler(errors.New("no rows affected"), "Teacher not found")
	}
	return nil
}

// DeleteTeachers — удаление множества учителей по списку ID
func DeleteTeachers(err error, ids []int) ([]int, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to DB")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error starting transaction")
	}

	stmt, err := tx.Prepare(utils.GenerateSQL(mod.Teacher{}, "delete"))
	if err != nil {
		tx.Rollback()
		return nil, utils.ErrorHandler(err, "Error preparing delete statement")
	}
	defer stmt.Close()

	var deletedIds []int
	for _, id := range ids {
		res, err := stmt.Exec(id)
		if err != nil {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Error executing delete")
		}
		rowsAf, err := res.RowsAffected()
		if err != nil {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Error checking rows affected")
		}
		if rowsAf > 0 {
			deletedIds = append(deletedIds, id)
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error committing transaction")
	}
	if len(deletedIds) < 1 {
		return nil, utils.ErrorHandler(errors.New("no deletions"), "No teachers were deleted")
	}
	return deletedIds, nil
}

// FindStudentsByTeacherId - нахождение студентов по классу у определенного учителя
func FindStudentsByTeacherId(w http.ResponseWriter, err error, id int) ([]mod.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to DB")
	}
	defer db.Close()

	var class string
	err = db.QueryRow("Select class from teachers where id = ?", id).Scan(&class)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.ErrorHandler(errors.New("no teachers found"), "Teacher not found")
		}
		http.Error(w, "DB query error", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "Error querying DB")
	}

	rows, err := db.Query("SELECT * from students where class = ?", class)
	if err != nil {
		http.Error(w, "DB query error", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "Error querying DB")
	}
	defer rows.Close()

	students := make([]mod.Student, 0)
	for rows.Next() {
		var s mod.Student
		err = rows.Scan(&s.ID, &s.FirstName, &s.LastName, &s.Email, &s.Class)
		if err != nil {
			http.Error(w, "DB scan error", http.StatusInternalServerError)
			return nil, utils.ErrorHandler(err, "Error querying DB")
		}
		students = append(students, s)
	}

	err = rows.Err()
	if err != nil {
		http.Error(w, "DB rows error", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "Error querying DB")
	}
	return students, nil
}
