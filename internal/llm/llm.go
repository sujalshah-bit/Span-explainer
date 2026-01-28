package llm

import (
	"context"
	"encoding/json"
)

type LLM interface {
	ExplainTrace(ctx context.Context, trace json.RawMessage, question string) (string, error)
}

type FakeLLM struct{}

func (f *FakeLLM) ExplainTrace(ctx context.Context, trace json.RawMessage, question string) (string, error) {
	return "explanation from fake LLM", nil
}
