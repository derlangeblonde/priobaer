package model

func validateNonEmpty(field, mapFieldName, errorMessage string, errorMap map[string]string) {
	if len(field) == 0 {
		errorMap[mapFieldName] = errorMessage
	}
}
