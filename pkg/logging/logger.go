package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

type Logger interface {
	Close()
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, err error, args ...any)
	Errorf(format string, err error, attrs ...any)
	Fatal(msg string, err error, args ...any)
	Fatalf(format string, args ...any)
	Panic(msg string, err error, args ...any)
	Panicf(format string, args ...any)
	With(args ...any) Logger
}

type Envelop map[string]any

type AppLogger struct {
	logger *slog.Logger
	file   *os.File
}

func NewAppLogger(logFilePath string, level slog.Leveler) (Logger, error) {
	var handlers []slog.Handler
	var file *os.File

	// Console handler
	consoleHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	handlers = append(handlers, consoleHandler)

	// File handler
	if logFilePath != "" {
		var err error
		file, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		fileHandler := slog.NewJSONHandler(file, &slog.HandlerOptions{
			Level: level,
		})
		handlers = append(handlers, fileHandler)
	}

	// Use your custom MultiHandler
	multiHandler := NewMultiHandler(handlers...)
	logger := slog.New(multiHandler)

	return &AppLogger{
		logger: logger,
		file:   file,
	}, nil
}

// Close gracefully closes any open resources like the log file.
func (l *AppLogger) Close() {
	if l.file != nil {
		if err := l.file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing log file: %v\n", err)
		}
	}
}

func (l *AppLogger) Info(msg string, args ...any) {
	if l == nil || l.logger == nil {
		return
	}
	attrs := processAttrs(args...)
	l.logger.LogAttrs(context.Background(), slog.LevelInfo, msg, attrs...)
}

func (l *AppLogger) Debug(msg string, args ...any) {
	if l == nil || l.logger == nil {
		return
	}
	l.logger.LogAttrs(context.Background(), slog.LevelDebug, msg, processAttrs(args...)...)
}

func (l *AppLogger) Warn(msg string, args ...any) {
	if l == nil || l.logger == nil {
		return
	}
	l.logger.LogAttrs(context.Background(), slog.LevelWarn, msg, processAttrs(args...)...)
}

func (l *AppLogger) Error(msg string, err error, args ...any) {
	if l == nil || l.logger == nil {
		return
	}
	attrs := append([]slog.Attr{slog.Any("error", err)}, processAttrs(args...)...)
	l.logger.LogAttrs(context.Background(), slog.LevelError, msg, attrs...)
}

func (l *AppLogger) Errorf(format string, err error, args ...any) {
	if l == nil || l.logger == nil {
		return
	}
	msg := fmt.Sprintf(format, err)
	attrs := append([]slog.Attr{slog.Any("error", err)}, processAttrs(args...)...)
	l.logger.LogAttrs(context.Background(), slog.LevelError, msg, attrs...)
}

func (l *AppLogger) Fatal(msg string, err error, args ...any) {
	if l == nil || l.logger == nil {
		fmt.Fprintf(os.Stderr, "FATAL: %s: %v\n", msg, err)
		os.Exit(1)
	}

	attrs := append([]slog.Attr{slog.Any("error", err)}, processAttrs(args...)...)
	l.logger.LogAttrs(context.Background(), slog.LevelError, msg, attrs...)
	l.Close()
	os.Exit(1)
}

func (l *AppLogger) Fatalf(format string, args ...any) {
	if l == nil || l.logger == nil {
		fmt.Fprintf(os.Stderr, "FATAL: "+format+"\n", args...)
		os.Exit(1)
	}

	msg := fmt.Sprintf(format, args...)
	l.logger.LogAttrs(context.Background(), slog.LevelError, msg)
	l.Close()
	os.Exit(1)
}

func (l *AppLogger) Panic(msg string, err error, args ...any) {
	if l == nil || l.logger == nil {
		fmt.Fprintf(os.Stderr, "PANIC: %s: %v\n", msg, err)
		panic(msg)
	}

	attrs := append([]slog.Attr{slog.Any("error", err)}, processAttrs(args...)...)
	l.logger.LogAttrs(context.Background(), slog.LevelError, msg, attrs...)
	panic(msg)
}

