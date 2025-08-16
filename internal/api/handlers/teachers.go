package handlers

import (
	mod "WebProject/internal/models"
	sqlc "WebProject/internal/repos/sqlconnect"
	"WebProject/pkg/utils"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func GetTeachersHandler(w http.ResponseWriter, r *http.Request) {

	teacherList, err := sqlc.GetAllTeachers(r)
	if err != nil {
		return
	}

	response := struct {
		Status string        `json:"status"`
		Count  int           `json:"count"`
		Data   []mod.Teacher `json:"data"`
	}{
		Status: "success",
		Count:  len(teacherList),
		Data:   teacherList,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func GetTeacherHandler(w http.ResponseWriter, r *http.Request) {

	path := r.PathValue("id")

	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	var teacher mod.Teacher

	err = sqlc.FindTeacherById(err, id, teacher)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teacher)
}

func AddTeacherHandler(w http.ResponseWriter, r *http.Request) {
	_, err := utils.AuthorizeUser(r.Context().Value(utils.ContextKey("role")).(string), "admin")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	addedTeachers, err := sqlc.SaveTeachers(r)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string        `json:"status"`
		Count  int           `json:"count"`
		Data   []mod.Teacher `json:"data"`
	}{
		Status: "success",
		Count:  len(addedTeachers),
		Data:   addedTeachers,
	}
	json.NewEncoder(w).Encode(response)
}

func UpdateTeacherHandler(w http.ResponseWriter, r *http.Request) {
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
	var updatedTeacher mod.Teacher
	err = json.NewDecoder(r.Body).Decode(&updatedTeacher)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	updatedTeacherDB, err := sqlc.UpdateTeacherById(err, id, updatedTeacher)

	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTeacherDB)

}

func PatchTeacherHandler(w http.ResponseWriter, r *http.Request) {
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

	existingTeacher, err := sqlc.PatchTeacherById(err, id, updates)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingTeacher)

}

func PatchTeachersHandler(w http.ResponseWriter, r *http.Request) {
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

	err = sqlc.PatchAllTeachers(err, updates)
	if err != nil {
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

func DeleteTeacherHandler(w http.ResponseWriter, r *http.Request) {
	_, err := utils.AuthorizeUser(r.Context().Value(utils.ContextKey("role")).(string), "admin", "manager")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/teachers")
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

	err = sqlc.DeleteTeacherById(err, id)
	if err != nil {
		return
	}

	w.WriteHeader(http.StatusNoContent)

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status string
		ID     int
	}{
		Status: "Teacher deleted",
		ID:     id,
	}
	json.NewEncoder(w).Encode(response)
}

func DeleteTeachersHandler(w http.ResponseWriter, r *http.Request) {
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

	deletedIdsFromBd, err := sqlc.DeleteTeachers(err, ids)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status string
		IDs    []int
	}{
		Status: "Teachers deleted",
		IDs:    deletedIdsFromBd,
	}
	json.NewEncoder(w).Encode(response)
}

func GetStudentsByTeacherHandler(w http.ResponseWriter, r *http.Request) {

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

	students, err := sqlc.FindStudentsByTeacherId(w, err, id)
	if err != nil {
		return
	}

	resp := struct {
		Status string        `json:"status"`
		Count  int           `json:"count"`
		Data   []mod.Student `json:"data"`
	}{
		Status: "success",
		Count:  len(students),
		Data:   students,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
