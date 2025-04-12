package ui

import (
	"embed"
	_ "embed"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"
)

//go:embed **/*.html
var templateFS embed.FS

func LoadTemplate() (*template.Template, error) {
	funcMap := map[string]any{
		"Field": Field,
	}
	templates := template.New("").Funcs(funcMap)
	err := fs.WalkDir(templateFS, ".", func(path string, d fs.DirEntry, err error) error {

		if !d.IsDir() && strings.HasSuffix(path, ".tmpl.html") {
			templateContent, err := templateFS.ReadFile(path)

			if err != nil {
				return err
			}

			name := filepath.ToSlash(path)

			name = strings.Replace(name, ".tmpl.html", "", 1)

			_, err = templates.New(name).Parse(string(templateContent))
			if err != nil {
				return err
			}
		}

		return nil
	})

	return templates, err
}
