# MQTT Audio Processing Flow

## Overview

This document describes the MQTT-based audio processing pipeline for the CWLB (Conversational Wireless Learning Bot) system. The system enables real-time voice interaction between ESP32 devices and a backend server using MQTT for communication, with speech-to-text (STT), language model (LLM), and text-to-speech (TTS) services.

## Architecture

- **ESP32 Device**: Captures audio via I2S, sends audio chunks via MQTT, receives and plays response audio.
- **MQTT Broker**: Mosquitto server handling message routing.
- **Backend Server**: Go application processing audio through STT → LLM → TTS pipeline.
- **Services**:
  - STT: Google Cloud Speech-to-Text (Spanish)
  - LLM: Google Gemini 2.5 Flash
  - TTS: Google Cloud Text-to-Speech (Spanish, female voice)

## MQTT Topics and Flow

### From ESP32 to Backend

1. **Start Session** (`/device/audio/start`)
   - Payload: JSON `{"robot_id": "device_id", "kid_id": "kid_id"}`
   - Initializes audio session for the device.

2. **Audio Chunks** (`/device/audio/chunk`)
   - Payload: Base64-encoded raw PCM audio data (16-bit, 16kHz, mono).
   - Chunks are accumulated in the backend until end signal.

3. **End Session** (`/device/audio/end`)
   - Payload: Any (currently "done" in tests).
   - Triggers processing of accumulated audio.

### Backend Processing Pipeline

Upon receiving `/device/audio/end`:

1. **Accumulate Audio**: Combine all chunks into full PCM data.
2. **Save as WAV**: Create temporary WAV file (internal/core/orchestrator.go:saveWAV).
3. **STT**: Convert audio to text using GCP Speech-to-Text (internal/services/gcp_stt_service.go).
4. **LLM**: Generate response using Gemini (internal/services/gemini_service.go).
5. **TTS**: Synthesize response to audio using GCP Text-to-Speech (internal/services/gcp_tts_service.go).
6. **Chunk Response**: Split TTS audio into chunks and send back.

### From Backend to ESP32

1. **Response Chunks** (`/device/{device_id}/audio/response_chunk`)
   - Payload: JSON `{"index": 0, "total": N, "data": "base64_encoded_pcm"}`
   - Each chunk is ~32KB (~1 second at 16kHz 16-bit).

2. **Response End** (`/device/{device_id}/audio/response_end`)
   - Payload: Empty string.
   - Signals end of response audio.

## Backend Implementation Details

### MQTT Client (internal/mqtt/client.go)

- Subscribes to `/device/audio/start`, `/device/audio/chunk`, `/device/audio/end`.
- Manages per-device audio sessions with mutex protection.
- On end: calls `orchestrator.HandleAudio()` and sends response chunks.

### Orchestrator (internal/core/orchestrator.go)

- `HandleAudio()`: Full pipeline STT → LLM → TTS.
- Currently echoes transcription for testing (LLM commented out).
- Uses temporary WAV files for STT input.

### Services

- **STT (GCP)**: Recognizes Spanish speech, expects 16kHz LINEAR16 PCM.
- **LLM (Gemini)**: Uses system prompt from env var `PROMPT`.
- **TTS (GCP)**: Generates Spanish female voice, 16kHz LINEAR16 PCM.

## ESP32 Implementation

The ESP32 should implement the following flow:

### Audio Capture and Transmission

1. **Initialize I2S**: Configure for 16kHz, 16-bit, mono PCM capture.
2. **Start Session**: Publish to `/device/audio/start` with device ID.
3. **Stream Audio Chunks**:
   - Capture audio in ~1-second buffers (~32KB).
   - Base64 encode each chunk.
   - Publish to `/device/audio/chunk`.
4. **End Session**: Publish to `/device/audio/end` when capture stops (e.g., button release or silence detection).

### Receive and Playback Response

1. **Subscribe to Response Topics**:
   - `/device/{device_id}/audio/response_chunk`
   - `/device/{device_id}/audio/response_end`

2. **Handle Response Chunks**:
   - Decode base64 to PCM data.
   - Accumulate chunks in order.
   - On `response_end`, play accumulated audio via I2S output.

3. **Playback**: Use I2S DAC or amplifier to play 16kHz 16-bit PCM.

### Key Considerations

- **Real-time Streaming**: Send chunks as captured, don't wait for full audio.
- **Chunk Size**: Match backend (32KB) for efficiency.
- **Error Handling**: Handle MQTT disconnections, resend failed chunks if needed.
- **Buffering**: Buffer incoming response chunks to avoid playback gaps.
- **Device ID**: Use unique robot_id for topic addressing.

### Sample ESP32 Pseudocode

```cpp
// MQTT Topics
#define TOPIC_START "/device/audio/start"
#define TOPIC_CHUNK "/device/audio/chunk"
#define TOPIC_END "/device/audio/end"
#define TOPIC_RESPONSE_CHUNK "/device/%s/audio/response_chunk"
#define TOPIC_RESPONSE_END "/device/%s/audio/response_end"

// Audio config
const int SAMPLE_RATE = 16000;
const int BITS_PER_SAMPLE = 16;
const int CHANNELS = 1;
const int CHUNK_SIZE = 32000; // ~1 sec

void setup() {
    // Init I2S for input/output
    // Init MQTT client
    mqtt.subscribe(TOPIC_RESPONSE_CHUNK, device_id);
    mqtt.subscribe(TOPIC_RESPONSE_END, device_id);
}

void captureAndSend() {
    // Start session
    mqtt.publish(TOPIC_START, "{\"robot_id\":\"esp32-01\",\"kid_id\":\"kid001\"}");

    while (recording) {
        uint8_t buffer[CHUNK_SIZE];
        i2s_read(buffer, CHUNK_SIZE);
        String encoded = base64::encode(buffer, CHUNK_SIZE);
        mqtt.publish(TOPIC_CHUNK, encoded);
    }

    mqtt.publish(TOPIC_END, "done");
}

void onResponseChunk(String payload) {
    // Parse JSON, decode base64, accumulate
    // On end, play audio
}
```

## MQTT Configuration

- **Broker**: Mosquitto with anonymous access on ports 1883 (MQTT) and 9001 (WebSockets).
- **Client ID**: "cwlb-mqtt-server" for backend.
- **QoS**: 0 (at most once) for all messages.
- **Test Script**: `test_mqtt.sh` simulates ESP32 by sending WAV file chunks and collecting response.

## Testing

Use `test_mqtt.sh` to test the pipeline:

```bash
./test_mqtt.sh test.wav localhost 1883
```

This converts WAV to chunks, sends via MQTT, and saves response as `esp32Output.pcm`.