# V1Claw

**Your 24/7 Personal AI Assistant — Like Jarvis, but open source.**

V1Claw is a self-hosted AI assistant that runs on your Mac, Linux PC, Windows machine, or Android phone via Termux. Connect any LLM (Claude, GPT, Gemini, or local models), talk to it through voice or text, and let it control your device — read files, run commands, browse the web, take photos, send messages, and more.

One binary. No cloud dependency. Your data stays on your machine.

---

## Features

### 🧠 13 LLM Providers
Connect to any AI model — paid APIs or self-hosted:

| Provider | Type | Models |
|----------|------|--------|
| **Anthropic** | Cloud API | Claude 4, Claude 3.5, etc. |
| **OpenAI** | Cloud API | GPT-5, GPT-4, etc. |
| **Google Gemini** | Cloud API | Gemini 2.x, etc. |
| **Groq** | Cloud API | LLaMA, Mixtral (fast inference) |
| **OpenRouter** | Cloud API | 100+ models via single API |
| **DeepSeek** | Cloud API | DeepSeek V3, Coder |
| **NVIDIA** | Cloud API | NIM models |
| **Moonshot** | Cloud API | Kimi |
| **Zhipu** | Cloud API | GLM-4 |
| **Ollama** | Local | Any GGUF model on your machine |
| **vLLM** | Local | Self-hosted inference server |
| **GitHub Copilot** | Cloud API | Via Copilot subscription |
| **Any OpenAI-compatible API** | REST API | Custom endpoints |

### 🎤 Voice I/O Pipeline
Talk to V1Claw like Jarvis:
- **Microphone recording** — continuous listening with configurable backends
- **Speech-to-Text** — powered by Groq Whisper (fast, accurate)
- **Text-to-Speech** — OpenAI TTS or Edge TTS (free)
- **Wake word detection** — "Hey V1Claw" or custom phrases
- **Push-to-talk mode** — manual recording trigger
- Works on **desktop** (arecord/ffplay) and **Android** (termux-microphone-record/termux-media-player)

### 📷 Vision & Camera
- Capture photos from camera (desktop or Termux)
- Analyze images with vision-capable LLMs
- Screenshot capture and OCR
- Object detection support

### 📱 10 Communication Channels
Run V1Claw as a bot on any platform:

| Channel | Protocol |
|---------|----------|
| Telegram | Bot API |
| Discord | Bot gateway |
| Slack | Socket Mode |
| WhatsApp | WebSocket bridge |
| LINE | Messaging API |
| DingTalk | Stream SDK |
| Feishu/Lark | Event subscription |
| QQ | Official bot API |
| OneBot | WebSocket (v11) |
| MaixCam | TCP (IoT devices) |

### 🛠️ 14 Built-in Tools
The AI can use these tools autonomously:

| Tool | Description |
|------|-------------|
| `read_file` | Read file contents |
| `write_file` | Create/overwrite files |
| `edit_file` | Surgical text replacements |
| `append_file` | Append to files |
| `list_dir` | Browse directories |
| `exec` | Run shell commands |
| `web_search` | Search the web (Brave/Perplexity) |
| `web_fetch` | Fetch and read web pages |
| `message` | Send messages to channels |
| `cron` | Schedule recurring tasks |
| `subagent` | Spawn sub-agents for complex tasks |
| `spawn` | Run async background tasks |
| `spi` | SPI device communication (Linux) |
| `i2c` | I2C device communication (Linux) |

### 📱 Android/Termux Integration
Full hardware access on Android via Termux:API:
- 🎤 Microphone recording
- 📷 Camera capture
- 🔊 Text-to-speech
- 📍 GPS location
- 📋 Clipboard read/write
- 💡 Flashlight control
- 📳 Vibration
- 📱 SMS send/receive
- 📞 Phone calls
- 🔋 Battery status
- 📶 WiFi info
- 🔔 Notifications
- 🔊 Volume control

