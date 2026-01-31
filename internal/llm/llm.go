package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sujalshah-bit/span-explainer/internal/trace"
	"github.com/sujalshah-bit/span-explainer/internal/util"
	"github.com/tmc/langchaingo/llms/ollama"
)

type LLM interface {
	ExplainTrace(ctx context.Context, traceData json.RawMessage, question, spanID string) (*ExplainResult, error)
}

type PhiLLM struct {
	model *ollama.LLM
}

func NewPhiLLM() (*PhiLLM, error) {
	llm, err := ollama.New(
		ollama.WithModel("phi3:mini"),
		ollama.WithServerURL("http://192.168.29.62:11434"),
	)
	if err != nil {
		return nil, err
	}

	return &PhiLLM{model: llm}, nil
}

// ExplainTrace builds a comprehensive context from trace data and prepares it for LLM analysis
func (pllm *PhiLLM) ExplainTrace(ctx context.Context, traceData json.RawMessage, question, spanID string) (*ExplainResult, error) {
	explainCtx, err := trace.BuildExplainContext(traceData, spanID)
	if err != nil {
		return nil, fmt.Errorf("failed to build explain context: %w", err)
	}

	// Construct the prompt for the LLM
	prompt := buildPrompt(explainCtx, question)

	// llmResponse := `{
	// 	"root_cause": "Database connection timeout in payment-service",
	// 	"impact": "User checkout failed, transaction rolled back",
	// 	"suggested_action": "Increase connection pool size and add retry logic"
	// }`

	fmt.Println()
	fmt.Println(prompt)
	fmt.Println()
	llmResponse, err := pllm.model.Call(context.Background(), prompt)
	fmt.Println()
	fmt.Println(llmResponse)
	fmt.Println()

	if err != nil {
		return nil, err
	}

	// Validate the LLM response format
	answer, err := ParseResult(util.ExtractJSON(llmResponse))
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

	// CRITICAL: Enhanced JSON format requirements
	b.WriteString("CRITICAL JSON FORMAT REQUIREMENTS:\n")
	b.WriteString("You MUST respond with ONLY valid JSON in this EXACT format:\n")
	b.WriteString("{\n")
	b.WriteString(`  "root_cause": "Brief description of what caused the error",` + "\n")
	b.WriteString(`  "impact": "What services/users were affected and how",` + "\n")
	b.WriteString(`  "suggested_action": "Concrete steps to fix or prevent this issue"` + "\n")
	b.WriteString("}\n\n")

	b.WriteString("STRICT RULES:\n")
	b.WriteString("1. NO comments allowed (no // or /* */)\n")
	b.WriteString("2. NO nested objects - ALL three fields MUST be simple strings\n")
	b.WriteString("3. NO markdown code blocks (no ```json or ```)\n")
	b.WriteString("4. NO explanatory text before or after the JSON\n")
	b.WriteString("5. NO null values - if you don't have information, write 'Not specified in trace data'\n")
	b.WriteString("6. Each field value MUST be a single string (not an object, not an array)\n\n")

	b.WriteString("VALID EXAMPLE:\n")
	b.WriteString(`{"root_cause":"TimeoutError - query exceeded 30 second limit","impact":"GET /api/users endpoint failed with 500 status, affecting all user list requests","suggested_action":"Increase timeout to 60 seconds or add index on users.active column"}` + "\n\n")

	b.WriteString("INVALID EXAMPLES (DO NOT DO THIS):\n")
	b.WriteString(`❌ {"root_cause": "error", "impact": {"service": "x"}} // NO nested objects!` + "\n")
	b.WriteString(`❌ {"root_cause": "error" // comment} // NO comments!` + "\n")
	b.WriteString(`❌ {"suggested_action": ["step1", "step2"]} // NO arrays!` + "\n\n")

	// Detailed extraction instructions
	b.WriteString("=== EXTRACTION REQUIREMENTS ===\n")
	b.WriteString("When analyzing the span data, you MUST:\n")
	b.WriteString("1. Extract and include EXACT numeric values (timeouts, limits, thresholds, counts, durations)\n")
	b.WriteString("2. Quote specific error messages and error types verbatim from logs/events\n")
	b.WriteString("3. Include configuration values (memory limits, connection pool sizes, TTLs)\n")
	b.WriteString("4. Reference specific file paths, URLs, or resource names when present\n")
	b.WriteString("5. Mention the exact HTTP status codes, database error codes, or exception types\n")
	b.WriteString("6. Include time durations with units (seconds, milliseconds, hours)\n")
	b.WriteString("7. When errors mention rate limits or retry periods, include the exact duration window\n")
	b.WriteString("8. When suggesting changes, mention both current and recommended values when possible\n\n")

	b.WriteString("EXAMPLES of good vs bad detail level:\n")
	b.WriteString("❌ BAD: 'Database timeout occurred'\n")
	b.WriteString("✅ GOOD: 'Database connection timeout - query exceeded 30 second limit'\n\n")
	b.WriteString("❌ BAD: 'Circuit breaker triggered due to high error rate'\n")
	b.WriteString("✅ GOOD: 'Circuit breaker in OPEN state due to 15 failures exceeding threshold of 10'\n\n")
	b.WriteString("❌ BAD: 'Memory issue in background task'\n")
	b.WriteString("✅ GOOD: 'OutOfMemoryError - heap memory maxed out at 4096MB during batch processing'\n\n")

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

	// Analysis focus areas (ENHANCED)
	b.WriteString("=== ANALYSIS FOCUS ===\n")
	b.WriteString("In your JSON response, ensure:\n")
	b.WriteString("- root_cause (SINGLE STRING): Identify the technical failure point WITH specific details:\n")
	b.WriteString("  * Include exact error type/exception class (e.g., 'TimeoutError', 'NullPointerException')\n")
	b.WriteString("  * Include numeric limits/thresholds if present (e.g., '30 second timeout', '4096MB heap')\n")
	b.WriteString("  * Include specific resource names (e.g., table names, API endpoints, file paths)\n")
	b.WriteString("- impact (SINGLE STRING): Describe business/user impact with specifics:\n")
	b.WriteString("  * Which exact endpoint/operation failed (e.g., 'POST /api/orders', 'GET /api/users')\n")
	b.WriteString("  * Scope of impact (e.g., 'all user list requests', 'next 3600 seconds')\n")
	b.WriteString("  * HTTP status codes or error responses users received\n")
	b.WriteString("- suggested_action (SINGLE STRING): Provide actionable remediation with technical details:\n")
	b.WriteString("  * Specific configuration changes with values (e.g., 'increase heap to 8192MB')\n")
	b.WriteString("  * Exact commands or parameters (e.g., 'chmod 755 /var/uploads')\n")
	b.WriteString("  * Reference specific tools/techniques (e.g., 'add index on users.email column')\n")
	b.WriteString("  * If multiple steps, separate with semicolons or write as prose, NOT as nested objects/arrays\n\n")

	b.WriteString("IMPORTANT: Look carefully at the Tags and Logs sections - they contain the specific values you need!\n\n")

	b.WriteString("FINAL REMINDER: Output ONLY the JSON object with three string fields. No comments, no nesting, no extra text.\n")

	return b.String()
}

