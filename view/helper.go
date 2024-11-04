package view

import (
	"errors"
	"fmt"
	"html/template"
	"reflect"
	"strings"
	"sync"
)

const renderFieldTemplateString = `<data class="{{ .Name }}"> {{ .Value }} </data>`

var getFieldTmpl = sync.OnceValue(func() *template.Template {
	tmpl := template.New("data")
	tmpl, err := tmpl.Parse(renderFieldTemplateString)

	if err != nil {
		return nil
	}

	return tmpl
})

func Field(name string, entity any) (template.HTML, error) {
	v := reflect.ValueOf(entity)
	field := v.FieldByName(name)

	if !field.IsValid() {
		return "", fmt.Errorf("field of this name is not valid: %s", name)
	}

	tmpl := getFieldTmpl()

	if tmpl == nil {
		return "", errors.New("Could not parse template")
	}

	data := struct {
		Name  string
		Value interface{}
	}{
		Name:  strings.ToLower(name),
		Value: field.Interface(),
	}

	var renderedField strings.Builder
	err := tmpl.ExecuteTemplate(&renderedField, "data", data)
	if err != nil {
		return "", err
	}

	return template.HTML(renderedField.String()), nil
}
