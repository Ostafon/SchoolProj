package router

import (
	hnd "WebProject/internal/api/handlers"
	"net/http"
)

func TeacherRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /teachers", hnd.GetTeachersHandler)
	mux.HandleFunc("GET /teachers/{id}", hnd.GetTeacherHandler)
	mux.HandleFunc("POST /teachers", hnd.AddTeacherHandler)
	mux.HandleFunc("PUT /teachers/{id}", hnd.UpdateTeacherHandler)
	mux.HandleFunc("PATCH /teachers", hnd.PatchTeachersHandler)
	mux.HandleFunc("PATCH /teachers/{id}", hnd.PatchTeacherHandler)
	mux.HandleFunc("DELETE /teachers", hnd.DeleteTeachersHandler)
	mux.HandleFunc("DELETE /teachers/{id}", hnd.DeleteTeacherHandler)
	mux.HandleFunc("GET /teachers/{id}/students", hnd.GetStudentsByTeacherHandler)

	return mux
}
