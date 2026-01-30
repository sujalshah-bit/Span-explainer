package trace

import (
	"encoding/json"
	"errors"
)

func buildSpanContext(
	sp Span,
	service string,
) SpanContext {
	logs := make([]string, 0, len(sp.Events))
	for _, ev := range sp.Events {
		logs = append(logs, ev.Name)
	}

	return SpanContext{
		Service:   service,
		Operation: sp.Name,
		Status:    sp.Status.Code,
		Tags:      sp.Attributes,
		Logs:      logs,
	}
}

func BuildExplainContext(
	raw json.RawMessage,
	targetSpanID string,
) (*ExplainContext, error) {

	t, err := ParseTrace(raw)
	if err != nil {
		return nil, err
	}

	idx := BuildTraceIndex(t)

	target, ok := findSpan(idx, targetSpanID)
	if !ok {
		return nil, errors.New("target span not found")
	}

	ctx := &ExplainContext{
		TargetSpan: buildSpanContext(
			target,
			idx.ServiceBySpan[target.SpanID],
		),
	}

	if target.ParentSpanID != "" {
		if parent, ok := findSpan(idx, target.ParentSpanID); ok {
			pctx := buildSpanContext(
				parent,
				idx.ServiceBySpan[parent.SpanID],
			)
			ctx.ParentSpan = &pctx
		}
	}

	// We can configure this, instead of hardcoded
	children := findChildren(idx, target.SpanID, 2)
	for _, ch := range children {
		ctx.ChildSpans = append(
			ctx.ChildSpans,
			buildSpanContext(
				ch,
				idx.ServiceBySpan[ch.SpanID],
			),
		)
	}

	return ctx, nil
}
