package otelvalidator

import "errors"

var (
	ErrInvalidJSON     = errors.New("invalid JSON")
	ErrNotOTLPTraces   = errors.New("not a valid OTLP traces payload")
	ErrMissingResource = errors.New("missing resourceSpans")
	ErrMissingSpans    = errors.New("no spans found")
	ErrInvalidSpan     = errors.New("invalid span structure")
)
