package llm

type ExplainContext struct {
	TargetSpan SpanContext   `json:"target_span"`
	ParentSpan *SpanContext  `json:"parent_span,omitempty"`
	ChildSpans []SpanContext `json:"child_spans,omitempty"`
}

type SpanContext struct {
	Service   string            `json:"service"`
	Operation string            `json:"operation"`
	Status    string            `json:"status"`
	Duration  string            `json:"duration"`
	Tags      map[string]string `json:"tags"`
	Logs      []string          `json:"logs"`
}

type ExplainResult struct {
	RootCause       string `json:"root_cause"`
	Impact          string `json:"impact"`
	SuggestedAction string `json:"suggested_action"`
}
