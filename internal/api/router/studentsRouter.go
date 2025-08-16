package router

import (
	hnd "WebProject/internal/api/handlers"
	"net/http"
)

func StudentsRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /students", hnd.GetStudentsHandler)
	mux.HandleFunc("GET /students/{id}", hnd.GetStudentHandler)
	mux.HandleFunc("POST /students", hnd.AddStudentHandler)
	mux.HandleFunc("PUT /students/{id}", hnd.UpdateStudentHandler)
	mux.HandleFunc("PATCH /students", hnd.PatchStudentsHandler)
	mux.HandleFunc("PATCH /students/{id}", hnd.PatchStudentHandler)
	mux.HandleFunc("DELETE /students", hnd.DeleteStudentsHandler)
	mux.HandleFunc("DELETE /students/{id}", hnd.DeleteStudentHandler)
	
	return mux
}
