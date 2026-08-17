package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

type Route struct {
	Prefix string
	Target *url.URL
}

type Gateway struct {
	routes []Route
}

func NewGateway() *Gateway {
	return &Gateway{
		routes: make([]Route, 0),
	}
}

func (g *Gateway) RegisterRoute(prefix string, rawTarget string) error {
	parsedURL, err := url.Parse(rawTarget)
	if err != nil {
		return err
	}

	g.routes = append(g.routes, Route{
		Prefix: prefix,
		Target: parsedURL,
	})
	return nil
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range g.routes {
		if strings.HasPrefix(r.URL.Path, route.Prefix) {
			proxy := httputil.NewSingleHostReverseProxy(route.Target)
			r.URL.Path = strings.TrimPrefix(r.URL.Path, route.Prefix)
			if !strings.HasPrefix(r.URL.Path, "/") {
				r.URL.Path = "/" + r.URL.Path
			}
			r.Host = route.Target.Host
			proxy.ServeHTTP(w, r)
			return
		}
	}

	http.Error(w, "Route Not Found", http.StatusNotFound)
}

func main() {
	gateway := NewGateway()

	if err := gateway.RegisterRoute("/api/v1/traffic", "http://localhost:8081"); err != nil {
		log.Fatalf("Failed to register traffic route: %v", err)
	}

	if err := gateway.RegisterRoute("/api/v1/transit", "http://localhost:8082"); err != nil {
		log.Fatalf("Failed to register transit route: %v", err)
	}

	server := &http.Server{
		Addr:         ":8080",
		Handler:      gateway,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Velo running on port :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failure: %v", err)
	}
}
