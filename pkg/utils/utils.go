package utils

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"multipass/pkg/common"
)

func GetParamID(r *http.Request) (int, error) {
	idStr := path.Base(r.URL.Path)
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		return 0, err
	}
	return id, nil
}

func GetQueryParam(r *http.Request, key string) (string, error) {
	val := r.URL.Query().Get(key)
	if val == "" {
		return "", fmt.Errorf("invalid query param value: %+v", val)
	}
	return strings.TrimSpace(val), nil
}

// A function to create a Metadata struct with key-value pairs
// converts variadic arguments into an Envelop (map[string]any).
func MakeEnvelop(args ...any) common.Envelop {
	// If there's only one argument, we assume it's an Envelop and return it directly
	if len(args) == 1 {
		if env, ok := args[0].(common.Envelop); ok {
			return env
		}
	}

	// If the number of arguments is odd, automatically add a default key-value pair
	if len(args)%2 != 0 {
		args = append(args, "extra_info", nil) // or another default key-value pair
	}

	envelop := make(common.Envelop)
	for i := 0; i < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			panic(fmt.Sprintf("Invalid key type at position %d. Expected string, but got: %T", i, args[i]))
		}
		envelop[key] = args[i+1]
	}
	return envelop
}

func MustParseDuration(val string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(val)
	if err != nil {
		return fallback
	}
	return d
}

// ConvertEnvelopToSlogArgs converts a map[string]any into a slice of any
// suitable for slog's variadic arguments.
func ConvertEnvelopToSlogArgs(envelop map[string]any) []any {
	if envelop == nil {
		return nil
	}
	args := make([]any, 0, len(envelop)*2) // Pre-allocate to avoid re-allocations
	for k, v := range envelop {
		args = append(args, k, v)
	}
	return args
}

func EnsureDirExists(dirPath string) error {
	// Get the absolute path
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		return fmt.Errorf("could not get absolute path: %w", err)
	}

	// Check if the directory exists
	info, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		// If directory doesn't exist, create it
		err = os.MkdirAll(absPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("could not create directory: %w", err)
		}
		return nil
	}

	// Check if it's a directory
	if !info.IsDir() {
		return fmt.Errorf("path %s is not a directory", absPath)
	}

	return nil
}

// func MakeEnvelop(args ...interface{}) common.Envelop {
// 	if len(args)%2 != 0 {
// 		panic("metadata arguments must be in key-value pairs (e.g., 'key1', value1, 'key2', value2, ...)")
// 	}

// 	envelop := make(common.Envelop)
// 	for i := 0; i < len(args); i += 2 {
// 		key, ok := args[i].(string)
// 		if !ok {
// 			panic(fmt.Sprintf("Invalid key type at position %d. Expected string, but got: %T", i, args[i]))
// 		}
// 		envelop[key] = args[i+1]
// 	}
// 	return envelop
// }
