package ag_conf

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// MulEl splits a string containing multiple EL expressions into individual expressions
// Returns an error if ${ and } are unbalanced
func MulEl(key string) ([]string, error) {
	subKeys := make([]string, 0, strings.Count(key, ConstPlaceholderPrefix)+1)

	// Fast path for non-EL strings
	if !strings.Contains(key, ConstPlaceholderPrefix) {
		return append(subKeys, key), nil
	}

	var start, end int
	for i := 0; i < len(key); {
		r, size := utf8.DecodeRuneInString(key[i:])
		i += size
		if r == '$' && i < len(key) && key[i] == '{' {
			if start < end {
				// fmt.Println("下一个元素:", key[start:end])
				subKeys = append(subKeys, key[start:end])
			}
			start = i - 1 // Include ${ in next segment
		} else if r == '}' {
			// fmt.Println("下一个元素:", key[start:i])
			subKeys = append(subKeys, key[start:i])
			start = i
			end = i
		}
		end = i
	}

	// Add remaining segment
	if start < end {
		subKeys = append(subKeys, key[start:end])
	}

	// Validate balanced braces
	if strings.Count(key, ConstPlaceholderPrefix) != strings.Count(key, ConstPlaceholderSuffix) {
		return nil, errors.New("${ and } must be balanced")
	}

	return subKeys, nil
}

// GetDefaultValue extracts default value from EL expression (format: key:default)
func GetDefaultValue(key string) (newkey, defaultVal string) {
	if idx := strings.Index(key, ConstValueSeparator); idx != -1 {
		return key[:idx], key[idx+1:]
	}
	return key, ""
}

// CheckEL checks if string contains EL expression markers
func CheckEL(key string) bool {
	return strings.Contains(key, ConstPlaceholderPrefix) &&
		strings.Contains(key, ConstPlaceholderSuffix)
}

// singleOrNot checks if string contains exactly one EL expression
func singleOrNot(key string) bool {
	return !CheckEL(key) || strings.Count(key, ConstPlaceholderPrefix) == 1
}

// EliminatePlaceholder removes EL expression markers and extracts default value
func EliminatePlaceholder(key string) (newkey, defaultVal string) {
	if !strings.HasPrefix(key, ConstPlaceholderPrefix) ||
		!strings.HasSuffix(key, ConstPlaceholderSuffix) {
		return key, ""
	}

	key = key[len(ConstPlaceholderPrefix) : len(key)-len(ConstPlaceholderSuffix)]
	return GetDefaultValue(key)
}
