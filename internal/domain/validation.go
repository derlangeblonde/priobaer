package domain

import "fmt"

type ValidationErrs map[string]string

func EmptyValidationErrs() ValidationErrs {
	return make(ValidationErrs, 0)
}

func (e ValidationErrs) Empty() bool {
	return len(e) == 0
}

func (e ValidationErrs) Error() (message string) {
	for fieldName, error := range e {
		message += fmt.Sprintf("%s: %s\n", fieldName, error)
	}

	return message
}

func GreaterEqualZero(value int, fieldName string, errs ValidationErrs) {
	if value < 0 {
		errs[fieldName] = fmt.Sprintf("Muss größer gleich 0 sein, ist aber: %d", value)
	}
}

func NotEmpty(value, fieldName string, errs ValidationErrs){
	if len(value) == 0 {
		errs[fieldName] = "Darf nicht leer sein"
	}
}
