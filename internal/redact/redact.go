package redact

import (
	"fmt"
	"regexp"
)

const redactedPlaceholder = "REDACTED"

// Redactor scrubs sensitive values from plan data.
type Redactor struct {
	patterns []*regexp.Regexp
}

// New creates a Redactor with the given regex patterns.
// If useDefaults is true, built-in patterns are prepended.
func New(useDefaults bool, extra []*regexp.Regexp) *Redactor {
	var patterns []*regexp.Regexp
	if useDefaults {
		patterns = append(patterns, DefaultPatterns()...)
	}
	patterns = append(patterns, extra...)
	return &Redactor{patterns: patterns}
}

// RedactValue replaces the value with REDACTED if it is flagged sensitive
// (by Terraform's sensitive markers) or matches any regex pattern.
func (r *Redactor) RedactValue(val interface{}, isSensitive bool) interface{} {
	if isSensitive {
		return redactedPlaceholder
	}
	return r.redactByPattern(val)
}

// RedactMap redacts values in a map based on a sensitivity map and regex patterns.
// sensitiveMap mirrors the structure of values: true means the field is sensitive,
// a nested map means recurse.
func (r *Redactor) RedactMap(values map[string]interface{}, sensitiveMap interface{}) map[string]interface{} {
	if values == nil {
		return nil
	}

	result := make(map[string]interface{}, len(values))
	sensMap, _ := sensitiveMap.(map[string]interface{})

	for k, v := range values {
		var fieldSensitive interface{}
		if sensMap != nil {
			fieldSensitive = sensMap[k]
		}

		switch fs := fieldSensitive.(type) {
		case bool:
			if fs {
				result[k] = redactedPlaceholder
			} else {
				result[k] = r.redactByPattern(v)
			}
		case map[string]interface{}:
			if nested, ok := v.(map[string]interface{}); ok {
				result[k] = r.RedactMap(nested, fs)
			} else {
				result[k] = r.redactByPattern(v)
			}
		default:
			result[k] = r.redactByPattern(v)
		}
	}
	return result
}

// redactByPattern checks a value against all configured regex patterns.
func (r *Redactor) redactByPattern(val interface{}) interface{} {
	s, ok := stringify(val)
	if !ok {
		return val
	}
	for _, p := range r.patterns {
		if p.MatchString(s) {
			return redactedPlaceholder
		}
	}
	return val
}

func stringify(val interface{}) (string, bool) {
	switch v := val.(type) {
	case string:
		return v, true
	case fmt.Stringer:
		return v.String(), true
	default:
		return "", false
	}
}
