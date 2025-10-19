package main

import (
	"fmt"
	"log"

	"github.com/ErickLopezDev/cwlb-server/internal/core"
	"github.com/ErickLopezDev/cwlb-server/internal/mqtt"
	"github.com/ErickLopezDev/cwlb-server/internal/services"
)

func main() {
	orchestrator := &core.Orchestrator{
		STT: &services.STT{},
		LLM: &services.LLM{},
		TTS: &services.TTS{},
	}

	client := mqtt.NewClient("tcp://localhost:1883", orchestrator)
	log.Println("Backend MQTT running...")
	fmt.Println(client)
	select {}
}
