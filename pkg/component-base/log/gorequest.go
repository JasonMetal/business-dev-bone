package log

import (
	"context"
	"fmt"
)

type gorequestLogger struct {
	prefix string
	logger *zapLogger
}

func NewGoRequestLogger(logger *zapLogger) *gorequestLogger {
	return &gorequestLogger{logger: logger}
}

func (l *gorequestLogger) Println(ctx context.Context, v ...any) {
	msg := fmt.Sprint(v...)
	if l.prefix != "" {
		msg = fmt.Sprintf("%s %s", l.prefix, msg)
	}
	l.logger.L(ctx).Info(msg)
}

func (l *gorequestLogger) Printf(ctx context.Context, format string, v ...any) {
	if l.prefix != "" {
		l.logger.L(ctx).Infof(fmt.Sprintf("%s %s", l.prefix, format), v...)
	} else {
		l.logger.L(ctx).Infof(format, v...)
	}
}

func (l *gorequestLogger) SetPrefix(prefix string) {
	l.prefix = prefix
}
