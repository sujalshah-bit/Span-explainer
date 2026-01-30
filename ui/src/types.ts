export interface RegisterResponse {
  user_id: string;
  token: string;
}

export interface UploadTraceResponse {
  upload_id: string;
}

export interface ExplainSpanRequest {
  upload_id: string;
  span_id: string;
  question: string;
}

export interface ExplainSpanAnswer {
  root_cause: string;
  impact: string;
  suggested_action: string;
}

export interface ExplainSpanResponse {
  answer: ExplainSpanAnswer;
}
