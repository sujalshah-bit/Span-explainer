package otelvalidator

import (
	"encoding/json"
	"fmt"
)

type otlpTraces struct {
	ResourceSpans []resourceSpans `json:"resourceSpans"`
}

type resourceSpans struct {
	Resource   json.RawMessage `json:"resource"`
	ScopeSpans []scopeSpans    `json:"scopeSpans"`
}

type scopeSpans struct {
	Spans []span `json:"spans"`
}

type span struct {
	TraceID string `json:"traceId"`
	SpanID  string `json:"spanId"`
	Name    string `json:"name"`
}

func ValidateOTLPTraces(raw json.RawMessage) error {
	var payload otlpTraces

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

func isValidSpan(s span) bool {
	return s.TraceID != "" &&
		s.SpanID != "" &&
		s.Name != ""
}
