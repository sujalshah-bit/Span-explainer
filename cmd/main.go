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
	llm, err := llm.NewPhiLLM()
	if err != nil {
		log.Fatal(err)
	}
	httpServer := &server.Server{
		Store: store,
		Llm:   llm,
	}
	mux := &http.ServeMux{}

	server.RegisterRoutes(mux, httpServer, store)

	log.Println("server running on http://localhost:9000")
	handler := server.CORS(mux)
	log.Fatal(http.ListenAndServe(":9000", handler))
}
