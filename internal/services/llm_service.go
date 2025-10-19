package services

import (
	"fmt"
	"log"
	"time"
)

type LLMService interface {
	Ask(text string) (string, error)
}

type LLM struct{}

func (l *LLM) Ask(text string) (string, error) {
	log.Printf("[LLM] Sending prompt to LLM model: %q", text)
	time.Sleep(800 * time.Millisecond)
	// Simulation
	response := fmt.Sprintf("Hello! You told me: %s. Do you want to learn something new today?", text)
	log.Printf("[LLM] Response received: %q\n", response)
	return response, nil
}
