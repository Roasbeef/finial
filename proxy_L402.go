package finial

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/btcsuite/btcd/btcutil"
)

// Pricer...
type Pricer interface {
	// GetPrice...
	GetPrice(r *http.Request) btcutil.Amount
}

type ProxyDest struct {
	// Host...
	Host string

	// APIKey...
	APIKey string
}

// ProxyCfg...
type ProxyCfg struct {
	// RequestPricer...
	RequestPricer Pricer

	// ListenAddr...
	ListenAddr string
	string
	// BackendDest...
	BackendDest *ProxyDest
}

// L402AIProxy...
type L402AIProxy struct {
	cfg *ProxyCfg

	// httpServer...
	httpServer *http.Server
}

// TODO(roasbeef):

// NewL402AIProxy...
func NewL402AIProxy(cfg *ProxyCfg) *L402AIProxy {
	proxy := &L402AIProxy{
		cfg: cfg,
	}

	proxy.httpServer = &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: http.HandlerFunc(proxy.proxyL402),
	}

	return proxy
}

// proxyL402...
func (l *L402AIProxy) proxyL402(w http.ResponseWriter, r *http.Request) {
	// First, swap out the destination of the request with the target LLM
	// backend.
	r.URL.Host = l.cfg.BackendDest.Host
	r.URL.Scheme = "https"
	r.Host = l.cfg.BackendDest.Host

	// If the API key for this backend is present, then we'll apply the
	// proper headers for the request here to ensure we can proxy properly.
	apiKey := l.cfg.BackendDest.APIKey
	if apiKey != "" {
		r.Header.Set("Authorization", "Bearer "+apiKey)
	}

	// With the request swapped, and our API key information attached,
	// we'll now proxy to true LLM backend.
	proxy := http.DefaultTransport
	resp, err := proxy.RoundTrip(r)
	if err != nil {
		fmt.Println("noope: ", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// With the response obtained, we'll now copy over the headers, writes
	// the status code, then copy the body. We use `io.Copy` here as it
	// works properly with steaming responses.
	for key, value := range resp.Header {
		for _, v := range value {
			w.Header().Add(key, v)
		}
	}

	w.WriteHeader(resp.StatusCode)

	io.Copy(w, resp.Body)
}

// Start...
func (l *L402AIProxy) Start() error {
	listener, err := net.Listen("tcp", l.cfg.ListenAddr)
	if err != nil {
		return nil
	}

	go func() {
		err := l.httpServer.Serve(listener)
		if err != http.ErrServerClosed {
			panic(err)
		}
	}()

	return nil
}

// Stop...
func (l *L402AIProxy) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return l.httpServer.Shutdown(ctx)
}

// withCancelOnSignal returns a copy of parent with a new Done channel. The
// returned context's Done channel is closed when the parent context's Done
// channel is closed, or when one of the signals is received.
func withCancelOnSignal(parent context.Context,
	signals ...os.Signal) context.Context {

	ctx, cancel := context.WithCancel(parent)

	c := make(chan os.Signal, 1)
	signal.Notify(c, signals...)

	go func() {
		select {
		case <-c:
			cancel()

		case <-parent.Done():
			// Parent context was cancelled, stop listening for
			// signals.
		}
	}()

	return ctx
}

// WaitForShutdown...
func (l *L402AIProxy) WaitForShutdown() {
	ctx := withCancelOnSignal(
		context.Background(), os.Interrupt, syscall.SIGTERM,
	)

	<-ctx.Done()

	l.Stop()
}
