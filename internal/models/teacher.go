package models

type Teacher struct {
	ID        int    `json:"id" db:"id"`
	FirstName string `json:"firstName" db:"firstName"`
	LastName  string `json:"lastName" db:"lastName"`
	Email     string `json:"email" db:"email"`
	Class     string `json:"class" db:"class"`
	Subject   string `json:"subject" db:"subject"`
}
