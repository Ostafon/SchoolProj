package utils

import "errors"

type ContextKey string

func AuthorizeUser(userRole string, allowedRoles ...string) (bool, error) {
	for _, role := range allowedRoles {
		if role == userRole {
			return true, nil
		}
	}
	return false, errors.New("User role not allowed")
}
