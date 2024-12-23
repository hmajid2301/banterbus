package logging

import (
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mdobak/go-xerrors"
	slogctx "github.com/veqryn/slog-context"
)

type stackFrame struct {
	Func   string `json:"func"`
	Source string `json:"source"`
	Line   int    `json:"line"`
}

func New(logLevel slog.Level, defaultAttrs []slog.Attr) *slog.Logger {
	var handler slog.Handler

	if os.Getenv("BANTERBUS_ENVIRONMENT") == "local" {
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			AddSource:   true,
			Level:       logLevel,
			TimeFormat:  time.Kitchen,
			ReplaceAttr: replaceAttr,
		})
	} else {
		opts := slog.HandlerOptions{
			AddSource:   true,
			Level:       logLevel,
			ReplaceAttr: replaceAttr,
		}
		handler = slog.NewJSONHandler(os.Stdout, &opts).WithAttrs(defaultAttrs)
	}

	customHandler := slogctx.NewHandler(handler, nil)
	logger := slog.New(customHandler)
	return logger
}

func StripSVGData(message string) string {
	re := regexp.MustCompile(`data:image/svg\+xml;base64,[A-Za-z0-9+/=]+`)
	return re.ReplaceAllString(message, "")
}

// Taken from this blog post https://betterstack.com/community/guides/logging/logging-in-go/#error-logging-with-slog
func replaceAttr(_ []string, a slog.Attr) slog.Attr {
	//nolint:gocritic
	switch a.Value.Kind() {
	case slog.KindAny:
		//nolint:gocritic
		switch v := a.Value.Any().(type) {
		case error:
			a.Value = fmtErr(v)
		}
	}

	return a
}

func marshalStack(err error) []stackFrame {
	trace := xerrors.StackTrace(err)

	if len(trace) == 0 {
		return nil
	}

	frames := trace.Frames()

	s := make([]stackFrame, len(frames))

	for i, v := range frames {
		f := stackFrame{
			Source: filepath.Join(
				filepath.Base(filepath.Dir(v.File)),
				filepath.Base(v.File),
			),
			Func: filepath.Base(v.Function),
			Line: v.Line,
		}

		s[i] = f
	}

	return s
}

func fmtErr(err error) slog.Value {
	var groupValues []slog.Attr

	groupValues = append(groupValues, slog.String("msg", err.Error()))

	frames := marshalStack(err)

	if frames != nil {
		groupValues = append(groupValues,
			slog.Any("trace", frames),
		)
	}

	return slog.GroupValue(groupValues...)
}
