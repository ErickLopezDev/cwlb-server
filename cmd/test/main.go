package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ErickLopezDev/cwlb-server/internal/services"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env")

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("missing GEMINI_API_KEY environment variable")
	}

	llm, err := services.NewGeminiLLM(apiKey)
	if err != nil {
		log.Fatalf("failed to initialize GeminiLLM: %v", err)
	}

	var model services.LLMService = llm

	response, err := model.Ask("por que la luna me sigue?")
	if err != nil {
		log.Fatalf("error calling Gemini: %v", err)
	}

	fmt.Println(response)
}
