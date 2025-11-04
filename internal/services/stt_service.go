package services

import (
// "log"
// "time"
)

type STTService interface {
	ConvertAudio(filePath string) (string, error)
}

// type STT struct{}
//
// func (s *STT) ConvertAudio(filePath string) (string, error) {
// 	log.Printf("[STT] Sending audio '%s' to remote STT service...", filePath)
// 	time.Sleep(500 * time.Millisecond)
// 	// Simulation
// 	text := "Hi, I'm Ana"
// 	log.Printf("[STT] Transcription received: %q\n", text)
// 	return text, nil
// }