func (l *AppLogger) Panicf(format string, args ...any) {
	if l == nil || l.logger == nil {
		fmt.Fprintf(os.Stderr, "PANIC: "+format+"\n", args...)
		panic(fmt.Sprintf(format, args...))
	}

	msg := fmt.Sprintf(format, args...)
	l.logger.LogAttrs(context.Background(), slog.LevelError, msg)
	panic(msg)
}

func (l *AppLogger) With(args ...any) Logger {
	if l == nil || l.logger == nil {
		return l
	}

	attrs := processAttrs(args...)

	// Convert []slog.Attr to []any for slog.With(...)
	anyAttrs := make([]any, len(attrs))
	for i, attr := range attrs {
		anyAttrs[i] = attr
	}

	newLogger := l.logger.With(anyAttrs...)
	return &AppLogger{
		logger: newLogger,
		file:   l.file, // share the file handle
	}
}

//------------------------------------------------------------
// func (l *AppLogger) Info(msg string, args ...any) {
// 	if l == nil || l.logger == nil {
// 		return
// 	}

// 	finalAttrs := processAttrs(args...)

// 	l.logger.Info(msg, finalAttrs...)
// }

// func (l *AppLogger) Debug(msg string, args ...any) {
// 	if l == nil || l.logger == nil {
// 		return
// 	}
// 	// l.logger.Debug(msg, processAttrs(args...)...)
// }

// func (l *AppLogger) Warn(msg string, args ...any) {
// 	if l == nil || l.logger == nil {
// 		return
// 	}
// 	// l.logger.Warn(msg, processAttrs(args...)...)
// }

// func (l *AppLogger) Error(msg string, err error, args ...any) {
// 	if l == nil || l.logger == nil {
// 		return
// 	}
// 	allAttrs := appendErrorAttr(err, nil)
// 	// allAttrs = append(allAttrs, processAttrs(args...)...)

// 	l.logger.Error(msg, allAttrs...)
// }

// func (l *AppLogger) Errorf(format string, err error, attrs ...any) {
// 	if l == nil || l.logger == nil {
// 		return
// 	}

// 	msg := fmt.Sprintf("%s: %+v", format, err)

// 	allAttrs := appendErrorAttr(err, nil)
// 	// allAttrs = append(allAttrs, processAttrs(attrs...)...)

// 	l.logger.Error(msg, allAttrs...)
// }

// func (l *AppLogger) Fatal(msg string, err error, attrs ...any) {
// 	if l == nil || l.logger == nil {
// 		// Fallback to standard library log if logger is not initialized
// 		fmt.Fprintf(os.Stderr, "FATAL: %s: %v\n", msg, err)
// 		os.Exit(1)
// 	}

// 	allAttrs := appendErrorAttr(err, nil)
// 	// allAttrs = append(allAttrs, processAttrs(attrs...)...)

// 	l.logger.Error(msg, allAttrs...)
// 	l.Close()
// 	os.Exit(1)
// }

// func (l *AppLogger) Fatalf(format string, args ...any) {
// 	if l == nil || l.logger == nil {
// 		fmt.Fprintf(os.Stderr, "FATAL: "+format+"\n", args...)
// 		os.Exit(1)
// 	}

// 	msg := fmt.Sprintf(format, args...)
// 	l.logger.Error(msg)
// 	l.Close()
// 	os.Exit(1)
// }

// // Panic logs a message at the Error level, includes an explicit error object,
// // and then panics.
// func (l *AppLogger) Panic(msg string, err error, attrs ...any) {
// 	if l == nil || l.logger == nil {
// 		// Fallback to standard library log if logger is not initialized
// 		fmt.Fprintf(os.Stderr, "PANIC: %s: %v\n", msg, err)
// 		panic(msg)
// 	}

// 	// Start with the error attribute
// 	allAttrs := appendErrorAttr(err, nil)

// 	// Process and append any additional attributes
// 	// allAttrs = append(allAttrs, processAttrs(attrs...)...)

// 	l.logger.Error(msg, allAttrs...) // Log before panicking
// 	panic(msg)
// }

// // Panicf logs a formatted message at the Error level, and then panics.
// func (l *AppLogger) Panicf(format string, args ...any) {
// 	if l == nil || l.logger == nil {
// 		fmt.Fprintf(os.Stderr, "PANIC: "+format+"\n", args...)
// 		panic(fmt.Sprintf(format, args...))
// 	}

