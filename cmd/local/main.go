package main

import (
	"log"
	"os"

	"github.com/ErickLopezDev/cwlb-server/internal/core"
	"github.com/ErickLopezDev/cwlb-server/internal/services"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env")

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "n8n-testing-469619-979325bcaed2.json")

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY not set")
	}

	log.Println("Initializing GCP STT...")
	stt, err := services.NewGCPSTT()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Initializing GCP TTS...")
	tts, err := services.NewGCPTTS()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Initializing Gemini LLM...")
	llm, err := services.NewGeminiLLM(apiKey)
	if err != nil {
		log.Fatal(err)
	}

	localOrchestrator := core.NewLocalOrchestrator(stt, llm, tts)

	log.Println("Processing local_record.wav...")
	err = localOrchestrator.ProcessLocalRecord()
	if err != nil {
		log.Fatal("Failed to process local record:", err)
	}

	log.Println("Local processing completed.")
}
