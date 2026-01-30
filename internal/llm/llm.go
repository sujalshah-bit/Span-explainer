package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"strings"

	"github.com/sujalshah-bit/span-explainer/internal/trace"
)

type LLM interface {
	ExplainTrace(ctx context.Context, traceData json.RawMessage, question, spanID string) (*ExplainResult, error)
}

type PhiLLM struct{}

// ExplainTrace builds a comprehensive context from trace data and prepares it for LLM analysis
func (f *PhiLLM) ExplainTrace(ctx context.Context, traceData json.RawMessage, question, spanID string) (*ExplainResult, error) {
	explainCtx, err := trace.BuildExplainContext(traceData, spanID)
	if err != nil {
		return nil, fmt.Errorf("failed to build explain context: %w", err)
	}

	// Construct the prompt for the LLM
	_ = buildPrompt(explainCtx, question)

	// TODO: Call actual LLM with the prompt

	llmResponse := `{
		"root_cause": "Database connection timeout in payment-service",
		"impact": "User checkout failed, transaction rolled back",
		"suggested_action": "Increase connection pool size and add retry logic"
	}`

	// Validate the LLM response format
	answer, err := ParseResult(llmResponse)
	if err != nil {
		return nil, fmt.Errorf("invalid LLM response format: %w", err)
	}

	return answer, nil
}

// buildPrompt constructs a detailed, structured prompt for the LLM
func buildPrompt(ctx *trace.ExplainContext, question string) string {
	var b strings.Builder

	// System context with strict format requirement
	b.WriteString("You are an expert distributed systems engineer analyzing OpenTelemetry trace data.\n")
	b.WriteString("Your goal is to explain errors, performance issues, and trace anomalies clearly and actionably.\n\n")

	b.WriteString("CRITICAL: You MUST respond ONLY with valid JSON in this exact format:\n")
	b.WriteString("{\n")
	b.WriteString(`  "root_cause": "Brief description of what caused the error",` + "\n")
	b.WriteString(`  "impact": "What services/users were affected and how",` + "\n")
	b.WriteString(`  "suggested_action": "Concrete steps to fix or prevent this issue"` + "\n")
	b.WriteString("}\n\n")
	b.WriteString("Do NOT include any text before or after the JSON. Do NOT use markdown code blocks.\n")
	b.WriteString("Respond with ONLY the raw JSON object.\n\n")

	// User question
	if question != "" {
		b.WriteString(fmt.Sprintf("Question: %s\n\n", question))
	}

	// Target span (the error span being investigated)
	b.WriteString("=== TARGET SPAN (Error/Issue) ===\n")
	writeSpanContext(&b, &ctx.TargetSpan, "Target")
	b.WriteString("\n")

	// Parent span context (what called this span)
	if ctx.ParentSpan != nil {
		b.WriteString("=== PARENT SPAN (Caller Context) ===\n")
		writeSpanContext(&b, ctx.ParentSpan, "Parent")
		b.WriteString("\n")
	}

	// Child spans context (what this span called)
	if len(ctx.ChildSpans) > 0 {
		b.WriteString("=== CHILD SPANS (Downstream Calls) ===\n")
		for i, child := range ctx.ChildSpans {
			writeSpanContext(&b, &child, fmt.Sprintf("Child %d", i+1))
			if i < len(ctx.ChildSpans)-1 {
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")
	}

	// Analysis focus areas
	b.WriteString("=== ANALYSIS FOCUS ===\n")
	b.WriteString("In your JSON response, ensure:\n")
	b.WriteString("- root_cause: Identifies the technical failure point (service, operation, dependency)\n")
	b.WriteString("- impact: Describes business/user impact and affected components\n")
	b.WriteString("- suggested_action: Provides actionable remediation steps\n")

	return b.String()
}

// writeSpanContext formats a single span's context in a readable way
func writeSpanContext(b *strings.Builder, span *trace.SpanContext, label string) {
	b.WriteString(fmt.Sprintf("%s Span:\n", label))
	b.WriteString(fmt.Sprintf("  Service: %s\n", span.Service))
	b.WriteString(fmt.Sprintf("  Operation: %s\n", span.Operation))
	b.WriteString(fmt.Sprintf("  Status: %s\n", span.Status))

	// Write tags/attributes
	if len(span.Tags) > 0 {
		b.WriteString("  Tags:\n")
		for k, v := range span.Tags {
			b.WriteString(fmt.Sprintf("    %s: %s\n", k, v))
		}
	}

	// Write logs/events
	if len(span.Logs) > 0 {
		b.WriteString("  Logs/Events:\n")
		for _, log := range span.Logs {
			b.WriteString(fmt.Sprintf("    - %s\n", log))
		}
	}
}

// validateLLMResponse ensures the LLM response is valid JSON matching ExplainResult structure
func ParseResult(response string) (*ExplainResult, error) {
	// Trim whitespace
	response = strings.TrimSpace(response)

	// Try to unmarshal into ExplainResult
	result, err := ParseExplainResult(response)
	if err != nil {
		return nil, fmt.Errorf("response is not valid JSON: %w", err)
	}

	// Validate required fields are non-empty
	if result.RootCause == "" {
		return nil, fmt.Errorf("root_cause field is missing or empty")
	}
	if result.Impact == "" {
		return nil, fmt.Errorf("impact field is missing or empty")
	}
	if result.SuggestedAction == "" {
		return nil, fmt.Errorf("suggested_action field is missing or empty")
	}

	return result, nil
}

// ParseExplainResult is a helper function to parse validated LLM response
func ParseExplainResult(response string) (*ExplainResult, error) {
	var result ExplainResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("failed to parse explain result: %w", err)
	}
	return &result, nil
}
