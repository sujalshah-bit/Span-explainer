package trace

type Trace struct {
	ResourceSpans []ResourceSpan `json:"resourceSpans"`
}

type ResourceSpan struct {
	Resource   Resource    `json:"resource"`
	ScopeSpans []ScopeSpan `json:"scopeSpans"`
}

type Resource struct {
	Attributes []Attribute `json:"attributes"` // Array, not map
}

type Attribute struct {
	Key   string                 `json:"key"`
	Value map[string]interface{} `json:"value"`
}

type ScopeSpan struct {
	Spans []Span `json:"spans"`
}

type Span struct {
	TraceID      string `json:"traceId"`
	SpanID       string `json:"spanId"`
	ParentSpanID string `json:"parentSpanId,omitempty"`
	Name         string `json:"name"`
	StartTime    string `json:"startTimeUnixNano,omitempty"`
	EndTime      string `json:"endTimeUnixNano,omitempty"`

	Status struct {
		Code string `json:"code"`
	} `json:"status"`

	Attributes map[string]string `json:"attributes,omitempty"`
	Events     []Event           `json:"events,omitempty"`
}

type Event struct {
	Name string `json:"name"`
}

type ExplainContext struct {
	TargetSpan SpanContext   `json:"target_span"`
	ParentSpan *SpanContext  `json:"parent_span,omitempty"`
	ChildSpans []SpanContext `json:"child_spans,omitempty"`
}
