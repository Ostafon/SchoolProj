package sqlconnect

import (
	model "WebProject/internal/models"
	"WebProject/pkg/utils"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-mail/mail/v2"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"
)

func GetAllExecs(r *http.Request) ([]model.Exec, error) {
	query := "SELECT id, firstname, lastname, email, username,  usercreatedat, inactivestatus, role FROM execs WHERE 1=1"
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

	ExecList := make([]model.Exec, 0)

	for rows.Next() {
		var Exec model.Exec
		err := rows.Scan(&Exec.ID, &Exec.FirstName, &Exec.LastName, &Exec.Email,
			&Exec.Username, &Exec.UserCreatedAt, &Exec.InactiveStatus, &Exec.Role)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error scanning DB")
		}
		ExecList = append(ExecList, Exec)
	}
	return ExecList, nil
}

func FindExecById(err error, id int, Exec model.Exec) (model.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return model.Exec{}, utils.ErrorHandler(err, "Error connecting to DB")
	}
	defer db.Close()

	err = db.QueryRow(
		"SELECT id, firstname, lastname, email, username,  usercreatedat, inactivestatus, role FROM execs WHERE id = ?",
		id,
	).Scan(&Exec.ID, &Exec.FirstName, &Exec.LastName, &Exec.Email,
		&Exec.Username, &Exec.UserCreatedAt, &Exec.InactiveStatus, &Exec.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Exec{}, utils.ErrorHandler(err, "Exec not found")
		}
		return model.Exec{}, utils.ErrorHandler(err, "Error querying DB")
	}

	return Exec, nil
}

func SaveExecs(r *http.Request) ([]model.Exec, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to DB")
	}
	defer db.Close()

	var newExecs []model.Exec
	err = json.NewDecoder(r.Body).Decode(&newExecs)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error decoding JSON")
	}

	stmt, err := db.Prepare(utils.GenerateSQL(model.Exec{}, "insert"))
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error preparing statement")
	}
	defer stmt.Close()

	addedExecs := make([]model.Exec, len(newExecs))
	for i, Exec := range newExecs {
		if Exec.Password == "" {
			return nil, utils.ErrorHandler(err, "Enter valid password")
		}

		err, encodedPass := utils.PasswordHashing(Exec.Password)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error hashing password")
		}

		Exec.Password = encodedPass

		res, err := stmt.Exec(utils.GetStructFields(Exec, true, false)...)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error inserting Exec")
		}
		lastId, err := res.LastInsertId()
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error getting last insert ID")
		}
		Exec.ID = int(lastId)
		addedExecs[i] = Exec
	}
	return addedExecs, nil
}

// PatchExecById — частичное обновление по ID
func PatchExecById(err error, id int, updates map[string]interface{}) (model.Exec, error) {
	var existingExec model.Exec

	db, err := ConnectDB()
	if err != nil {
		return model.Exec{}, utils.ErrorHandler(err, "Error connecting to DB")
	}
	defer db.Close()

	err = db.QueryRow("SELECT id, firstname, lastname, email, username,  usercreatedat, inactivestatus, role FROM execs WHERE id = ?", id).
		Scan(&existingExec.ID, &existingExec.FirstName, &existingExec.LastName, &existingExec.Email,
			&existingExec.Username, &existingExec.UserCreatedAt, &existingExec.InactiveStatus, &existingExec.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Exec{}, utils.ErrorHandler(err, "Exec not found")
		}
		return model.Exec{}, utils.ErrorHandler(err, "Error fetching Exec")
	}

	ExecVal := reflect.ValueOf(&existingExec).Elem()
	ExecType := ExecVal.Type()

	for k, v := range updates {
		for i := 0; i < ExecVal.NumField(); i++ {
			field := ExecType.Field(i)
			if field.Tag.Get("json") == k {
				if ExecVal.Field(i).CanSet() {
					val := reflect.ValueOf(v)
					if val.Type().ConvertibleTo(ExecVal.Field(i).Type()) {
						ExecVal.Field(i).Set(val.Convert(ExecVal.Field(i).Type()))
					} else {
						return model.Exec{}, utils.ErrorHandler(errors.New("type mismatch"), "Invalid JSON value for field "+k)
					}
				}
			}
		}
	}

	fields := utils.GetStructFields(existingExec, false, false)
	fields = append(fields, existingExec.ID)

	_, err = db.Exec(utils.GenerateSQL(model.Exec{}, "update"), fields...)
	if err != nil {
		return model.Exec{}, utils.ErrorHandler(err, "Error updating Exec")
	}
	return existingExec, nil
}

// DeleteExecById — удаление по ID
func DeleteExecById(err error, id int) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "Error connecting to DB")
	}
	defer db.Close()

	res, err := db.Exec("DELETE FROM execs where id = ?", id)
	if err != nil {
		return utils.ErrorHandler(err, "Error deleting Exec")
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return utils.ErrorHandler(err, "Error checking deletion result")
	}
	if rows == 0 {
		return utils.ErrorHandler(errors.New("no rows affected"), "Exec not found")
	}
	return nil
}

