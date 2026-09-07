// Sentinel errors. Every error returned by the library wraps one of these,
// so callers can classify failures with errors.Is without parsing messages:
//
//	if _, err := doc.GetString("app.name"); errors.Is(err, yamler.ErrNotFound) {
//		// key missing
//	}

package yamler

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound is returned when a path or key does not exist in the document.
	ErrNotFound = errors.New("not found")
	// ErrType is returned when a value has a different type than requested,
	// or a path navigates into a node of the wrong kind (e.g. an index into
	// a mapping).
	ErrType = errors.New("type mismatch")
	// ErrIndex is returned when an array index is out of range.
	ErrIndex = errors.New("index out of range")
	// ErrPath is returned when a path or pattern is malformed.
	ErrPath = errors.New("invalid path")
	// ErrRoot is returned when an operation is not applicable to the kind
	// of the document root (mapping vs sequence vs empty).
	ErrRoot = errors.New("invalid document root")
	// ErrParse is returned when YAML (or a schema) cannot be parsed.
	ErrParse = errors.New("parse error")
	// ErrIO is returned when a file cannot be read or written.
	ErrIO = errors.New("i/o error")
	// ErrMultiDocument is returned by Load for a stream with several documents.
	ErrMultiDocument = errors.New("multi-document stream")
	// ErrValidation is returned by Validate for every schema violation.
	ErrValidation = errors.New("validation failed")
	// ErrUnsupported is returned for Go values or YAML nodes the library
	// cannot convert.
	ErrUnsupported = errors.New("unsupported")
)

// wrapErr formats a message and attaches a sentinel error so that both the
// message and errors.Is(err, sentinel) work: "path a.b: key b not found".
func wrapErr(sentinel error, format string, args ...interface{}) error {
	return &libError{msg: fmt.Sprintf(format, args...), sentinel: sentinel}
}

// wrapCause is wrapErr for errors that also have an underlying cause (an
// os or yaml.v3 error); errors.Is matches both the sentinel and the cause.
func wrapCause(sentinel, cause error, format string, args ...interface{}) error {
	return &libError{msg: fmt.Sprintf(format, args...), sentinel: sentinel, cause: cause}
}

// libError carries a human-readable message, the sentinel it belongs to and
// an optional underlying cause.
type libError struct {
	msg      string
	sentinel error
	cause    error
}

func (e *libError) Error() string { return e.msg }

func (e *libError) Unwrap() []error {
	if e.cause != nil {
		return []error{e.sentinel, e.cause}
	}
	return []error{e.sentinel}
}
