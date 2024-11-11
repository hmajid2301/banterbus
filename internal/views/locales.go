package views

import (
	"embed"
	"io"
	"io/fs"

	"gopkg.in/yaml.v3"
)

//go:embed langs

var Locales embed.FS

type Navbar struct {
	Language string `yaml:"language"`
}

type LanguageFile struct {
	Navbar Navbar `yaml:"navbar"`
}

func ListLanguages() (map[string]string, error) {
	languages := map[string]string{}

	entries, err := fs.ReadDir(Locales, "langs")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		file, err := Locales.Open("langs/" + entry.Name())
		if err != nil {
			return nil, err
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}

		var data map[string]LanguageFile
		err = yaml.Unmarshal(content, &data)
		if err != nil {
			return nil, err
		}

		for key, langFile := range data {
			languages[key] = langFile.Navbar.Language
		}
	}

	return languages, nil
}
