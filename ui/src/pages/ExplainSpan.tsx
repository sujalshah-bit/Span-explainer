import { useLocation } from "react-router-dom";
import { useState } from "react";
import { explainSpan } from "../api";
import type { ExplainSpanAnswer } from "../types";

interface LocationState {
  uploadId?: string;
}

export default function ExplainSpan() {
  const location = useLocation();
  const state = location.state as LocationState | null;

  const [uploadId, setUploadId] = useState<string>(
    state?.uploadId ?? ""
  );
  const [spanId, setSpanId] = useState<string>("");
  const [question, setQuestion] = useState<string>("");
  const [answer, setAnswer] = useState<ExplainSpanAnswer | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function handleExplain() {
    setError(null);

    if (!uploadId.trim()|| !spanId.trim() || !question.trim()) {
      setError("Required field is missing!!");
      return;
    }

    const payload = {
      upload_id: uploadId.trim(),
      span_id: spanId.trim(),
      question: question.trim(),
    };

    console.log("Explain payload:", payload);

    try {
      const res = await explainSpan(payload);
      console.log(res.answer.impact)
      setAnswer(res.answer);
    } catch {
      setError("Failed to get explanation. Please check inputs.");
    }
  }

  return (
    <div className="max-w-3xl mx-auto p-6">
      <h2 className="text-xl font-semibold mb-4">Explain Span</h2>

      {/* Upload ID */}
      <input
        className="w-full border rounded-lg p-2 mb-3"
        placeholder="Upload ID"
        value={uploadId}
        onChange={(e) => setUploadId(e.target.value)}
      />

      {/* Span ID */}
      <input
        className="w-full border rounded-lg p-2 mb-3"
        placeholder="Span ID (optional)"
        value={spanId}
        onChange={(e) => setSpanId(e.target.value)}
      />

      {/* Question */}
      <textarea
        className="w-full border rounded-lg p-3 mb-3"
        rows={4}
        placeholder="Ask a question about this span"
        value={question}
        onChange={(e) => setQuestion(e.target.value)}
      />

      {error && (
        <p className="text-sm text-red-600 mb-3">
          {error}
        </p>
      )}

      <button
        onClick={handleExplain}
        className="bg-pink-600 text-white px-4 py-2 rounded-lg"
      >
        Get Explanation
      </button>

      {answer && (
        <div className="mt-6 bg-white shadow rounded-xl p-4">
          <h3 className="font-semibold">Impact</h3>
          <p className="mb-3">{answer.impact}</p>

          <h3 className="font-semibold">RootCause</h3>
          <p className="mb-3">{answer.root_cause}</p>

          <h3 className="font-semibold">Suggestions</h3>
          <ul className="list-disc pl-5">
            {answer.suggested_action}
          </ul>
        </div>
      )}
    </div>
  );
}