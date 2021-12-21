package httputil

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"net/http"
)

type ValidationErrorResponse struct {
	Errors ValidationErrors `json:"errors"`
}

type ValidationErrors []ValidationError

type ValidationError struct {
	Msg   string `json:"msg"`
	Param string `json:"param"`
	Value string `json:"value"`
}

// ValidateData and send errors, returns true if no validation errors
func ValidateData(w http.ResponseWriter, v interface{}) bool {
	validate := validator.New()
	err := validate.Struct(v)
	if err != nil {
		errors := make(ValidationErrors, 0)
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, ValidationError{
				Msg:   err.Error(),
				Param: err.Field(),
				Value: fmt.Sprintf("%s", err.Value()),
			})
		}
		WriteValidationErrors(w, errors)
		return false
	}

	return true
}

// WriteValidationErrors formatted in json
func WriteValidationErrors(w http.ResponseWriter, errors ValidationErrors) {
	WriteResponse(w, ValidationErrorResponse{errors}, http.StatusUnprocessableEntity)
}
