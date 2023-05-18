package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/rakyll/openai-go"
	"github.com/rakyll/openai-go/chat"
)

const (
	openAiKey = "OPENAI_API_KEY"
)

func proxyLLMRequest(w http.ResponseWriter, r *http.Request) {
	log.Println("ok doing it")

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

	completeParams := apiReq.CreateCompletionParams

	if !apiReq.Stream {
		client := chat.NewClient(s, apiReq.Model)
		apiResp, err := client.CreateCompletion(ctx, completeParams)
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

	} else {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Write an SSE comment to establish the SSE stream
		fmt.Fprint(w, ": SSE stream\n")

		// Flush the response writer to ensure the comment is sent
		// immediately
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported",
				http.StatusInternalServerError)
			return
		}
		flusher.Flush()

		client := chat.NewStreamingClient(s, apiReq.Model)

		var wg sync.WaitGroup
		wg.Add(1)

		client.CreateCompletion(ctx, completeParams, func(r *chat.CreateCompletionStreamingResponse) {
			// Encode the data as JSON
			jsonData, err := json.Marshal(r)
			if err != nil {
				http.Error(w, "Failed to encode JSON",
					http.StatusInternalServerError)
				return
			}

			// Write the JSON chunk to the SSE stream
			fmt.Fprintf(w, "data: %s\n\n", jsonData)
			flusher.Flush()

			// Delay between sending chunks (optional)
			time.Sleep(1 * time.Second)
		})
	}
}
