package handlers

import (
	"io"
	"net/http"
)

type requestData struct {
	username  string
	bodyBytes []byte
}

func handleRequestData(r *http.Request) (requestData, error) {
	bodyBytes, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return requestData{}, err
	}
	login := r.Header.Get("X-User-Login")
	return requestData{
		username:  login,
		bodyBytes: bodyBytes,
	}, nil
}
