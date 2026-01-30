import type{
  RegisterResponse,
  UploadTraceResponse,
  ExplainSpanRequest,
  ExplainSpanResponse,
} from "./types";

const BASE_URL = "http://localhost:9000";

const TRACE_STORAGE_KEY = "span_explainer_trace";
const UPLOAD_ID_STORAGE_KEY = "span_explainer_upload_id";

let token: string | null = localStorage.getItem("token");

export async function register(): Promise<void> {
  const res = await fetch(`${BASE_URL}/register`, { method: "POST" });
  const data: RegisterResponse = await res.json();
  token = data.token;
  console.log({token})
  localStorage.setItem("token", token);
  localStorage.removeItem(TRACE_STORAGE_KEY);
  localStorage.removeItem(UPLOAD_ID_STORAGE_KEY);
}

function authHeaders(): HeadersInit {
  if (!token) {
    throw new Error("Missing auth token");
  }

  return {
    Authorization: `Bearer ${token}`,
    "Content-Type": "application/json",
  };
}

export async function uploadTrace(
  traceJson: string
): Promise<UploadTraceResponse> {
  const res = await fetch(`${BASE_URL}/upload-trace`, {
    method: "POST",
    headers: authHeaders(),
    body: traceJson,
  });

  return res.json();
}

export async function explainSpan(
  payload: ExplainSpanRequest,
  setLoading: React.Dispatch<React.SetStateAction<boolean>>

): Promise<ExplainSpanResponse> {
  setLoading(true)
  const res = await fetch(`${BASE_URL}/explain-span`, {
    method: "POST",
    headers: authHeaders(),
    body: JSON.stringify(payload),
  }).finally(()=>{setLoading(false)});

  return res.json();
}