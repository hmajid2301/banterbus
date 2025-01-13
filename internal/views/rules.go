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

func RuleMarkdown() (templ.Component, error) {
	_, err := fs.ReadDir(Rules, "rules")
	if err != nil {
		return templ.ComponentScript{}, err
	}

	// TODO: do not hardcode to just a single file (more games equals more rules)
	file, err := Rules.Open("rules/fibbing_it.md")
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