// 	msg := fmt.Sprintf(format, args...)
// 	l.logger.Error(msg)
// 	panic(msg)
// }

// func processAttrs(args ...any) []slog.Attr {
// 	var attrs []slog.Attr

// 	i := 0
// 	for i < len(args) {
// 		switch v := args[i].(type) {
// 		case map[string]any:
// 			for k, val := range v {
// 				attrs = append(attrs, slog.Any(k, val))
// 			}
// 			i++ // map consumes one argument
// 		case string:
// 			if i+1 < len(args) {
// 				attrs = append(attrs, slog.Any(v, args[i+1]))
// 				i += 2 // string + value
// 			} else {
// 				// unmatched key with no value
// 				attrs = append(attrs, slog.String("malformed_arg", v))
// 				i++
// 			}
// 		default:
// 			// unknown or malformed key/value
// 			attrs = append(attrs, slog.Any(fmt.Sprintf("arg_%d", i), v))
// 			i++
// 		}
// 	}

// 	return attrs
// }

////////////////////////////////////////////////////////////

// Info logs a message at the Info level, processing flexible arguments (e.g., string, map, error, etc.).
// func (l *AppLogger) Info(msg string, args ...any) {
// 	// Check if the logger is valid (this is a basic check to avoid nil pointer dereference)
// 	if l == nil || l.logger == nil {
// 		return
// 	}

// 	// Initialize slices to hold "context" and "metadata"
// 	var contextAttrs []any
// 	var metaAttrs []any

// 	// Loop through all the args passed to process them
// 	for _, arg := range args {
// 		switch v := arg.(type) {
// 		case common.Envelop: // If the argument is of type Envelop (map[string]any)
// 			// Group Envelop into "metadata"
// 			metaAttrs = append(metaAttrs, wrapInGroup("metadata", convertEnvelopToSlogArgs(v)...))
// 		case map[string]any: // If the argument is a regular map
// 			// Group the map into "metadata"
// 			metaAttrs = append(metaAttrs, wrapInGroup("metadata", convertEnvelopToSlogArgs(v)...))
// 		case context.Context: // If the argument is of type context (useful for distributed tracing)
// 			// Enrich the log context with trace information (traceID, requestID, etc.)
// 			contextAttrs = append(contextAttrs, wrapInGroup("context", enrichLogContext(v)...))
// 		default:
// 			// For other types of args (non-map, non-context), treat them as generic context
// 			contextAttrs = append(contextAttrs, slog.Any("value", v))
// 		}
// 	}

// Combine context and metadata into the final log attributes
// 	var finalAttrs []any
// 	if len(contextAttrs) > 0 {
// 		// Wrap context attributes if we have them
// 		finalAttrs = append(finalAttrs, wrapInGroup("context", contextAttrs...))
// 	}
// 	if len(metaAttrs) > 0 {
// 		// Wrap metadata attributes if we have them
// 		finalAttrs = append(finalAttrs, wrapInGroup("metadata", metaAttrs...))
// 	}

// 	// Finally, log the message with all the collected attributes
// 	l.logger.Info(msg, finalAttrs...)
// }

// func convertEnvelopToSlogAttrs(envelop map[string]any) []slog.Attr { // Renamed func & changed return type
// 	if envelop == nil {
// 		return nil
// 	}
// 	attrs := make([]slog.Attr, 0, len(envelop)) // Now holds slog.Attr
// 	for k, v := range envelop {
// 		attrs = append(attrs, slog.Any(k, v)) // Use slog.Any
// 	}
// 	return attrs
// }

// func wrapInGroup(key string, attrs ...any) slog.Attr {
// 	// Only group attributes if there are any. Otherwise, avoid unnecessary grouping.
// 	if len(attrs) == 0 {
// 		return slog.Any(key, nil)
// 	}
// 	return slog.Group(key, attrs...)
// }

// Converts an Envelop (map[string]any) to a slice of slog.Attr
func convertEnvelopToSlogArgs(envelop map[string]any) []any {
	var attrs []any
	for k, v := range envelop {
		attrs = append(attrs, slog.Any(k, v)) // Convert map keys and values into slog.Attr
	}
	return attrs
}

