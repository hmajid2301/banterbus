package views

import (
	"bytes"
	"context"
	"embed"
	"io"
	"io/fs"

	"github.com/a-h/templ"
	"github.com/yuin/goldmark"
)

//go:embed rules

var Rules embed.FS

func RuleMarkdown(gameName string) (templ.Component, error) {
	_, err := fs.ReadDir(Rules, "rules")
	if err != nil {
		return templ.ComponentScript{}, err
	}

	// Support multiple games by using the game name to determine which rules file to load
	filename := "rules/" + gameName + ".md"
	file, err := Rules.Open(filename)
	if err != nil {
		return templ.ComponentScript{}, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return templ.ComponentScript{}, err
	}

	var buf bytes.Buffer
	if err := goldmark.Convert(content, &buf); err != nil {
		return templ.ComponentScript{}, err
	}

	markdownHTML := unsafe(buf.String())
	return markdownHTML, nil
}

func unsafe(html string) templ.Component {
	return templ.ComponentFunc(func(_ context.Context, w io.Writer) (err error) {
		_, err = io.WriteString(w, html)
		return
	})
}
