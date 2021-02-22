package frontend

import "go.uber.org/zap/zapcore"

type ILogger interface {

	// Check returns a CheckedEntry if logging a message at the specified level
	// is enabled. It's a completely optional optimization; in high-performance
	// applications, Check can help avoid allocating a slice to hold fields.
	Check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry

	// Debug logs a message at DebugLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	Debug(msg string, fields ...zapcore.Field)

	// Info logs a message at InfoLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	Info(msg string, fields ...zapcore.Field)

	// Warn logs a message at WarnLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	Warn(msg string, fields ...zapcore.Field)

	// Error logs a message at ErrorLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	Error(msg string, fields ...zapcore.Field)

	// DPanic logs a message at DPanicLevel. The message includes any fields
	// passed at the log site, as well as any fields accumulated on the logger.
	//
	// If the logger is in development mode, it then panics (DPanic means
	// "development panic"). This is useful for catching errors that are
	// recoverable, but shouldn't ever happen.
	DPanic(msg string, fields ...zapcore.Field)

	// Panic logs a message at PanicLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	//
	// The logger then panics, even if logging at PanicLevel is disabled.
	Panic(msg string, fields ...zapcore.Field)

	// Fatal logs a message at FatalLevel. The message includes any fields passed
	// at the log site, as well as any fields accumulated on the logger.
	//
	// The logger then calls os.Exit(1), even if logging at FatalLevel is
	// disabled.
	Fatal(msg string, fields ...zapcore.Field)

	// Sync calls the underlying Core's Sync method, flushing any buffered log
	// entries. Applications should take care to call Sync before exiting.
	Sync() error
}
