package main

import (
	"github.com/mkj-gram/go_email_service/internal/emailprovider"
	"github.com/mkj-gram/go_email_service/internal/emailsender"
	"github.com/mkj-gram/go_email_service/internal/sendgrid"
	"github.com/mkj-gram/go_email_service/internal/server"
	"github.com/mkj-gram/go_email_service/internal/sparkpost"
	"log"
	"net/http"
	"os"
)

const LOG_FILE = "log"

func main() {
	// Set up log to print to a file
	f, err := os.OpenFile(LOG_FILE, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	// Create providers
	providers := []emailprovider.Provider{
		sparkpost.SparkPostProvider{},
		sendgrid.SendGridProvider{},
	}
	for _, p := range providers {
		if err := p.Init(); err != nil {
			log.Println(err)
		}
	}
	// Set the strategy to be used
	strategy := emailsender.RoundRobinSender{Providers: providers}
	// Start the web server
	app := server.ServerApp{
		Strategy: &strategy,
		LogFile:  LOG_FILE,
	}
	app.Serve()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Started and listening on port %s\n", port)
	http.ListenAndServe(":"+port, nil)
}