// Helper: safely attach an error to slog args.
func appendErrorAttr(err error, initialAttrs []any) []any {
	if err != nil {

		if initialAttrs == nil {
			initialAttrs = make([]any, 0, 1)
		}
		return append(initialAttrs, slog.Any("error", err))
	}
	return initialAttrs
}

// func processAttrs(args ...any) []any {
// 	var finalAttrs []any

// 	for _, arg := range args {
// 		switch v := arg.(type) {
// 		case map[string]any:
// 			// Handle Envelop (map[string]any) as individual key-value pairs
// 			finalAttrs = append(finalAttrs, convertEnvelopToSlogAttrs(v)...)
// 		default:
// 			// For non-map values (strings, ints, etc.), directly add them
// 			finalAttrs = append(finalAttrs, slog.Any(fmt.Sprintf("%v", v), v))
// 		}
// 	}

// 	return finalAttrs
// }

// Converts an Envelop (map[string]any) to a slice of any, for logging.
// func convertEnvelopToSlogAttrs(envelop map[string]any) []any {
// 	var attrs []any
// 	for k, v := range envelop {
// 		attrs = append(attrs, slog.Any(k, v)) // Converts each map key-value pair to slog.Attr
// 	}
// 	return attrs
// }

// processAttrs checks if the first element is an Envelop (metadata map), and adds it as a single group
// func processAttrs(args ...any) []any {
// 	if len(args) == 0 {
// 		return nil
// 	}

// 	// Check if the first argument is an Envelop (map[string]any)
// 	if env, ok := args[0].(map[string]any); ok {
// 		// Only wrap the first Envelop in the "metadata" group and process the remaining args
// 		return append([]any{slog.Group("metadata", convertEnvelopToSlogArgs(env)...)}, args[1:]...)
// 	}

// 	// If no Envelop, just pass the attributes as they are (non-map)
// 	return args
// }

// // processAttrs: Process different kinds of attributes
// // func processAttrs(args ...any) []any {
// // 	var flatArgs []any

// // 	for _, arg := range args {
// // 		switch v := arg.(type) {
// // 		case map[string]any:
// // 			flatArgs = append(flatArgs, convertEnvelopToSlogArgs(v)...)
// // 		case slog.Attr:
// // 			flatArgs = append(flatArgs, v)
// // 		case string, int, float64, bool, fmt.Stringer, error:
// // 			flatArgs = append(flatArgs, slog.Any("value", v))
// // 		default:
// // 			flatArgs = append(flatArgs, slog.Any("value", v))
// // 		}
// // 	}

// // 	return flatArgs
// // }

// // Group logs by adding a context or metadata wrapper
// // wrapInGroup groups attributes into named categories, like "metadata", "context", etc.
// func wrapInGroup(key string, attrs ...any) slog.Attr {
// 	// If no attributes, just return an empty group
// 	if len(attrs) == 0 {
// 		return slog.Any(key, nil)
// 	}
// 	return slog.Group(key, attrs...)
// }

// // Converts an Envelop (map[string]any) to a slice of slog.Attr
// func convertEnvelopToSlogArgs(envelop map[string]any) []any {
// 	var attrs []any
// 	for k, v := range envelop {
// 		attrs = append(attrs, slog.Any(k, v)) // Converts each map key-value pair to slog.Attr
// 	}
// 	return attrs
// }

// Helper: safely attach an error to slog args.
// func appendErrorAttr(err error, initialAttrs []any) []any {
// 	if err != nil {
// 		// If initialAttrs is nil, make a new slice with capacity for the error.
// 		if initialAttrs == nil {
// 			initialAttrs = make([]any, 0, 1)
// 		}
// 		return append(initialAttrs, slog.Any("error", err))
// 	}
// 	return initialAttrs // If err is nil, just return the initial attributes without adding an "error" key.
// }

// Extract context from the arguments if it's available
func getContextFromArgs(args []any) (context.Context, bool) {
	if len(args) > 0 {
		if ctx, ok := args[0].(context.Context); ok {
			return ctx, true
		}
	}
	return nil, false
}

