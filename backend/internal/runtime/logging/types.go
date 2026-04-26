package logging

import "time"

type Fields map[string]any

type Event struct {
	Message   string
	Level     string
	Fields    Fields
	Timestamp time.Time
}
