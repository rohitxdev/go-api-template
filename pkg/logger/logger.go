package logger

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"sync"
)

const (
	ansiReset           = "\033[0m"
	ansiBold            = "\033[1m"
	ansiDim             = "\033[2m"
	ansiItalic          = "\033[3m"
	ansiUnderline       = "\033[4m"
	ansiBlinkSlow       = "\033[5m"
	ansiBlinkFast       = "\033[6m"
	ansiReverse         = "\033[7m"
	ansiHidden          = "\033[8m"
	ansiStrikethrough   = "\033[9m"
	ansiBlack           = "\033[30m"
	ansiRed             = "\033[31m"
	ansiGreen           = "\033[32m"
	ansiYellow          = "\033[33m"
	ansiBlue            = "\033[34m"
	ansiMagenta         = "\033[35m"
	ansiCyan            = "\033[36m"
	ansiWhite           = "\033[37m"
	ansiBgBlack         = "\033[40m"
	ansiBgRed           = "\033[41m"
	ansiBgGreen         = "\033[42m"
	ansiBgYellow        = "\033[43m"
	ansiBgBlue          = "\033[44m"
	ansiBgMagenta       = "\033[45m"
	ansiBgCyan          = "\033[46m"
	ansiBgWhite         = "\033[47m"
	ansiLowIntensity    = "\033[02m"
	ansiHighIntensity   = "\033[06m"
	ansiBrightBlack     = "\033[90m"
	ansiBrightRed       = "\033[91m"
	ansiBrightGreen     = "\033[92m"
	ansiBrightYellow    = "\033[93m"
	ansiBrightBlue      = "\033[94m"
	ansiBrightMagenta   = "\033[95m"
	ansiBrightCyan      = "\033[96m"
	ansiBrightWhite     = "\033[97m"
	ansiBgBrightBlack   = "\033[100m"
	ansiBgBrightRed     = "\033[101m"
	ansiBgBrightGreen   = "\033[102m"
	ansiBgBrightYellow  = "\033[103m"
	ansiBgBrightBlue    = "\033[104m"
	ansiBgBrightMagenta = "\033[105m"
	ansiBgBrightCyan    = "\033[106m"
	ansiBgBrightWhite   = "\033[107m"
)

const (
	levelDebug = "DBG"
	levelInfo  = "INF"
	levelWarn  = "WRN"
	levelError = "ERR"
)

type HandlerOpts struct {
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
	TimeFormat  string
	Level       slog.Level
	NoColor     bool
}

type handler struct {
	handler slog.Handler
	w       io.Writer
	opts    HandlerOpts
	mu      sync.Mutex
}

var _ slog.Handler = (*handler)(nil)

func NewHandler(w io.Writer, opts HandlerOpts) *handler {
	return &handler{
		handler: slog.NewTextHandler(w, &slog.HandlerOptions{
			Level:       opts.Level,
			ReplaceAttr: opts.ReplaceAttr,
		}),
		w:    w,
		opts: opts,
	}
}

func (h *handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *handler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	var sb strings.Builder
	var level, color string

	sb.Grow(256) //Pre-allocate some space

	switch r.Level {
	case slog.LevelDebug:
		level, color = levelDebug, ansiDim
	case slog.LevelInfo:
		level, color = levelInfo, ansiCyan
	case slog.LevelWarn:
		level, color = levelWarn, ansiYellow
	case slog.LevelError:
		level, color = levelError, ansiRed
	default:
		level = r.Level.String()
	}

	msg := r.Message

	if !h.opts.NoColor && color != "" {
		if level == levelDebug {
			msg = ansiDim + r.Message + ansiReset
		}
		sb.WriteString(ansiDim + r.Time.Format(h.opts.TimeFormat) + ansiReset + " " + color + level + ansiReset + " " + msg)
	} else {
		sb.WriteString(r.Time.Format(h.opts.TimeFormat) + " " + level + " " + msg)
	}

	r.Attrs(func(a slog.Attr) bool {
		if h.opts.ReplaceAttr != nil {
			a = h.opts.ReplaceAttr(nil, a)
		}
		if a.Equal(slog.Attr{}) {
			return true
		}
		if !h.opts.NoColor && color != "" {
			sb.WriteString(" " + ansiDim + a.Key + "=" + ansiReset + "\"" + a.Value.String() + "\"")
		} else {
			sb.WriteString(" " + a.Key + "=" + "\"" + a.Value.String() + "\"")
		}
		return true
	})

	sb.WriteString("\n")

	_, err := h.w.Write([]byte(sb.String()))
	return err
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &handler{
		handler: h.handler.WithAttrs(attrs),
		w:       h.w,
		opts:    h.opts,
	}
}

func (h *handler) WithGroup(name string) slog.Handler {
	return &handler{
		handler: h.handler.WithGroup(name),
		w:       h.w,
		opts:    h.opts,
	}
}