// writeSpanContext formats a single span's context in a readable way
// IMPROVED VERSION - highlights critical error information and numeric values
func writeSpanContext(b *strings.Builder, span *trace.SpanContext, label string) {
	b.WriteString(fmt.Sprintf("%s Span:\n", label))
	b.WriteString(fmt.Sprintf("  Service: %s\n", span.Service))
	b.WriteString(fmt.Sprintf("  Operation: %s\n", span.Operation))
	b.WriteString(fmt.Sprintf("  Status: %s\n", span.Status))

	// Write tags/attributes - IMPROVED: Highlight error-related and numeric tags
	if len(span.Tags) > 0 {
		b.WriteString("  Tags/Attributes:\n")

		// First, prioritize error-related tags
		errorTags := []string{}
		numericTags := []string{}
		otherTags := []string{}

		for k, v := range span.Tags {
			kLower := strings.ToLower(k)
			tag := fmt.Sprintf("%s: %s", k, v)

			// Categorize tags for better visibility
			if strings.Contains(kLower, "error") ||
				strings.Contains(kLower, "exception") ||
				strings.Contains(kLower, "failure") ||
				strings.Contains(kLower, "status") ||
				k == "http.status_code" {
				errorTags = append(errorTags, tag)
			} else if isNumericValue(v) ||
				strings.Contains(kLower, "timeout") ||
				strings.Contains(kLower, "limit") ||
				strings.Contains(kLower, "threshold") ||
				strings.Contains(kLower, "size") ||
				strings.Contains(kLower, "count") ||
				strings.Contains(kLower, "duration") ||
				strings.Contains(kLower, "ttl") ||
				strings.Contains(kLower, "retry") ||
				strings.Contains(kLower, "attempt") {
				numericTags = append(numericTags, tag)
			} else {
				otherTags = append(otherTags, tag)
			}
		}

		// Write error tags first (most important)
		if len(errorTags) > 0 {
			b.WriteString("    [ERROR DETAILS]:\n")
			for _, tag := range errorTags {
				b.WriteString("      " + tag + "\n")
			}
		}

		// Write numeric/threshold tags second
		if len(numericTags) > 0 {
			b.WriteString("    [NUMERIC VALUES/LIMITS]:\n")
			for _, tag := range numericTags {
				b.WriteString("      " + tag + "\n")
			}
		}

		// Write other tags last
		if len(otherTags) > 0 {
			b.WriteString("    [OTHER]:\n")
			for _, tag := range otherTags {
				b.WriteString("      " + tag + "\n")
			}
		}
	}

	// Write logs/events - IMPROVED: Highlight error messages
	if len(span.Logs) > 0 {
		b.WriteString("  Logs/Events:\n")
		for _, log := range span.Logs {
			// Check if log contains error information
			logLower := strings.ToLower(log)
			if strings.Contains(logLower, "error") ||
				strings.Contains(logLower, "exception") ||
				strings.Contains(logLower, "failed") ||
				strings.Contains(logLower, "timeout") {
				b.WriteString(fmt.Sprintf("    [ERROR] %s\n", log))
			} else {
				b.WriteString(fmt.Sprintf("    - %s\n", log))
			}
		}
	}
}

// Helper function to detect numeric values
func isNumericValue(s string) bool {
	// Check if string contains digits and common numeric patterns
	hasDigit := false
	for _, r := range s {
		if r >= '0' && r <= '9' {
			hasDigit = true
			break
		}
	}
	return hasDigit && (strings.Contains(s, "MB") ||
		strings.Contains(s, "GB") ||
		strings.Contains(s, "KB") ||
		strings.Contains(s, "ms") ||
		strings.Contains(s, "seconds") ||
		strings.Contains(s, "sec") ||
		strings.Contains(s, "min") ||
		strings.Contains(s, "hour") ||
		strings.Contains(s, "%") ||
		len(s) < 50) // Likely a number if short and has digits
}

// ParseResult validates and parses the LLM response
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
