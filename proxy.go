package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/rakyll/openai-go"
	"github.com/rakyll/openai-go/chat"
)

const (
	openAiKey = "OPENAI_API_KEY"
)

func proxyLLMRequest(w http.ResponseWriter, r *http.Request) {
	apiReq, err := UmarshalJsonReq(r.Body)
	if err != nil {
		log.Println("unable to parse req: %v", err)

		http.Error(w, "Failed to read request body",
			http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	log.Println("Proxying request: %v", spew.Sdump(apiReq))

	ctx := context.Background()
	s := openai.NewSession(os.Getenv(openAiKey))

	client := chat.NewClient(s, apiReq.Model)
	apiResp, err := client.CreateCompletion(ctx, apiReq.CreateCompletionParams)
	if err != nil {
		log.Fatalf("Failed to complete: %v", err)
	}

	log.Println("proxy resp: ", spew.Sdump(apiResp))

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(apiResp)
	if err != nil {
		http.Error(w, "Failed to encode response",
			http.StatusInternalServerError)
		return
	}
}
