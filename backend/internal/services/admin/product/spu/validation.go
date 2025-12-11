package spu

import (
	"fmt"
	"strings"
)

// ValidationError captures a single field level violation.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors aggregates all validation issues for transport-friendly reporting.
type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return "validation failed"
	}
	parts := make([]string, len(v))
	for i, err := range v {
		if err.Field != "" {
			parts[i] = fmt.Sprintf("%s: %s", err.Field, err.Message)
			continue
		}
		parts[i] = err.Message
	}
	return strings.Join(parts, "; ")
}

func (v ValidationErrors) add(field, message string) ValidationErrors {
	return append(v, ValidationError{Field: field, Message: message})
}

func (v ValidationErrors) empty() bool { return len(v) == 0 }
