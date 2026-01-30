package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sujalshah-bit/span-explainer/internal/llm"
	otelvalidator "github.com/sujalshah-bit/span-explainer/internal/otel_validator"
	"github.com/sujalshah-bit/span-explainer/internal/store"
)

const (
	REGISTER_ENDPOINT     = "/register"
	UPLOAD_TRACE_ENDPOINT = "/upload-trace"
	QUERY_LLM_ENDPOINT    = "/explain-span"
)

type Server struct {
	Store *store.Store
	Llm   llm.LLM
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RegisterRoutes(mux *http.ServeMux, server *Server, store *store.Store) {
	// Public
	mux.HandleFunc(REGISTER_ENDPOINT, server.register)

	// Protected
	mux.Handle(UPLOAD_TRACE_ENDPOINT,
		Auth(store, http.HandlerFunc(server.uploadTrace)),
	)
	mux.Handle(QUERY_LLM_ENDPOINT,
		Auth(store, http.HandlerFunc(server.queryLLM)),
	)
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	user := s.Store.RegisterUser()
	sujal := map[string]string{
		"user_id": user.ID,
		"token":   user.Token,
	}
	fmt.Println(sujal)
	json.NewEncoder(w).Encode(map[string]string{
		"user_id": user.ID,
		"token":   user.Token,
	})
}

func (s *Server) uploadTrace(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	var raw json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := otelvalidator.ValidateOTLPTraces(raw); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	trace := &store.StoredTrace{
		ID:        uuid.NewString(),
		RawJSON:   raw,
		CreatedAt: time.Now(),
	}

	s.Store.AddTrace(userID, trace)

	json.NewEncoder(w).Encode(map[string]string{
		"upload_id": trace.ID,
	})
}

func (s *Server) queryLLM(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	var req struct {
		TraceID  string `json:"upload_id"`
		SpanID   string `json:"span_id"`
		Question string `json:"question"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.TraceID == "" || req.SpanID == "" || req.Question == "" {
		http.Error(
			w,
			"upload_id, span_id, and question are required",
			http.StatusBadRequest,
		)
		return
	}

	trace, ok := s.Store.GetTrace(userID, req.TraceID)
	if !ok {
		http.Error(w, "trace not found", http.StatusNotFound)
		return
	}

	answer, err := s.Llm.ExplainTrace(
		r.Context(),
		trace.RawJSON,
		req.Question,
		req.SpanID,
	)
	if err != nil {
		fmt.Println("ERROR:", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]*llm.ExplainResult{
		"answer": answer,
	})
}
func Auth(store *store.Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "missing auth", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(auth, "Bearer ")
		userID, ok := store.UserFromToken(token)
		if !ok {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		r.Header.Set("X-User-ID", userID)
		next.ServeHTTP(w, r)
	})
}
