package main

import (
	"log"

	"github.com/nicolasbonnici/gorest"
	"github.com/nicolasbonnici/gorest/pluginloader"

	ai "github.com/nicolasbonnici/gorest-ai"
)

func init() {
	// Register the AI plugin
	pluginloader.RegisterPluginFactory("ai", ai.NewPlugin)
}

func main() {
	// Configure GoREST
	cfg := gorest.Config{
		ConfigPath: ".",
	}

	log.Println("Starting GoREST with AI plugin...")
	log.Println("===========================================")
	log.Println("")
	log.Println("AI Chat endpoints:")
	log.Println("  POST /api/ai/chat         - Send chat completion request")
	log.Println("  POST /api/ai/chat/stream  - Send streaming chat request")
	log.Println("")
	log.Println("Provider Management (Admin):")
	log.Println("  POST   /api/ai/providers     - Create provider")
	log.Println("  GET    /api/ai/providers/:id - Get provider by ID")
	log.Println("  GET    /api/ai/providers     - List all providers")
	log.Println("  PUT    /api/ai/providers/:id - Update provider")
	log.Println("  DELETE /api/ai/providers/:id - Delete provider")
	log.Println("")
	log.Println("Usage & Statistics:")
	log.Println("  GET /api/ai/usage       - Get usage statistics")
	log.Println("  GET /api/ai/usage/quota - Get user quota status")
	log.Println("")
	log.Println("Request History:")
	log.Println("  GET /api/ai/requests     - List request history")
	log.Println("  GET /api/ai/requests/:id - Get request details")
	log.Println("")
	log.Println("===========================================")
	log.Println("")

	// Start the server
	gorest.Start(cfg)
}
