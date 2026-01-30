package main

import (
	"log"
	"net/http"

	"github.com/sujalshah-bit/span-explainer/internal/llm"
	"github.com/sujalshah-bit/span-explainer/internal/server"
	"github.com/sujalshah-bit/span-explainer/internal/store"
)

func main() {
	store := store.NewStore()
	llm := &llm.PhiLLM{}
	httpServer := &server.Server{
		Store: store,
		Llm:   llm,
	}
	mux := &http.ServeMux{}

	server.RegisterRoutes(mux, httpServer, store)

	log.Println("🚀 server running on :8080")
	handler := server.CORS(mux)
	log.Fatal(http.ListenAndServe(":8080", handler))
}