### 🔒 Security
- **Deny-by-default permissions** — camera, microphone, SMS, phone, location, clipboard, sensors each require explicit opt-in
- **Permissions freeze after startup** — the AI cannot grant itself new permissions at runtime
- **Shell command filtering** — 26 deny patterns block dangerous commands (reverse shells, rc file modification, encoding tricks)
- **SSRF protection** — web fetch blocks localhost, private networks, cloud metadata endpoints
- **Path traversal protection** — skill install/load validates against directory escape
- **API authentication** — configurable API key with constant-time comparison
- **Hardened file permissions** — all config/state files use 0600/0700
- **Non-root Docker** — container runs as unprivileged user
- **Content boundary markers** — mitigates prompt injection from user-provided files

### ⏰ Proactive & Scheduled Tasks
- **Cron scheduling** — schedule recurring tasks (reminders, checks, reports)
- **Proactive engine** — AI can initiate actions based on context
- **Heartbeat monitoring** — system health checks

### 🧩 Skills System
Extend V1Claw with installable skills:
- Built-in skills included
- Install community skills from URLs
- Skills are sandboxed markdown agent configurations

---

## Quick Start

### 1. Build

```bash
git clone https://github.com/amit-vikramaditya/V1Claw.git
cd V1Claw
make build
```

The binary will be at `build/v1claw-<os>-<arch>` (e.g., `build/v1claw-darwin-arm64`).

### 2. Configure

```bash
./build/v1claw-* onboard
```

This creates `~/.v1claw/config.json`. Edit it to add your API key:

```json
{
  "agents": [
    {
      "name": "v1claw",
      "model": "claude-sonnet-4-20250514"
    }
  ],
  "providers": [
    {
      "kind": "anthropic",
      "api_key": "sk-ant-your-key-here"
    }
  ]
}
```

### 3. Run

```bash
# Interactive CLI
./build/v1claw-* agent

# One-shot query
./build/v1claw-* agent "What's the weather in Tokyo?"

# Start as a 24/7 daemon (Telegram, Discord, etc.)
./build/v1claw-* gateway
```

---

## Deployment

### Standalone Binary (Recommended)

Works on any machine — no dependencies, no Docker needed.

```bash
# Build for your platform
make build

# Or cross-compile for another platform
GOOS=linux GOARCH=amd64 make build    # Linux x86_64
GOOS=linux GOARCH=arm64 make build    # Linux ARM64 / Android
GOOS=darwin GOARCH=arm64 make build   # macOS Apple Silicon
GOOS=windows GOARCH=amd64 make build  # Windows

# Install to ~/.local/bin
make install
```

### Docker

```bash
# Copy and edit the example config
cp config/config.example.json config/config.json

# Run as a 24/7 gateway
docker compose --profile gateway up -d

# Or run a one-shot query
docker compose run --rm v1claw-agent -m "Hello V1Claw"
```

### Android (Termux)

