package trace

type SpanContext struct {
	Service   string            `json:"service"`
	Operation string            `json:"operation"`
	Status    string            `json:"status"`
	Tags      map[string]string `json:"tags"`
	Logs      []string          `json:"logs"`
}

func findSpan(idx *TraceIndex, spanID string) (Span, bool) {
	sp, ok := idx.SpansByID[spanID]
	return sp, ok
}

func findChildren(idx *TraceIndex, spanID string, limit int) []Span {
	children := idx.Children[spanID]

	if len(children) > limit {
		return children[:limit]
	}
	return children
}
