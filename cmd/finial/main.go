package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/roasbeef/finial"
)

const (
	openAiKey = "OPENAI_API_KEY"
)

var (
	listenAddr = flag.String("listenaddr", ":9090", "the address to "+
		"listen on")

	proxyDest = flag.String("proxydest", "api.openai.com", "the address "+
		"to proxy requests to")
)

func main() {
	flag.Parse()

	apiKey := os.Getenv(openAiKey)
	llmProxy := finial.NewL402AIProxy(&finial.ProxyCfg{
		ListenAddr: *listenAddr,
		BackendDest: &finial.ProxyDest{
			Host:   *proxyDest,
			APIKey: apiKey,
		},
	})

	fmt.Println(apiKey)

	log.Printf("Starting LLM API Proxy, target=%v...", *proxyDest)

	if err := llmProxy.Start(); err != nil {
		fmt.Println("unable to start proxy: ", err)
		return
	}

	llmProxy.WaitForShutdown()
}
