package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"graffiti-2/cache"
	"graffiti-2/handlers"

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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	client, err := firestore.NewClient(ctx, project)
	if err != nil {
		log.Fatal(err)
	}

	var tc cache.TemplateCache
	tc.Listen(ctx, client) // Wait until cache gets loaded.

	http.HandleFunc("/test_endpoint_1", handlers.TestEndpoint1Handler(&tc))
	http.HandleFunc("/test_endpoint_2", handlers.TestEndpoint2Handler(&tc))
	http.ListenAndServe(":"+port, nil)

	<-ctx.Done() // Allow closing of the server when context is done.
}
