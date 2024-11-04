package view

import (
	"embed"
	_ "embed"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"
)

//go:embed **/*.html
var templateFS embed.FS

func LoadTemplate() (*template.Template, error) {
	templates := template.New("")
	err := fs.WalkDir(templateFS, ".", func(path string, d fs.DirEntry, err error) error {

		if !d.IsDir() && strings.HasSuffix(path, ".tmpl.html") {
			templateContent, err := templateFS.ReadFile(path)

			if err != nil {
				return err
			}

			name := filepath.ToSlash(path)

			name = strings.Replace(name, ".tmpl.html", "", 1)

			fmt.Printf("- Registered new template: %s\n", name)

			_, err = templates.New(name).Parse(string(templateContent))
			if err != nil {
				return err
			}
		}

		return nil
	})

	return templates, err
}
