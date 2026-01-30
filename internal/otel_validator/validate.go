package otelvalidator

import (
	"encoding/json"
	"fmt"

	"github.com/sujalshah-bit/span-explainer/internal/trace"
)

// ValidateOTLPTraces validates the OTLP trace format using the trace package structs
func ValidateOTLPTraces(raw json.RawMessage) error {
	// Use the same trace.Trace struct from trace package
	var payload trace.Trace

	if err := json.Unmarshal(raw, &payload); err != nil {
		return ErrInvalidJSON
	}

	if len(payload.ResourceSpans) == 0 {
		return ErrMissingResource
	}

	spanCount := 0

	for _, rs := range payload.ResourceSpans {
		for _, ss := range rs.ScopeSpans {
			for _, sp := range ss.Spans {
				spanCount++

				if !isValidSpan(sp) {
					return fmt.Errorf("%w: traceId=%s spanId=%s",
						ErrInvalidSpan, sp.TraceID, sp.SpanID)
				}
			}
		}
	}

	if spanCount == 0 {
		return ErrMissingSpans
	}

	return nil
}

// Use trace.Span directly
func isValidSpan(s trace.Span) bool {
	return s.TraceID != "" &&
		s.SpanID != "" &&
		s.Name != ""
}
