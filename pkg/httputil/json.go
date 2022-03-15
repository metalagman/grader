package httputil

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// ReadBody into json struct
func ReadBody(r *http.Request, v interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	_ = r.Body.Close()
	if err != nil {
		return fmt.Errorf("body read: %w", err)
	}

	err = json.Unmarshal(body, v)
	if err != nil {
		return fmt.Errorf("json decode: %w", err)
	}

	return nil
}

type JsonError struct {
	Message string `json:"message"`
}

// WriteError formatted in json
func WriteError(w http.ResponseWriter, err error, statusCode int) {
	WriteResponse(w, &JsonError{Message: err.Error()}, statusCode)
}

// WriteResponse formatted in json
func WriteResponse(w http.ResponseWriter, v interface{}, statusCode int) {
	resBody, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(statusCode)
	_, _ = w.Write(resBody)
}
