package logger

import "context"

// NewContext creates a child context by adding logger.Logger to it
func NewContext(ctx context.Context, log Logger) context.Context {
	return context.WithValue(ctx, loggerKey, log)
}

// FromContext retrieves logger.Logger from the context
func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(loggerKey).(Logger); ok {
		return logger
	}

	return defLogger
}