1. **Install Termux** from [F-Droid](https://f-droid.org/en/packages/com.termux/) (not Play Store)

2. **Install Termux:API** from F-Droid (for mic, camera, GPS, etc.)

3. **Set up Termux:**
   ```bash
   pkg update && pkg install termux-api golang git make
   ```

4. **Build V1Claw:**
   ```bash
   git clone https://github.com/amit-vikramaditya/V1Claw.git
   cd V1Claw
   make build
   ```

   Or cross-compile on your PC and transfer:
   ```bash
   GOOS=linux GOARCH=arm64 make build
   # Transfer build/v1claw-linux-arm64 to your phone
   ```

5. **Configure and run:**
   ```bash
   ./build/v1claw-linux-arm64 onboard
   # Edit ~/.v1claw/config.json — add API key and enable permissions
   ./build/v1claw-linux-arm64 gateway
   ```

6. **Enable hardware features** in config:
   ```json
   {
     "permissions": {
       "microphone": true,
       "camera": true,
       "location": true,
       "sms": false,
       "phone_calls": false,
       "clipboard": true,
       "sensors": true,
       "shell_hardware": true,
       "notifications": true
     },
     "voice": {
       "enabled": true,
       "mode": "wake-word",
       "wake_phrases": ["hey v1claw", "hey jarvis"],
       "tts_provider": "edge",
       "recorder_backend": "termux",
       "player_backend": "termux"
     }
   }
   ```

### Linux Server (24/7 Daemon)

```bash
# Build
make build

# Create a systemd service
sudo tee /etc/systemd/system/v1claw.service << 'EOF'
[Unit]
Description=V1Claw AI Assistant
After=network.target

[Service]
Type=simple
User=v1claw
ExecStart=/home/v1claw/.local/bin/v1claw gateway
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable --now v1claw
```

### Multi-Device via Tailscale

If you have a Tailscale mesh network, run V1Claw on multiple devices:

```bash
# On your main server (the "brain"):
# Edit config to bind to Tailscale interface
# Set gateway.host to "0.0.0.0" (safe behind Tailscale)
./v1claw gateway

# From any other device on your Tailscale network:
curl http://your-server.tail1234.ts.net:18790/api/v1/chat \
  -H "Authorization: Bearer your-api-key" \
  -d '{"message": "Hello from my phone"}'
```

---

## Configuration Reference

See [`config/config.example.json`](config/config.example.json) for the full configuration schema.

### Key Sections

| Section | Purpose |
|---------|---------|
| `agents` | Agent name, model, temperature, max tokens, workspace |
| `providers` | LLM provider credentials and endpoints |
| `channels` | Telegram/Discord/Slack/etc. bot tokens |
| `tools` | Web search API keys (Brave, Perplexity) |
| `voice` | Voice I/O settings, TTS provider, wake phrases |
| `permissions` | Per-feature hardware access toggles |
| `gateway` | Host, port, API key |
| `heartbeat` | System health monitoring interval |

### Permissions

Each hardware feature can be individually enabled or disabled. All are **off by default**:

| Permission | What It Controls |
|------------|-----------------|
| `camera` | Photo capture, vision analysis |
| `microphone` | Audio recording, voice input |
| `sms` | Send/receive text messages |
| `phone_calls` | Initiate phone calls |
| `location` | GPS coordinates |
| `clipboard` | Read/write clipboard |
| `sensors` | Device sensors (accelerometer, etc.) |
| `shell_hardware` | Hardware shell commands (Termux) |
| `notifications` | System notifications |
| `screen` | Screenshot capture |

Permissions are **frozen after startup** — the AI cannot escalate its own access at runtime. Change permissions by editing your config and restarting.

---

## CLI Commands

```
v1claw onboard              # First-time setup
v1claw agent                # Interactive chat
v1claw agent "your query"   # One-shot query
v1claw gateway              # Start 24/7 daemon
v1claw auth login           # Authenticate
v1claw auth status          # Check auth status
v1claw status               # Show system status
v1claw cron                 # Manage scheduled tasks
v1claw skills list          # List installed skills
v1claw skills install <url> # Install a skill
v1claw version              # Show version
```

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        V1Claw                               │
│                                                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────────┐  │
│  │ Channels │  │  Voice   │  │  Vision  │  │   Tools    │  │
│  │----------│  │----------│  │----------│  │------------│  │
│  │ Telegram │  │ Mic→STT  │  │ Camera   │  │ Files      │  │
│  │ Discord  │  │ TTS→Spk  │  │ Screen   │  │ Shell      │  │
│  │ Slack    │  │ Wake Word│  │ OCR      │  │ Web        │  │
│  │ WhatsApp │  │          │  │          │  │ Cron       │  │
│  │ LINE     │  │          │  │          │  │ Subagent   │  │
│  │ 5 more.. │  │          │  │          │  │ I2C/SPI    │  │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └─────┬──────┘  │
│       │              │              │              │         │
│       └──────────────┴──────┬───────┴──────────────┘         │
│                             │                                │
│                      ┌──────┴──────┐                         │
│                      │ Agent Loop  │                         │
│                      │ (pkg/agent) │                         │
│                      └──────┬──────┘                         │
│                             │                                │
│                      ┌──────┴──────┐                         │
│                      │  Providers  │                         │
│                      │-------------│                         │
│                      │ Claude      │                         │
│                      │ GPT         │                         │
│                      │ Gemini      │                         │
│                      │ Ollama      │                         │
│                      │ 9 more..    │                         │
│                      └─────────────┘                         │
│                                                             │
│  ┌──────────┐  ┌────────────┐  ┌──────────┐  ┌───────────┐ │
│  │ Perms    │  │ Termux API │  │  Skills  │  │ Security  │ │
│  │ (deny    │  │ (Android   │  │ (extend- │  │ (shell    │ │
│  │ default) │  │  hardware) │  │  able)   │  │  guards)  │ │
│  └──────────┘  └────────────┘  └──────────┘  └───────────┘ │
└─────────────────────────────────────────────────────────────┘
```

---

## Roadmap — Towards a True 24/7 AI Assistant

V1Claw works today as a powerful single-device assistant. Here's what's next to make it a full Jarvis:

### ✅ Done
- [x] 13 LLM providers (cloud + local)
- [x] 10 communication channels
- [x] Voice I/O pipeline (mic → STT → agent → TTS → speaker)
- [x] Vision/camera integration
- [x] Android/Termux full hardware access
- [x] Deny-by-default permission system with per-feature toggles
- [x] Security hardening (hostile audit — 13 critical + 30 high vulns fixed)
- [x] Cron scheduling and proactive tasks
- [x] Skills system (installable agent extensions)
- [x] Docker and cross-platform builds

### 🚧 In Progress
- [ ] Multi-device sync — share one brain across phone, laptop, and server
- [ ] Client mode — thin client connecting to a remote V1Claw brain over Tailscale/LAN
- [ ] Device capability routing — "use the phone's camera" from desktop
- [ ] Always-on wake word — persistent low-power listening mode

### 🔮 Future
- [ ] Local STT/TTS — fully offline voice without cloud APIs
- [ ] Smart home integration — control IoT devices
- [ ] Context-aware proactive suggestions — "you have a meeting in 10 minutes"
- [ ] End-to-end encryption for multi-device communication
- [ ] Web dashboard with real-time status and conversation history
- [ ] Plugin marketplace for community-built skills

---

## Contributing

V1Claw is an ambitious project to build a truly personal, open-source AI assistant that rivals commercial offerings like Alexa, Siri, or Google Assistant — but with full privacy, no cloud lock-in, and the power of frontier LLMs.

**We're looking for developers to help build the future of personal AI:**

- 🧠 **AI/ML engineers** — local model integration, fine-tuning, RAG pipelines
- 🎤 **Audio/voice engineers** — wake word detection, noise cancellation, streaming STT
- 📱 **Mobile developers** — native Android/iOS clients, Wear OS support
- 🏠 **IoT/smart home developers** — Home Assistant, Matter/Thread integration
- 🔒 **Security researchers** — sandboxing, formal verification, threat modeling
- 🌐 **Full-stack developers** — web dashboard, admin panel, device management
- 📝 **Technical writers** — documentation, tutorials, deployment guides

Every contribution matters — from fixing a typo to implementing a new provider. Open an issue, submit a PR, or just star the repo to show support.

**[GitHub Issues](https://github.com/amit-vikramaditya/V1Claw/issues)** — Report bugs or request features
**[Pull Requests](https://github.com/amit-vikramaditya/V1Claw/pulls)** — Submit your contributions

---

## License

MIT License — see [LICENSE](LICENSE) for details.

---

## Acknowledgements

V1Claw is heavily inspired by and based on **[PicoClaw](https://github.com/sipeed/picoclaw)** by Sipeed. We are grateful to the PicoClaw team for building the foundation that made this project possible.
