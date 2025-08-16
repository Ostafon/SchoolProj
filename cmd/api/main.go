package main

import (
	mw "WebProject/internal/api/middlewares"
	"WebProject/internal/api/router"
	"WebProject/internal/repos/sqlconnect"
	"fmt"
	"net/http"
	"os"
)

func main() {

	_, err := sqlconnect.ConnectDB()
	if err != nil {
		panic(err)
	}

	//rl := mw.NewRateLimiter(5, time.Minute)
	//hpp := mw.HPPOptions{
	//	CheckQuery:          true,
	//	CheckBody:           true,
	//	CheckOnlyForContent: "application/x-www-form-urlencoded",
	//	Whitelist:           []string{"sortBy", "sortOrder", "class", "age", "name"},
	//}

	jwtMiddleware := mw.MiddlewaresExcludeRoute(mw.JWTMiddleware, "/execs/login", "/execs/forgotpassword", "/execs/resetpassword/reset")
	secureMux := jwtMiddleware(mw.SecurityHeaders(router.MainRouter()))
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", os.Getenv("API_PORT")),
		Handler: secureMux,
	}

	fmt.Printf("Starting server on port %d\n", os.Getenv("API_PORT"))
	server.ListenAndServe()
}
