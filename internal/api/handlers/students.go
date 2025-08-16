package handlers

import (
	mod "WebProject/internal/models"
	sqlc "WebProject/internal/repos/sqlconnect"
	"WebProject/pkg/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func GetStudentsHandler(w http.ResponseWriter, r *http.Request) {

	StudentList, err := sqlc.GetAllStudents(r)
	if err != nil {
		return
	}

	response := struct {
		Status string        `json:"status"`
		Count  int           `json:"count"`
		Data   []mod.Student `json:"data"`
	}{
		Status: "success",
		Count:  len(StudentList),
		Data:   StudentList,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func GetStudentHandler(w http.ResponseWriter, r *http.Request) {

	path := r.PathValue("id")

	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	var Student mod.Student

	err = sqlc.FindStudentById(err, id, Student)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Student)
}

func AddStudentHandler(w http.ResponseWriter, r *http.Request) {

	_, err := utils.AuthorizeUser(r.Context().Value(utils.ContextKey("role")).(string), "admin", "manager")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	addedStudents, err := sqlc.SaveStudents(r)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string        `json:"status"`
		Count  int           `json:"count"`
		Data   []mod.Student `json:"data"`
	}{
		Status: "success",
		Count:  len(addedStudents),
		Data:   addedStudents,
	}
	json.NewEncoder(w).Encode(response)
}

func UpdateStudentHandler(w http.ResponseWriter, r *http.Request) {
	_, err := utils.AuthorizeUser(r.Context().Value(utils.ContextKey("role")).(string), "admin", "manager")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	path := r.PathValue("id")
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Cannot read body", http.StatusBadRequest)
		return
	}
	fmt.Println("Request body:", string(bodyBytes))

	// Восстановить r.Body, так как io.ReadAll его опустошит
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var updatedStudent mod.Student
	err = json.NewDecoder(r.Body).Decode(&updatedStudent)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	updatedStudentDB, err := sqlc.UpdateStudentById(err, id, updatedStudent)

	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedStudentDB)

}

func PatchStudentHandler(w http.ResponseWriter, r *http.Request) {
	_, err := utils.AuthorizeUser(r.Context().Value(utils.ContextKey("role")).(string), "admin", "manager")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	path := r.PathValue("id")

	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	existingStudent, err := sqlc.PatchStudentById(err, id, updates)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingStudent)

}

func PatchStudentsHandler(w http.ResponseWriter, r *http.Request) {
	_, err := utils.AuthorizeUser(r.Context().Value(utils.ContextKey("role")).(string), "admin", "manager")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var updates []map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err = sqlc.PatchAllStudents(err, updates)
	if err != nil {
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

func DeleteStudentHandler(w http.ResponseWriter, r *http.Request) {

	_, err := utils.AuthorizeUser(r.Context().Value(utils.ContextKey("role")).(string), "admin", "manager")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/Students")
	path = strings.Trim(path, "/")
	if path == "" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = sqlc.DeleteStudentById(err, id)
	if err != nil {
		return
	}

	w.WriteHeader(http.StatusNoContent)

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status string
		ID     int
	}{
		Status: "Student deleted",
		ID:     id,
	}
	json.NewEncoder(w).Encode(response)
}

func DeleteStudentsHandler(w http.ResponseWriter, r *http.Request) {
	_, err := utils.AuthorizeUser(r.Context().Value(utils.ContextKey("role")).(string), "admin", "manager")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var ids []int

	err = json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	deletedIdsFromBd, err := sqlc.DeleteStudents(err, ids)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status string
		IDs    []int
	}{
		Status: "Students deleted",
		IDs:    deletedIdsFromBd,
	}
	json.NewEncoder(w).Encode(response)
}
