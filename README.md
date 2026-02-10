# Span Explainer (Experiment)

https://github.com/user-attachments/assets/d17ebf09-791f-472c-8dec-c640887944df

## Overview

Span Explainer is an experimental backend application that analyzes OpenTelemetry traces and explains **error spans** using a structured, deterministic LLM output.

Instead of acting as a generic chatbot, the system focuses on **bounded, low-hallucination tasks** that small language models (SLMs) handle well—specifically **error summarization**.


## What Problem This Solves

When debugging distributed systems:

* Error spans contain raw logs, tags, and stack traces
* Root causes are not immediately obvious
* Users manually inspect multiple attributes to understand failures

This project answers a simple question:

> **“Given a specific error span, what went wrong, why it matters, and what to do next?”**

## Core Capability: Explain Error Span

Given:

* An uploaded OTEL trace
* A specific `span_id`
* A natural language question

The system returns a **strictly structured explanation**:

```json
{
  "root_cause": "string",
  "impact": "string",
  "suggested_action": "string"
}
```

## Request Flow

1. **Register User**

   * Client receives a token
2. **Upload Trace**

   * Raw OTLP JSON is validated and stored
   * An `upload_id` is returned
3. **Explain Span**

   * Client provides `upload_id`, `span_id`, and question
   * Backend:

     * Extracts only relevant span data
     * Prunes noise
     * Sends minimal context to the LLM
4. **Structured Response**

   * LLM output is validated against the schema
   * Invalid or hallucinated responses are rejected

## Testing Strategy

### Why Testing LLMs Is Hard

Exact string matching does not work for LLMs.
Instead, this project evaluates **semantic correctness**.

### Test Suite

* `error_span_test_cases.json`

  * Covers common production failures:

    * Timeouts
    * Null pointer exceptions
    * Rate limits
    * Permission issues
* Each test defines:

  * Input trace
  * Span ID
  * Expected structured answer

### Semantic Evaluation

* LLM outputs are compared with expected answers
* Similarity is measured **semantically**, not textually
* Results are saved in `llm_test_results.json`

### Metrics & Visualization

The `metrics.py` script:

* Computes semantic similarity scores
* Generates visual charts using seaborn
* Saves results in `metrics-assessments/`

<img width="5370" height="3569" alt="llm_evaluation_dashboard" src="https://github.com/user-attachments/assets/c940a846-d5e6-427d-95af-d483827e5e11" />

<img width="5970" height="2371" alt="llm_evaluation_detailed" src="https://github.com/user-attachments/assets/72cf72eb-1195-413c-b1f9-c60bb1fab1e9" />


This provides **quantitative evidence** of how reliably the LLM explains error spans.


## Why This Works with Small Language Models

* The task is **pure summarization**, not reasoning
* Input context is aggressively pruned
* Output format is strict and validated
* No open-ended chat or speculation

This aligns with the design philosophy of:

* Local-first AI
* Deterministic behavior
* Production-safe outputs

## Project Status

⚠️ **Experimental / Prototype**

This project is **not a production system**.
It exists to explore feasibility and validate prompt design


* Tune expected answers for **higher semantic scores**
* Review `metrics.py` and suggest better scoring thresholds
