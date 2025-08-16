package handlers

import (
	"WebProject/internal/models"
	sqlc "WebProject/internal/repos/sqlconnect"
	"WebProject/pkg/utils"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func GetExecsHandler(w http.ResponseWriter, r *http.Request) {

	ExecList, err := sqlc.GetAllExecs(r)
	if err != nil {
		return
	}

	response := struct {
		Status string        `json:"status"`
		Count  int           `json:"count"`
		Data   []models.Exec `json:"data"`
	}{
		Status: "success",
		Count:  len(ExecList),
		Data:   ExecList,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func GetExecHandler(w http.ResponseWriter, r *http.Request) {

	path := r.PathValue("id")

	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	var exec models.Exec

	exec, err = sqlc.FindExecById(err, id, exec)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exec)
}

func AddExecHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(utils.ContextKey("role"))
	fmt.Println(r.Context().Value(utils.ContextKey("role")).(string))
	_, err := utils.AuthorizeUser(r.Context().Value(utils.ContextKey("role")).(string), "admin")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	addedExecs, err := sqlc.SaveExecs(r)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string        `json:"status"`
		Count  int           `json:"count"`
		Data   []models.Exec `json:"data"`
	}{
		Status: "success",
		Count:  len(addedExecs),
		Data:   addedExecs,
	}
	json.NewEncoder(w).Encode(response)
}

func PatchExecHandler(w http.ResponseWriter, r *http.Request) {
	_, err := utils.AuthorizeUser(r.Context().Value(utils.ContextKey("role")).(string), "admin")
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

	existingExec, err := sqlc.PatchExecById(err, id, updates)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingExec)

}

func DeleteExecHandler(w http.ResponseWriter, r *http.Request) {
	_, err := utils.AuthorizeUser(r.Context().Value(utils.ContextKey("role")).(string), "admin")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	path := r.PathValue("id")
	if path == "" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = sqlc.DeleteExecById(err, id)
	if err != nil {
		return
	}

	w.WriteHeader(http.StatusNoContent)

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status string
		ID     int
	}{
		Status: "Exec deleted",
		ID:     id,
	}
	json.NewEncoder(w).Encode(response)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	//validation json
	var req models.Exec
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Username == "" || req.Password == "" {
		http.Error(w, "Invalid username or password", http.StatusBadRequest)
		return
	}

	//verify user
	err, user := sqlc.GetUserByUsername(req.Username)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusBadRequest)
		return
	}

	//verify activity
	if user.InactiveStatus {
		http.Error(w, "User is inactive", http.StatusForbidden)
		return
	}

	//verify password
	err = utils.VerifyPassword(user.Password, req.Password)
	if err != nil {
		http.Error(w, "Invalid password", http.StatusForbidden)
		return
	}

	//generate token
	tokenString, err := utils.SignToken(user.ID, user.Username, user.Role)
	if err != nil {
		http.Error(w, "Cannot create token", http.StatusInternalServerError)
		return
	}

	//set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(24 * time.Hour),
		SameSite: http.SameSiteStrictMode,
	})

	//response body
	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Token string `json:"token"`
	}{
		Token: tokenString,
	}
	json.NewEncoder(w).Encode(response)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Unix(0, 0),
		SameSite: http.SameSiteStrictMode,
	})
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message" : "Logout Successful"}`))
}

func UpdatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("id")
	userId, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	var req models.UpdatePasswordRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.CurrentPassword == "" || req.NewPassword == "" {
		http.Error(w, "Required password", http.StatusBadRequest)
		return
	}

	token, err := sqlc.UpdatePasswordById(userId, req)

	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(24 * time.Hour),
		SameSite: http.SameSiteStrictMode,
	})

	response := struct {
		Token string `json:"token"`
	}{
		Token: token,
	}
	json.NewEncoder(w).Encode(response)

}

func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.ErrorHandler(err, "Invalid JSON")
		return
	}
	defer r.Body.Close()
	sqlc.ForgotPasswordDB(req.Email)
}

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	resetCode := r.PathValue("resetcode")
	if resetCode == "" {
		utils.ErrorHandler(nil, "Invalid resetcode")
		return
	}

	var req struct {
		NewPassword string `json:"newpassword"`
		Confirm     string `json:"confirm"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.ErrorHandler(err, "Invalid JSON")
		return
	}
	defer r.Body.Close()

	if req.Confirm == "" || req.NewPassword == "" {
		utils.ErrorHandler(nil, "Invalid JSON")
		return
	}

	if req.NewPassword != req.Confirm {
		utils.ErrorHandler(nil, "Invalid password")
		return
	}

	tokenBytes, err := hex.DecodeString(resetCode)
	if err != nil {
		utils.ErrorHandler(err, "Innvalid resetcode")
	}
	hashedToken := sha256.Sum256([]byte(tokenBytes))
	hash := hex.EncodeToString(hashedToken[:])

	err = sqlc.PasswordResetDB(req.NewPassword, hash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := struct {
		Message string `json:"message"`
	}{
		Message: "Password updated successfully",
	}
	json.NewEncoder(w).Encode(response)

}
