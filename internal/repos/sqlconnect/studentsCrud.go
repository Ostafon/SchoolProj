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

func GetAllStudents(r *http.Request) ([]mod.Student, error) {
	query := "SELECT * FROM Students WHERE 1=1"
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

	StudentList := make([]mod.Student, 0)

	for rows.Next() {
		var Student mod.Student
		err := rows.Scan(utils.GetStructFields(&Student, true, true)...)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error scanning DB")
		}
		StudentList = append(StudentList, Student)
	}
	return StudentList, nil
}

// FindStudentById — найти студента по ID
func FindStudentById(err error, id int, Student mod.Student) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "Error connecting to DB")
	}
	defer db.Close()

	err = db.QueryRow(
		utils.GenerateSQL(mod.Student{}, "select"),
		id,
	).Scan(utils.GetStructFields(&Student, true, true)...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return utils.ErrorHandler(err, "Student not found")
		}
		return utils.ErrorHandler(err, "Error querying DB")
	}

	return nil
}

// SaveStudents — вставка новых студентов из JSON
func SaveStudents(r *http.Request) ([]mod.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to DB")
	}
	defer db.Close()

	var newStudents []mod.Student
	err = json.NewDecoder(r.Body).Decode(&newStudents)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error decoding JSON")
	}

	stmt, err := db.Prepare(utils.GenerateSQL(mod.Student{}, "insert"))
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error preparing statement")
	}
	defer stmt.Close()

	addedStudents := make([]mod.Student, len(newStudents))
	for i, Student := range newStudents {
		res, err := stmt.Exec(utils.GetStructFields(Student, true, false)...)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error inserting Student")
		}
		lastId, err := res.LastInsertId()
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error getting last insert ID")
		}
		Student.ID = int(lastId)
		addedStudents[i] = Student
	}
	return addedStudents, nil
}

// UpdateStudentById — полное обновление студента по ID
func UpdateStudentById(err error, id int, updatedStudent mod.Student) (mod.Student, error) {
	db, err := ConnectDB()
	if err != nil {
		return mod.Student{}, utils.ErrorHandler(err, "Error connecting to DB")
	}
	defer db.Close()

	var existingStudent mod.Student
	err = db.QueryRow(utils.GenerateSQL(mod.Student{}, "select"), id).Scan(utils.GetStructFields(&existingStudent, true, true)...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return mod.Student{}, utils.ErrorHandler(err, "Student not found")
		}
		return mod.Student{}, utils.ErrorHandler(err, "Error querying Student")
	}

	updatedStudent.ID = existingStudent.ID
	fields := utils.GetStructFields(updatedStudent, false, false)
	fields = append(fields, updatedStudent.ID) // для WHERE id = ?

	_, err = db.Exec(utils.GenerateSQL(mod.Student{}, "update"), fields...)
	if err != nil {
		return mod.Student{}, utils.ErrorHandler(err, "Error updating Student")
	}
	return updatedStudent, nil
}

// PatchStudentById — частичное обновление по ID
func PatchStudentById(err error, id int, updates map[string]interface{}) (mod.Student, error) {
	var existingStudent mod.Student

	db, err := ConnectDB()
	if err != nil {
		return mod.Student{}, utils.ErrorHandler(err, "Error connecting to DB")
	}
	defer db.Close()

	err = db.QueryRow(utils.GenerateSQL(mod.Student{}, "select"), id).Scan(utils.GetStructFields(&existingStudent, true, true)...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return mod.Student{}, utils.ErrorHandler(err, "Student not found")
		}
		return mod.Student{}, utils.ErrorHandler(err, "Error fetching Student")
	}

	StudentVal := reflect.ValueOf(&existingStudent).Elem()
	StudentType := StudentVal.Type()

	for k, v := range updates {
		for i := 0; i < StudentVal.NumField(); i++ {
			field := StudentType.Field(i)
			if field.Tag.Get("json") == k {
				if StudentVal.Field(i).CanSet() {
					val := reflect.ValueOf(v)
					if val.Type().ConvertibleTo(StudentVal.Field(i).Type()) {
						StudentVal.Field(i).Set(val.Convert(StudentVal.Field(i).Type()))
					} else {
						return mod.Student{}, utils.ErrorHandler(errors.New("type mismatch"), "Invalid JSON value for field "+k)
					}
				}
			}
		}
	}

	fields := utils.GetStructFields(existingStudent, false, false)
	fields = append(fields, existingStudent.ID)

	_, err = db.Exec(utils.GenerateSQL(mod.Student{}, "update"), fields...)
	if err != nil {
		return mod.Student{}, utils.ErrorHandler(err, "Error updating Student")
	}
	return existingStudent, nil
}

// PatchAllStudents — частичное обновление множества студентов (транзакция)
func PatchAllStudents(err error, updates []map[string]interface{}) error {
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

		var existingStudent mod.Student
		err = tx.QueryRow(utils.GenerateSQL(mod.Student{}, "select"), id).Scan(utils.GetStructFields(&existingStudent, true, true)...)
		if err != nil {
			tx.Rollback()
			return utils.ErrorHandler(err, "Student not found or error fetching")
		}

		StudentVal := reflect.ValueOf(&existingStudent).Elem()
		StudentType := StudentVal.Type()

		for k, v := range update {
			if k == "id" {
				continue
			}
			for i := 0; i < StudentVal.NumField(); i++ {
				field := StudentType.Field(i)
				if field.Tag.Get("json") == k {
					fieldVal := StudentVal.Field(i)
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

		fields := utils.GetStructFields(existingStudent, true, false)
		fields = append(fields, id)

		_, err = tx.Exec(utils.GenerateSQL(mod.Student{}, "update"), fields...)
		if err != nil {
			tx.Rollback()
			return utils.ErrorHandler(err, "Error updating Student with ID "+strconv.Itoa(id))
		}
	}

	err = tx.Commit()
	if err != nil {
		return utils.ErrorHandler(err, "Error committing transaction")
	}
	return nil
}

// DeleteStudentById — удаление по ID
func DeleteStudentById(err error, id int) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "Error connecting to DB")
	}
	defer db.Close()

	res, err := db.Exec(utils.GenerateSQL(mod.Student{}, "delete"), id)
	if err != nil {
		return utils.ErrorHandler(err, "Error deleting Student")
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return utils.ErrorHandler(err, "Error checking deletion result")
	}
	if rows == 0 {
		return utils.ErrorHandler(errors.New("no rows affected"), "Student not found")
	}
	return nil
}

// DeleteStudents — удаление множества учителей по списку ID
func DeleteStudents(err error, ids []int) ([]int, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to DB")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error starting transaction")
	}

	stmt, err := tx.Prepare(utils.GenerateSQL(mod.Student{}, "delete"))
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
		return nil, utils.ErrorHandler(errors.New("no deletions"), "No Students were deleted")
	}
	return deletedIds, nil
}
