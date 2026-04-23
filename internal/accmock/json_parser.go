package accmock

import (
	"encoding/json"
	"fmt"
	"strings"
)

func getItem(contents []byte, itemPath string) (item string, data any, err error) {
	if err := json.Unmarshal(contents, &data); err != nil {
		return "", nil, err
	}

	if item, err = extractItem(data, itemPath); err != nil {
		return "", nil, err
	}

	return item, data, nil
}

func extractItem(parsedData any, itemPath string) (string, error) {
	if itemPath == "" {
		return "", fmt.Errorf("item path is empty")
	}

	current := parsedData
	for _, part := range strings.Split(itemPath, ".") {
		object, ok := current.(map[string]any)
		if !ok {
			return "", fmt.Errorf("path %q cannot descend into %T at %q", itemPath, current, part)
		}

		next, ok := object[part]
		if !ok {
			return "", fmt.Errorf("path %q not found: missing key %q", itemPath, part)
		}

		current = next
	}

	return stringifyItem(current, itemPath)
}

func stringifyItem(value any, itemPath string) (string, error) {
	switch typed := value.(type) {
	case nil:
		return "", fmt.Errorf("path %q resolved to null", itemPath)
	case string:
		return typed, nil
	case float64, bool:
		return fmt.Sprint(typed), nil
	case []any:
		if len(typed) == 0 {
			return "", fmt.Errorf("path %q resolved to an empty array", itemPath)
		}

		return stringifyItem(typed[0], itemPath)
	case map[string]any:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return "", fmt.Errorf("marshal object at %q: %w", itemPath, err)
		}

		return string(encoded), nil
	default:
		return fmt.Sprint(typed), nil
	}
}
