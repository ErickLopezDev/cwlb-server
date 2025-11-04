# Voice Assistant Project

This project implements a voice assistant using Google Cloud services for Speech-to-Text (STT), Large Language Model (LLM) via Gemini, and Text-to-Speech (TTS). It supports local audio recording and processing, with plans for MQTT-based distributed processing.

## Architecture

### Components

- **Local Recorder** (`internal/utils/local_recorder.go`): Handles audio recording from microphone, saves to WAV, and triggers processing.
- **Local Orchestrator** (`internal/core/local_orchestrator.go`): Coordinates STT → LLM → TTS pipeline.
- **Services**:
  - **GCP STT** (`internal/services/gcp_stt_service.go`): Converts audio to text using Google Speech-to-Text.
  - **Gemini LLM** (`internal/services/gemini_service.go`): Processes text prompts and generates responses.
  - **GCP TTS** (`internal/services/gcp_tts_service.go`): Converts text responses to audio and handles playback.

### Local Simulation Flow

1. **Recording**: User presses 'R' to start/stop recording via PortAudio.
2. **Audio Processing**: Raw PCM chunks are collected and saved as `local_record.wav`.
3. **STT**: WAV file is sent to Google STT (Spanish), returns transcribed text.
4. **LLM**: Text is sent to Gemini with system prompt, returns response text.
5. **TTS**: Response text is synthesized to audio (Spanish voice).
6. **Playback**: Audio is played via PortAudio, and saved as `local_response.wav` for debugging.

### Prerequisites

- Go 1.19+
- Google Cloud credentials (`n8n-testing-469619-979325bcaed2.json`)
- Environment variables:
  - `GEMINI_API_KEY`: Your Gemini API key
  - `PROMPT`: System prompt for the LLM (e.g., "You are a helpful assistant.")
- Dependencies: Install with `go mod tidy`

### Running Local Simulation

```bash
cd cmd/iot
go run main.go
```

- Press 'R' then Enter to start recording.
- Speak in Spanish.
- Press 'R' again to stop and process.
- Response audio plays automatically.

### Files Generated

- `local_record.wav`: Recorded audio.
- `local_response.wav`: Synthesized response.

## Future: MQTT Client Integration

To enable distributed processing (e.g., for IoT devices), integrate MQTT for audio chunk streaming.

### Proposed MQTT Flow

1. **Audio Encoding**: Split recorded audio into chunks (e.g., 1-second PCM segments), encode to base64 or binary.
2. **MQTT Publish**: Send chunks to MQTT topic (e.g., `audio/chunks`).
3. **Server Processing**: Remote server subscribes, reassembles audio, runs STT/LLM/TTS, splits response into chunks.
4. **MQTT Subscribe**: Client receives response chunks, decodes, reassembles, and plays.
5. **Chunk Management**: Use sequence numbers, timestamps, or unique IDs for ordering.

### Implementation Plan

#### 1. Add MQTT Client
- Use `github.com/eclipse/paho.mqtt.golang` for MQTT.
- Create `internal/mqtt/client.go` with connect, publish, subscribe methods.

#### 2. Modify Recorder for Chunking
- In `LocalRecorder`, instead of saving full audio, send chunks via MQTT.
- Encode chunks: Convert PCM bytes to base64 for text-based MQTT.

#### 3. Orchestrator Changes
- Add MQTT mode: If MQTT enabled, publish audio chunks instead of local processing.
- Subscribe to response topic, reassemble audio.

#### 4. Chunk Encoding/Decoding
- **Encoding**: Split audio into fixed-size buffers (e.g., 16000 samples/second), base64 encode.
- **Decoding**: Receive chunks, base64 decode, concatenate in order.
- Handle out-of-order with sequence IDs in MQTT payload.

#### 5. Configuration
- Add flags/env vars for MQTT broker URL, topics, chunk size.

#### 6. Testing
- Local MQTT broker (e.g., Mosquitto) for testing.
- Ensure low latency for real-time feel.

This setup allows scaling to multiple devices while keeping local fallback.</content>
<parameter name="filePath">README.md