package handlers

import (
	"fmt"
	"net/http"
)

func handleGetUser(r *http.Request) (string, error) {
	login := r.Header.Get("X-User-Login")
	if login == "" {
		return "", fmt.Errorf("no user login")
	}
	return login, nil
}
