package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yaq-cc/graffiti/cache"
	"github.com/yaq-cc/graffiti/handlers"

	"cloud.google.com/go/firestore"
)

func main() {
	project := os.Getenv("PROJECT_ID")
	if project == "" {
		project = "holy-diver-297719"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	parent := context.Background()
	notify, stop := signal.NotifyContext(parent, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	client, err := firestore.NewClient(parent, project)
	if err != nil {
		log.Fatal(err)
	}

	var tc cache.TemplateCache
	// Terminate tc.Listen on signal notification.
	tc.Listen(notify, client) // Wait until cache gets loaded.
	// ~~ HTTP Server ~~ //
	mux := http.NewServeMux()
	mux.HandleFunc("/endpoints", handlers.GetAllHandler(&tc))
	mux.HandleFunc("/test_endpoint_1", handlers.TestEndpoint1Handler(&tc))
	mux.HandleFunc("/test_endpoint_2", handlers.TestEndpoint2Handler(&tc))
	mux.HandleFunc("/get_definitions", handlers.GetAllHandler(&tc))
	mux.HandleFunc("/update_definitions", handlers.UpdateAllHandler(client))
	server := &http.Server{
		Addr:        ":" + port,
		Handler:     mux,
		BaseContext: func(net.Listener) context.Context { return parent },
	}
	go server.ListenAndServe()
	<-notify.Done() // Allow closing of the server when context is done.
	shutCtx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	server.Shutdown(shutCtx)
}
