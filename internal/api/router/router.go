package router

import (
	"net/http"
)

func MainRouter() *http.ServeMux {

	tRout := TeacherRouter()
	sRout := StudentsRouter()
	eRout := ExecsRouter()

	sRout.Handle("/", eRout)
	tRout.Handle("/", sRout)

	return tRout
}
