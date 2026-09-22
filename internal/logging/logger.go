package logging

import (
	"io"
	"log/slog"
)

// New builds the logger the resolved Options describe. Records go to
// opts.Writer, or to fallback when opts.Writer is nil.
func New(opts Options, fallback io.Writer) (*slog.Logger, error) {
	resolved, err := opts.Resolve()
	if err != nil {
		return nil, err
	}

	w := opts.Writer
	if w == nil {
		w = fallback
	}
	if w == nil {
		w = io.Discard
	}

	handlerOpts := &slog.HandlerOptions{Level: resolved.Level}
	var handler slog.Handler
	if resolved.Format == FormatJSON {
		handler = slog.NewJSONHandler(w, handlerOpts)
	} else {
		handler = slog.NewTextHandler(w, handlerOpts)
	}
	return slog.New(handler), nil
}
