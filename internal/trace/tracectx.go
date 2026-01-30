package trace

import (
	"encoding/json"
)

func ParseTrace(raw json.RawMessage) (*Trace, error) {
	var t Trace
	if err := json.Unmarshal(raw, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

type TraceIndex struct {
	SpansByID     map[string]Span
	Children      map[string][]Span
	ServiceBySpan map[string]string
}

// Helper function to extract service name from resource attributes
func getServiceName(attrs []Attribute) string {
	for _, attr := range attrs {
		if attr.Key == "service.name" {
			// Handle OTel value format: {"stringValue": "actual-value"}
			if sv, ok := attr.Value["stringValue"].(string); ok {
				return sv
			}
			// Fallback for other possible formats
			if sv, ok := attr.Value["value"].(string); ok {
				return sv
			}
		}
	}
	return ""
}

func BuildTraceIndex(t *Trace) *TraceIndex {
	idx := &TraceIndex{
		SpansByID:     make(map[string]Span),
		Children:      make(map[string][]Span),
		ServiceBySpan: make(map[string]string),
	}

	for _, rs := range t.ResourceSpans {
		// Extract service name from attributes array
		service := getServiceName(rs.Resource.Attributes)

		for _, ss := range rs.ScopeSpans {
			for _, sp := range ss.Spans {
				idx.SpansByID[sp.SpanID] = sp
				idx.ServiceBySpan[sp.SpanID] = service

				if sp.ParentSpanID != "" {
					idx.Children[sp.ParentSpanID] =
						append(idx.Children[sp.ParentSpanID], sp)
				}
			}
		}
	}
	return idx
}
