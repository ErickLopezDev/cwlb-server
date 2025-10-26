package mqtt

import (
	"encoding/base64"
	"log"
	"strings"
	"sync"

	"github.com/ErickLopezDev/cwlb-server/internal/core"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type AudioSession struct {
	mu     sync.Mutex
	chunks [][]byte
}

func NewClient(broker string, orchestrator *core.Orchestrator) mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID("cwlb-mqtt-server")

	sessions := make(map[string]*AudioSession)
	var sessionsMu sync.Mutex

	opts.OnConnect = func(c mqtt.Client) {
		log.Println("Connected to MQTT broker")

		c.Subscribe("/device/+/audio/start", 0,
			func(_ mqtt.Client, msg mqtt.Message) {
				handleStart(&sessionsMu, sessions, msg)
			},
		)

		c.Subscribe("/device/+/audio/chunk", 0,
			func(_ mqtt.Client, msg mqtt.Message) {
				handleChunk(&sessionsMu, sessions, msg)
			},
		)

		c.Subscribe("/device/+/audio/end", 0,
			func(client mqtt.Client, msg mqtt.Message) {
				handleEnd(&sessionsMu, sessions, client, orchestrator, msg)
			},
		)
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	return client
}

// Starts a new audio session
func handleStart(mu *sync.Mutex, sessions map[string]*AudioSession, msg mqtt.Message) {
	deviceID := extractDeviceID(msg.Topic())
	mu.Lock()
	sessions[deviceID] = &AudioSession{}
	mu.Unlock()
	log.Println("Recording started for", deviceID)
}

// Saves a chunk of audio
func handleChunk(mu *sync.Mutex, sessions map[string]*AudioSession, msg mqtt.Message) {
	deviceID := extractDeviceID(msg.Topic())

	data, err := base64.StdEncoding.DecodeString(string(msg.Payload()))
	if err != nil {
		log.Println("Error decoding chunk:", err)
		return
	}

	mu.Lock()
	session, ok := sessions[deviceID]
	mu.Unlock()
	if !ok {
		log.Println("No active session for", deviceID)
		return
	}

	session.mu.Lock()
	session.chunks = append(session.chunks, data)
	session.mu.Unlock()

	log.Println("Chunk received for", deviceID, "size:", len(data))
}

// Combines and process the audio, then publish the resposne
func handleEnd(mu *sync.Mutex, sessions map[string]*AudioSession, client mqtt.Client, orchestrator *core.Orchestrator, msg mqtt.Message) {
	deviceID := extractDeviceID(msg.Topic())

	mu.Lock()
	session := sessions[deviceID]
	delete(sessions, deviceID)
	mu.Unlock()

	if session == nil {
		log.Println("Error: no active session for", deviceID)
		return
	}

	fullAudio := []byte{}
	for _, chunk := range session.chunks {
		fullAudio = append(fullAudio, chunk...)
	}

	audioResponse, err := orchestrator.HandleAudio(fullAudio)
	if err != nil {
		log.Println("Error processing audio:", err)
		return
	}

	client.Publish("/device/"+deviceID+"/audio/output", 0, false, audioResponse)
	log.Println("Response published to", deviceID)
}

func extractDeviceID(topic string) string {
	// example: /device/crowbot-01/audio/chunk
	parts := strings.Split(topic, "/")
	if len(parts) >= 3 {
		return parts[2] // "crowbot-01"
	}
	return "unknown"
}
