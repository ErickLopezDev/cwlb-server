package services

import (
	"log"
	"time"
)

type TTSService interface {
	Synthesize(text string) ([]byte, error)
}

type TTS struct{}

func (t *TTS) Synthesize(text string) ([]byte, error) {
	log.Printf("[TTS] Sending text to TTS service: %q", text)
	time.Sleep(400 * time.Millisecond)
	// Simulation
	audio := []byte("SIMULATED_BINARY_AUDIO")
	log.Printf("[TTS] Synthesized response (length %d bytes)\n", len(audio))
	return audio, nil
}
