package domain

import (
	"errors"
	"strings"
)

func validateNonEmpty(field, mapFieldName, errorMessage string, errorMap map[string]string) {
	if len(field) == 0 {
		errorMap[mapFieldName] = errorMessage
	}
}

func stackValidationErrors(validationErrors map[string]string) error {
	if len(validationErrors) == 0 {
		return nil
	}

	validErrMessages := make([]string, 0)
	for _, value := range validationErrors {
		validErrMessages = append(validErrMessages, value)
	}

	return errors.New(strings.Join(validErrMessages, "\n"))
}
