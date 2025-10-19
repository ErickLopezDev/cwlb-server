package mqtt

import (
	"encoding/base64"
	"github.com/ErickLopezDev/cwlb-server/internal/core"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"sync"
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

		// start
		c.Subscribe("/device/audio/start", 0, func(_ mqtt.Client, msg mqtt.Message) {
			deviceID := msg.Topic() // could be parsed for multiple robots
			sessionsMu.Lock()
			sessions[deviceID] = &AudioSession{}
			sessionsMu.Unlock()
			log.Println("Recording started for", deviceID)
		})

		// chunk
		c.Subscribe("/device/audio/chunk", 0, func(_ mqtt.Client, msg mqtt.Message) {
			data, err := base64.StdEncoding.DecodeString(string(msg.Payload()))
			if err != nil {
				log.Println("Error decoding chunk:", err)
				return
			}

			sessionsMu.Lock()
			session, ok := sessions["/device/audio/start"]
			sessionsMu.Unlock()
			if !ok {
				log.Println("No active session, ignoring chunk")
				return
			}

			session.mu.Lock()
			session.chunks = append(session.chunks, data)
			session.mu.Unlock()
			log.Println("Chunk received, size:", len(data))
		})

		// end
		c.Subscribe("/device/audio/end", 0, func(client mqtt.Client, msg mqtt.Message) {
			log.Println("Recording finished. Processing audio...")

			sessionsMu.Lock()
			session := sessions["/device/audio/start"]
			delete(sessions, "/device/audio/start")
			sessionsMu.Unlock()

			if session == nil {
				log.Println("Error: no active session.")
				return
			}

			// Combine all chunks
			fullAudio := []byte{}
			for _, chunk := range session.chunks {
				fullAudio = append(fullAudio, chunk...)
			}

			// Process audio
			audioResponse, err := orchestrator.HandleAudio(fullAudio)
			if err != nil {
				log.Println("Error processing audio:", err)
				return
			}

			// Publish response
			client.Publish("/device/audio/output", 0, false, audioResponse)
			log.Println("Response published to the robot.")
		})
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	return client
}