// Enrich the log context with trace info like traceID, requestID, sessionID, etc.
func enrichLogContext(ctx context.Context) []any {
	// Retrieve traceID, requestID, sessionID, etc., from the context (you can add more as needed)
	traceID, _ := ctx.Value("traceID").(string)
	requestID, _ := ctx.Value("requestID").(string)
	sessionID, _ := ctx.Value("sessionID").(string)

	// Return these as a slice of key-value pairs (this will be wrapped into the "context" group)
	return []any{
		slog.Any("traceID", traceID),
		slog.Any("requestID", requestID),
		slog.Any("sessionID", sessionID),
	}
}

// TYPED WRAPPERS
// Typed Wrapper for String
func String(key string, value string) slog.Attr {
	return slog.String(key, value)
}

// Typed Wrapper for Int
func Int(key string, value int) slog.Attr {
	return slog.Int(key, value)
}

// Typed Wrapper for Bool
func Bool(key string, value bool) slog.Attr {
	return slog.Bool(key, value)
}

// Typed Wrapper for Error
func Error(key string, value error) slog.Attr {
	return slog.Any(key, value)
}

// Typed Wrapper for Float64
func Float64(key string, value float64) slog.Attr {
	return slog.Float64(key, value)
}

// func NewLogger(path string) (*AppLogger, error) {
// 	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 	if err != nil {
// 		return nil, err
// 	}

// 	consoleHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})

// 	jsonHandler := slog.NewJSONHandler(file, &slog.HandlerOptions{Level: slog.LevelInfo})

// 	multi := slog.New(NewMultiHandler(jsonHandler, consoleHandler))

// 	return &AppLogger{logger: multi, file: file}, nil
// }

// type Logger interface {
// 	Info(msg string, args ...any)
// 	Error(msg string, err error, args ...any)
// 	Errorf(format string, err error, args ...any)
// 	Fatal(msg string, err error, args ...any)
// 	Fatalf(format string, err error, args ...any)
// 	Panic(msg string, err error, args ...any)
// 	Panicf(format string, err error, args ...any)
// }

// type AppLogger struct {
// 	logger *slog.Logger
// 	file   *os.File
// }

// func NewLogger(logFilePath string) (*AppLogger, error) {
// 	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 	if err != nil {
// 		return nil, err
// 	}

// 	jsonHandler := slog.NewJSONHandler(file, &slog.HandlerOptions{
// 		Level: slog.LevelInfo,
// 	})

// 	consoleHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
// 		Level: slog.LevelInfo,
// 	})

// 	multiHandler := NewMultiHandler(jsonHandler, consoleHandler)

// 	return &AppLogger{
// 		logger: slog.New(multiHandler),
// 		file:   file,
// 	}, nil

// }

// func (l *AppLogger) Close() {
// 	l.file.Close()
// }

// // Info logs an informational message.
// func (l *AppLogger) Info(msg string, args ...any) {
// 	l.logger.Info(msg, args...)
// }

// // Error logs an error message with a structured error field.
// func (l *AppLogger) Error(msg string, err error, args ...any) {
// 	l.logger.Error(msg, append(args, slog.Any("error", err))...)
// }

// // Errorf logs a formatted error message.
// func (l *AppLogger) Errorf(format string, args ...any) {
// 	l.logger.Error(fmt.Sprintf(format, args...))
// }

// // Fatal logs an error and exits the program.
// func (l *AppLogger) Fatal(msg string, err error, args ...any) {
// 	l.logger.Error(msg, append(args, slog.Any("error", err))...)
// 	l.Close()
// 	os.Exit(1)
// }

// // Fatalf logs a formatted error and exits the program.
// func (l *AppLogger) Fatalf(format string, args ...any) {
// 	l.logger.Error(fmt.Sprintf(format, args...))
// 	l.Close()
// 	os.Exit(1)
// }

// // Panic logs an error and panics.
// func (l *AppLogger) Panic(msg string, args ...any) {
// 	l.logger.Error(msg, args...)
// 	panic(msg)
// }

// // Panicf logs a formatted error and panics.
// func (l *AppLogger) Panicf(format string, args ...any) {
// 	msg := fmt.Sprintf(format, args...)
// 	l.logger.Error(msg)
// 	panic(msg)
// }
