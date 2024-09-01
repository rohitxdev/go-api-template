package prettylog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
)

const (
	ansiReset           = "\x1b[0m"
	ansiBold            = "\x1b[1m"
	ansiItalic          = "\x1b[3m"
	ansiUnderline       = "\x1b[4m"
	ansiBlinkSlow       = "\x1b[5m"
	ansiBlinkFast       = "\x1b[6m"
	ansiReverse         = "\x1b[7m"
	ansiHidden          = "\x1b[8m"
	ansiStrikethrough   = "\x1b[9m"
	ansiBlack           = "\x1b[30m"
	ansiRed             = "\x1b[31m"
	ansiGreen           = "\x1b[32m"
	ansiYellow          = "\x1b[33m"
	ansiBlue            = "\x1b[34m"
	ansiMagenta         = "\x1b[35m"
	ansiCyan            = "\x1b[36m"
	ansiWhite           = "\x1b[37m"
	ansiBgBlack         = "\x1b[40m"
	ansiBgRed           = "\x1b[41m"
	ansiBgGreen         = "\x1b[42m"
	ansiBgYellow        = "\x1b[43m"
	ansiBgBlue          = "\x1b[44m"
	ansiBgMagenta       = "\x1b[45m"
	ansiBgCyan          = "\x1b[46m"
	ansiBgWhite         = "\x1b[47m"
	ansiLowIntensity    = "\x1b[02m"
	ansiHighIntensity   = "\x1b[06m"
	ansiBrightBlack     = "\x1b[90m"
	ansiBrightRed       = "\x1b[91m"
	ansiBrightGreen     = "\x1b[92m"
	ansiBrightYellow    = "\x1b[93m"
	ansiBrightBlue      = "\x1b[94m"
	ansiBrightMagenta   = "\x1b[95m"
	ansiBrightCyan      = "\x1b[96m"
	ansiBrightWhite     = "\x1b[97m"
	ansiBgBrightBlack   = "\x1b[100m"
	ansiBgBrightRed     = "\x1b[101m"
	ansiBgBrightGreen   = "\x1b[102m"
	ansiBgBrightYellow  = "\x1b[103m"
	ansiBgBrightBlue    = "\x1b[104m"
	ansiBgBrightMagenta = "\x1b[105m"
	ansiBgBrightCyan    = "\x1b[106m"
	ansiBgBrightWhite   = "\x1b[107m"
	ansiGrey            = "\x1b[38;5;240m"
)

func colorize(ansiCode string, text string) string {
	return ansiCode + text + ansiReset
}

func getLevelColor(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return ansiGrey
	case slog.LevelInfo:
		return ansiBrightCyan
	case slog.LevelWarn:
		return ansiBrightYellow
	case slog.LevelError:
		return ansiBrightRed
	default:
		return ansiBrightWhite
	}
}

type logHandler struct {
	handler slog.Handler
	buf     *bytes.Buffer
	mu      *sync.Mutex
	w       io.Writer
}

var _ slog.Handler = (*logHandler)(nil)

// Prints pretty logs in both plain text and JSON formats depending on the number of attributes. To be used only for development.
func NewHandler(w io.Writer, opts *slog.HandlerOptions) *logHandler {
	buf := new(bytes.Buffer)
	return &logHandler{
		//pass buffer instead of given writer to inner handler
		handler: slog.NewJSONHandler(buf, opts),
		buf:     buf,
		mu:      &sync.Mutex{},
		w:       w,
	}
}

func (h *logHandler) computeAttrs(ctx context.Context, r slog.Record) (map[string]any, error) {
	h.mu.Lock()
	defer func() {
		h.buf.Reset()
		h.mu.Unlock()
	}()

	if err := h.handler.Handle(ctx, r); err != nil {
		return nil, errors.Join(errors.New("inner handler handle"), err)
	}

	var attrs = make(map[string]any)

	if err := json.Unmarshal(h.buf.Bytes(), &attrs); err != nil {
		return nil, errors.Join(errors.New("handler unmarshal"), err)
	}
	return attrs, nil
}

func (h *logHandler) Handle(ctx context.Context, r slog.Record) error {
	attrs, err := h.computeAttrs(ctx, r)
	if err != nil {
		return err
	}

	timeStr := colorize(ansiGrey, r.Time.Format("[15:04:05]"))
	levelStr := colorize(getLevelColor(r.Level), r.Level.String())
	msgStr := colorize(ansiWhite, r.Message)

	logStr := fmt.Sprintf("%s %s %s", timeStr, levelStr, msgStr)

	//Delete level, msg and time from attributes as they are already printed
	delete(attrs, "level")
	delete(attrs, "msg")
	delete(attrs, "time")

	//If still any attributes, print them
	if len(attrs) > 0 {
		bytes, err := json.MarshalIndent(attrs, "", "  ")
		if err != nil {
			return errors.Join(errors.New("handler marshal"), err)
		}
		logStr += fmt.Sprintf(" %s\n", colorize(ansiGrey, string(bytes)))
	}

	fmt.Fprintln(h.w, logStr)

	return nil
}

func (h *logHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *logHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &logHandler{
		handler: h.handler.WithAttrs(attrs),
		buf:     h.buf,
		mu:      h.mu,
	}
}

func (h *logHandler) WithGroup(name string) slog.Handler {
	return &logHandler{
		handler: h.handler.WithGroup(name),
		buf:     h.buf,
		mu:      h.mu,
	}
}
