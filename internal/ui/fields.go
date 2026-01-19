package ui

import "time"

type FieldType int

const (
	FieldBool FieldType = iota
	FieldString
	FieldInt
	FieldDuration
)

type Field struct {
	Label string
	Type  FieldType

	// pointers into cfg
	Bool     *bool
	String   *string
	Int      *int
	Duration *time.Duration
}
