package logging

import (
	"fmt"
	"log/slog"
)

// Process attributes, handling both maps and non-maps.
func processAttrs(args ...any) []slog.Attr {
	var attrs []slog.Attr
	i := 0
	for i < len(args) {
		switch v := args[i].(type) {
		case map[string]any:
			for k, val := range v {
				attrs = append(attrs, slog.Any(k, val))
			}
			i++
		case string:
			// Check for key-value pair
			if i+1 < len(args) {
				attrs = append(attrs, slog.Any(v, args[i+1]))
				i += 2
			} else {
				// Unmatched string (likely a mistake)
				attrs = append(attrs, slog.Any("malformed_key", v))
				i++
			}
		default:
			// Skip unknown positional values (optional: log them under debug)
			attrs = append(attrs, slog.Any(fmt.Sprintf("arg_%d", i), v))
			i++
		}
	}
	return attrs
}

// wrapInGroup: Safely wrap attributes into a group. If no attributes, return an empty group.
// func wrapInGroup(key string, attrs ...any) slog.Attr {
// 	if len(attrs) == 0 {
// 		return slog.Any(key, nil)
// 	}
// 	return slog.Group(key, attrs...)
// }

// Flatten map[string]any to a slice of any (this is the core fix).
// func convertEnvelopToSlogAttrs(envelop map[string]any) []slog.Attr {
// 	var attrs []slog.Attr
// 	for k, v := range envelop {
// 		attrs = append(attrs, slog.Any(k, v))
// 	}
// 	return attrs
// }
