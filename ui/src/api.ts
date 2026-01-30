import type{
  RegisterResponse,
  UploadTraceResponse,
  ExplainSpanRequest,
  ExplainSpanResponse,
} from "./types";

const BASE_URL = "http://localhost:8080";

let token: string | null = localStorage.getItem("token");

export async function register(): Promise<void> {
  const res = await fetch(`${BASE_URL}/register`, { method: "POST" });
  const data: RegisterResponse = await res.json();
  token = data.token;
  console.log({token})
  localStorage.setItem("token", token);
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
  payload: ExplainSpanRequest
): Promise<ExplainSpanResponse> {
  const res = await fetch(`${BASE_URL}/explain-span`, {
    method: "POST",
    headers: authHeaders(),
    body: JSON.stringify(payload),
  });

  return res.json();
}