import { useEffect, useState } from "react";
import { uploadTrace } from "../api";
import { useNavigate } from "react-router-dom";

const TRACE_STORAGE_KEY = "span_explainer_trace";
const UPLOAD_ID_STORAGE_KEY = "span_explainer_upload_id";

const SAMPLE_TRACE = `{
  "resourceSpans": [
    {
      "resource": {
        "attributes": [
          { "key": "service.name", "value": { "stringValue": "sample-service" } }
        ]
      },
      "scopeSpans": [
        {
          "scope": { "name": "sample-scope" },
          "spans": [
            {
              "traceId": "abc123",
              "spanId": "e1f7c9a2",
              "name": "GET /api/users",
              "kind": 1,
              "startTimeUnixNano": "1609459200000000000",
              "endTimeUnixNano": "1609459200500000000"
            }
          ]
        }
      ]
    }
  ]
}`;

export default function UploadTrace() {
  const [trace, setTrace] = useState<string>(
    () => localStorage.getItem(TRACE_STORAGE_KEY) ?? "",
  );

  const [uploadId, setUploadId] = useState<string | null>(() =>
    localStorage.getItem(UPLOAD_ID_STORAGE_KEY),
  );

  useEffect(() => {
    if (trace.trim()) {
      localStorage.setItem(TRACE_STORAGE_KEY, trace);
    } else {
      localStorage.removeItem(TRACE_STORAGE_KEY);
    }
  }, [trace]);

  useEffect(() => {
    if (uploadId) {
      localStorage.setItem(UPLOAD_ID_STORAGE_KEY, uploadId);
    } else {
      localStorage.removeItem(UPLOAD_ID_STORAGE_KEY);
    }
  }, [uploadId]);

  const navigate = useNavigate();
  async function handleUpload() {
    if (!trace.trim()) {
      alert("Trace JSON cannot be empty");
      return;
    }

    setUploadId(null);
    localStorage.removeItem(UPLOAD_ID_STORAGE_KEY);

    const res = await uploadTrace(trace);
    setUploadId(res.upload_id);
  }

  function handleLoadSample() {
    setTrace(SAMPLE_TRACE);
  }

  return (
    <div className="max-w-3xl mx-auto p-6">
      <h2 className="text-xl font-semibold mb-2">Upload Trace</h2>
      <p className="text-gray-600 mb-4">Paste your OTLP trace JSON below.</p>

      <div className="flex justify-end mb-2">
        <button
          onClick={handleLoadSample}
          className="text-sm border px-3 py-1 rounded-lg hover:bg-gray-50"
        >
          Load Sample Trace
        </button>
      </div>

      <textarea
        rows={12}
        className="w-full border rounded-lg p-3 font-mono text-sm"
        value={trace}
        onChange={(e) => setTrace(e.target.value)}
      />

      <button
        onClick={handleUpload}
        className="mt-4 bg-pink-600 text-white px-4 py-2 rounded-lg"
      >
        Upload Trace
      </button>

      {uploadId && (
        <div className="mt-4 bg-pink-50 p-3 rounded-lg">
          <p className="text-sm">
            <strong>Upload ID:</strong> {uploadId}
          </p>
          <button
            onClick={() => navigate("/explain", { state: { uploadId } })}
            className="mt-2 text-pink-600 underline"
          >
            Explain a span →
          </button>
        </div>
      )}

      <button
        onClick={() => {
          setTrace("");
          setUploadId(null);
          localStorage.removeItem(TRACE_STORAGE_KEY);
          localStorage.removeItem(UPLOAD_ID_STORAGE_KEY);
        }}
        className="mt-2 text-sm text-gray-500 underline"
      >
        Clear trace
      </button>
    </div>
  );
}