func GetUserByUsername(username string) (error, *model.Exec) {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "Error connect to DB"), nil
	}
	defer db.Close()

	var user = &model.Exec{}
	err = db.QueryRow("SELECT id, firstname, lastname, email, username,password,  usercreatedat, inactivestatus, role FROM execs WHERE username = ?", username).
		Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email,
			&user.Username, &user.Password, &user.UserCreatedAt, &user.InactiveStatus, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {

			return utils.ErrorHandler(err, "User not found"), nil
		}
		return utils.ErrorHandler(err, "Error fetching User"), nil
	}
	return nil, user
}

// UpdatePasswordById обновляем пароль по определенному ID и возвращаем токен
func UpdatePasswordById(userId int, req model.UpdatePasswordRequest) (string, error) {
	db, err := ConnectDB()
	if err != nil {
		return "", utils.ErrorHandler(err, "Cannot connect to database")
	}
	defer db.Close()

	var userName string
	var curPassword string
	var uRole string

	err = db.QueryRow("SELECT username,password,role  FROM Execs WHERE id=?", userId).Scan(&userName, &curPassword, &uRole)
	if err != nil {
		return "", utils.ErrorHandler(err, "User not found")
	}

	err = utils.VerifyPassword(curPassword, req.CurrentPassword)
	if err != nil {
		return "", utils.ErrorHandler(err, "Invalid password")
	}

	err, encodedPass := utils.PasswordHashing(req.NewPassword)
	if err != nil {
		return "", utils.ErrorHandler(err, "Cannot hash password")
	}

	passwordChangedAt := time.Now().Format(time.RFC3339)

	_, err = db.Exec("UPDATE Execs SET password=?, passwordChangedAt = ? WHERE id=?", encodedPass, passwordChangedAt, userId)
	if err != nil {
		return "", utils.ErrorHandler(err, "Cannot update password,db error")
	}
	token, err := utils.SignToken(userId, userName, uRole)
	if err != nil {
		return "", utils.ErrorHandler(err, "Cannot create token")

	}
	return token, nil
}

func ForgotPasswordDB(email string) {
	db, err := ConnectDB()
	if err != nil {
		utils.ErrorHandler(err, "Cannot connect to database")
		return
	}
	defer db.Close()

	var exec model.Exec
	err = db.QueryRow("SELECT id FROM execs WHERE email = ?", email).Scan(&exec.ID)
	if err != nil {
		utils.ErrorHandler(err, "Cannot find exec")
		return
	}

	duration, err := strconv.Atoi(os.Getenv("RESET_TOKEN_EXP_DURATION"))
	if err != nil {
		utils.ErrorHandler(err, "Failed to send password reset email ")
	}

	mins := time.Duration(duration)
	expiry := time.Now().Add(mins * time.Minute).Format(time.RFC3339)

	tokenBytes := make([]byte, 32)
	_, err = rand.Read(tokenBytes)
	if err != nil {
		utils.ErrorHandler(err, "Failed to send password reset email ")
	}

	token := hex.EncodeToString(tokenBytes)
	hashedToken := sha256.Sum256(tokenBytes)
	hashTokenString := hex.EncodeToString(hashedToken[:])

	_, err = db.Exec("UPDATE execs set passwordResetToken = ?, tokenExpiresAt = ? where id = ? ", hashTokenString, expiry, exec.ID)
	if err != nil {
		utils.ErrorHandler(err, "Failed to update password")
		return
	}

	//send email
	resetUrl := fmt.Sprintf("http://localhost:8080/execs/resetpassword/reset/%s", token)
	message := fmt.Sprintf("Use this link to reset the password: %s\nReset link valid %s minutes", resetUrl, os.Getenv("RESET_TOKEN_EXP_DURATION"))
	m := mail.NewMessage()
	m.SetHeader("From", "school.admin@example.com")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Password Reset Link")
	m.SetBody("text/html", message)
	d := mail.NewDialer("localhost", 1025, "", "")
	err = d.DialAndSend(m)
	if err != nil {
		utils.ErrorHandler(err, "Failed to send password reset email ")
		return
	}
}

func PasswordResetDB(newPassword string, hash string) error {
	db, err := ConnectDB()
	if err != nil {
		return utils.ErrorHandler(err, "Cannot connect to database")

	}
	defer db.Close()
	var exec model.Exec

	err = db.QueryRow("select id,username,role from execs where passwordResetToken = ?", hash).Scan(&exec.ID, &exec.Username, &exec.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {

			return utils.ErrorHandler(nil, "Cannot find user")
		}

		return utils.ErrorHandler(err, "Cannot find user")
	}

	err, encodedPass := utils.PasswordHashing(newPassword)
	if err != nil {

		return utils.ErrorHandler(err, "Cannot hash password")
	}

	passwordChangedAt := time.Now().Format(time.RFC3339)

	_, err = db.Exec("UPDATE Execs SET password=?, passwordChangedAt = ?,passwordResetToken = NULL,tokenExpiresAt = NULL WHERE id=?", encodedPass, passwordChangedAt, exec.ID)
	if err != nil {

		return utils.ErrorHandler(err, "Cannot update password,db error")

	}
	return nil
}
