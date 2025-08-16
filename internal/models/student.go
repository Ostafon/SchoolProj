package models

type Student struct {
	ID        int    `json:"id" db:"id"`
	FirstName string `json:"firstName" db:"firstName"`
	LastName  string `json:"lastName" db:"lastName"`
	Email     string `json:"email" db:"email"`
	Class     string `json:"class" db:"class"`
}
