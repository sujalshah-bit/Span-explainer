# Span Explainer – API Contract (Frontend Reference)

This document defines the **API contract** for the Span Explainer backend.

---

## Base URL

```
http://<HOST>:<PORT>
```

Example (local):

```
http://localhost:9000
```

All requests and responses use **JSON**.

---

## Authentication Model

* Authentication is **token-based**.
* Client must first call **Register API** to obtain:

  * `user_id`
  * `token`
* For all protected APIs, send:

```
Authorization: Bearer <token>
```

The backend internally resolves the user via this token.

---

## 1. Register User

### Endpoint

```
POST /register
```

### Auth

❌ Not required

### Request Body

*No body required*

### Response – 200 OK

```json
{
  "user_id": "string",
  "token": "string"
}
```

### Notes

* Call this **once** when the app loads (or on first visit).
* Store the token securely (e.g., memory, localStorage).

---

## 2. Upload Trace

Uploads an **OTLP trace JSON** and stores it for later querying.

### Endpoint

```
POST /upload-trace
```

### Auth

✅ Required

Headers:

```
Authorization: Bearer <token>
Content-Type: application/json
```

### Request Body

Raw **OTLP Traces JSON** (exact format expected by OpenTelemetry).

Example (simplified):

```json
{
  "resourceSpans": [
    {
      "resource": { "attributes": [] },
      "scopeSpans": []
    }
  ]
}
```

### Success Response – 200 OK

```json
{
  "upload_id": "string"
}
```

### Error Responses

| Status | Meaning                   |
| ------ | ------------------------- |
| 400    | Invalid JSON              |
| 401    | Missing / invalid token   |
| 422    | Invalid OTLP trace format |

### Notes for Frontend

* Treat `upload_id` as the **primary trace identifier**.
* Store it in state; required for LLM queries.

---

## 3. Query LLM (Explain Trace / Span)

Asks an LLM a question about a previously uploaded trace.

### Endpoint

```
POST /explain-span
```

### Auth

✅ Required

Headers:

```
Authorization: Bearer <token>
Content-Type: application/json
```

### Request Body

```json
{
  "upload_id": "string",
  "span_id": "string",
  "question": "string"
}
```

### Field Semantics

| Field     | Required | Description                                                                          |
| --------- | -------- | ------------------------------------------------------------------------------------ |
| upload_id | ✅        | ID returned from `/upload-trace`                                                     |
| span_id   | ✅        | Span ID inside the trace (can be empty if full trace explanation is supported later) |
| question  | ✅        | Natural language question for the LLM                                                |

Example:

```json
{
  "upload_id": "8f0c9c0e-5a1e-4b7d-9c4c-3f6a2c",
  "span_id": "e1f7c9a2",
  "question": "Why is this span taking so long?"
}
```

### Success Response – 200 OK

```json
{
  "answer": {
    "summary": "string",
    "details": "string",
    "suggestions": ["string"]
  }
}
```

> ⚠️ The exact structure of `answer` depends on `llm.ExplainResult`.
> Frontend should treat it as a **structured explanation object**, not plain text.

### Error Responses

| Status | Meaning                    |
| ------ | -------------------------- |
| 400    | Bad request (invalid JSON) |
| 401    | Unauthorized               |
| 404    | Trace not found            |
| 500    | LLM processing error       |



```bash


I have developed the application which explain the error span

I want to test it for that i need test cases whether the llm is explaining correctly different types of error or problem.  



I want you to create a json file full of test cases each case will have three fields 



name : string

trace: json (you create otel traces it shouldn't be too long) 

span_id

llm answer: json(format

{

  "root_cause": "string",

  "impact": "impact",

  "suggested_action": ""

}

)

expected answer

```