package router

import (
	hnd "WebProject/internal/api/handlers"
	"net/http"
)

func ExecsRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /execs", hnd.GetExecsHandler)
	mux.HandleFunc("POST /execs", hnd.AddExecHandler)

	mux.HandleFunc("GET /execs/{id}", hnd.GetExecHandler)
	mux.HandleFunc("PATCH /execs/{id}", hnd.PatchExecHandler)
	mux.HandleFunc("DELETE /execs/{id}", hnd.DeleteExecHandler)

	mux.HandleFunc("POST /execs/login", hnd.LoginHandler)
	mux.HandleFunc("POST /execs/logout", hnd.LogoutHandler)
	mux.HandleFunc("POST /execs/{id}/updatepassword", hnd.UpdatePasswordHandler)
	mux.HandleFunc("POST /execs/forgotpassword", hnd.ForgotPasswordHandler)
	mux.HandleFunc("POST /execs/resetpassword/reset/{resetcode}", hnd.ResetPasswordHandler)

	return mux
}
