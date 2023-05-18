package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("RUNNING LLM PROXY")

	http.HandleFunc("/", proxyLLMRequest)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
