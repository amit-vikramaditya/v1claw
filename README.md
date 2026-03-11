# V1Claw

**Your 24/7 Personal AI Assistant — Like Jarvis, but open source.**

V1Claw is a self-hosted AI assistant that runs on your Mac, Linux PC, Windows machine, or Android phone via Termux. Connect any LLM provider you want, talk to it through voice or text, and let it control your device — read files, run commands, browse the web, take photos, send messages, and more.

One binary. No required V1Claw cloud service. Your data stays on your machine unless you choose a cloud model or channel.

Default home directory: `~/.v1claw` on macOS/Linux, `%APPDATA%\\V1Claw` on Windows.
Set `V1CLAW_HOME` to override it.

---

## Features

### 🧠 15 LLM Providers
Connect to any AI model — paid APIs or self-hosted:

| Provider | Type | Models |
|----------|------|--------|
| **Anthropic** | Cloud API | Claude 4, Claude 3.5, etc. |
| **OpenAI** | Cloud API | GPT-5, GPT-4, etc. |
| **Google Gemini** | Cloud API | Gemini 2.x, etc. |
| **Google Vertex AI** | Cloud API | Gemini on Vertex |
| **AWS Bedrock** | Cloud API | Claude, Llama, Nova |
| **Azure OpenAI** | Cloud API | GPT deployments on Azure |
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
- **Deny-by-default permissions** — camera, microphone, screen, notifications, SMS, phone, location, clipboard, sensors, and hardware shell access each require explicit opt-in
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
- Install community skills from GitHub repo paths
- Skills are sandboxed markdown agent configurations

---

## Setup Guide

Pick your device. Follow the steps. You'll have a working AI assistant in under 10 minutes.

> **You need one provider before you start:** either an API key for a cloud model, or a local/self-hosted endpoint like Ollama or vLLM.
> The easiest free cloud option is **Google Gemini** — get a key at [aistudio.google.com/apikey](https://aistudio.google.com/apikey).
> Other cloud options: [OpenAI](https://platform.openai.com/api-keys), [Anthropic](https://console.anthropic.com/), [Groq](https://console.groq.com/keys), [OpenRouter](https://openrouter.ai/keys).

---

### ⚡ Quick Install — macOS & Linux

Run one command. The script prefers the latest GitHub Release binary. If no release is published yet and Go is already installed, it falls back to building from source.

```bash
curl -fsSL https://raw.githubusercontent.com/amit-vikramaditya/v1claw/main/install.sh | bash
```

Then run the 2-minute setup wizard:

```bash
v1claw onboard
```

Then verify the install:

```bash
v1claw doctor
```

Or skip the wizard entirely with one line:

```bash
v1claw onboard --auto --provider gemini --api-key "YOUR_KEY"
```

Other API-key providers: `openai` · `anthropic` · `groq` · `deepseek` · `openrouter` · `nvidia` · `zhipu` · `moonshot`
Keyless or local providers: `vertex` (uses `gcloud auth`) · `bedrock` (uses `~/.aws/credentials`) · `ollama` · `vllm` · `github_copilot`
If your provider does not have a built-in default model, add `--model YOUR_MODEL`.
For local/self-hosted endpoints, you can also pass `--api-base URL`, for example:

```bash
v1claw onboard --auto --provider ollama --model llama3.2
v1claw onboard --auto --provider vllm --api-base http://localhost:8000/v1 --model my-model
```

> **Windows users:** PowerShell quick install:
> ```powershell
> $installer = Join-Path $env:TEMP "v1claw-install.ps1"
> Invoke-WebRequest "https://raw.githubusercontent.com/amit-vikramaditya/v1claw/main/install.ps1" -OutFile $installer
> powershell -ExecutionPolicy Bypass -File $installer
> ```
>
> Then verify:
> ```powershell
> v1claw doctor
> ```

---

### 📱 Android (Termux)

<details>
<summary><b>Click to expand — full step-by-step Android setup</b></summary>

#### Step 1: Install two apps from F-Droid

You need two apps. Install both from [F-Droid](https://f-droid.org/) (NOT from Play Store — the Play Store versions are outdated and broken):

1. **[Termux](https://f-droid.org/en/packages/com.termux/)** — a Linux terminal for Android
2. **[Termux:API](https://f-droid.org/en/packages/com.termux.api/)** — lets V1Claw use your phone's mic, camera, GPS, etc.

> 💡 If you don't have F-Droid, download it first from [f-droid.org](https://f-droid.org/).

#### Step 2: Grant permissions to Termux:API

Open your phone **Settings → Apps → Termux:API → Permissions** and turn on:
- ✅ Microphone
- ✅ Camera
- ✅ Location (optional)

#### Step 3: Install developer tools in Termux

Open **Termux** and type these commands one at a time. Press Enter after each one:

```bash
pkg update
```

It may ask "Do you want to continue?" — type `y` and press Enter.

```bash
pkg upgrade -y
```

```bash
pkg install -y git golang make termux-api
```

This installs Git, Go (the programming language), Make, and the Termux API tools. It takes about 2-3 minutes.

#### Step 4: Download V1Claw

```bash
git clone https://github.com/amit-vikramaditya/v1claw.git
```

```bash
cd v1claw
```

#### Step 5: Build V1Claw

```bash
make build
```

This compiles V1Claw into a single file. It takes 2-5 minutes on a phone. When it finishes, you'll see a file at `build/v1claw-android-arm64`.

> 💡 The Makefile auto-detects Termux and handles everything for you. No extra flags needed.

#### Step 6: Run first-time setup

```bash
./build/v1claw-android-arm64 onboard --auto --provider gemini --api-key "YOUR_API_KEY"
```

*(Alternatively, run just `onboard` to use the interactive wizard, which will also let you name your Agent and configure its Memory Soul).*

That's it — your config is ready in your V1Claw home directory.

#### Step 7: Test it!

```bash
./build/v1claw-android-arm64 agent -m "Hello! What can you do?"
```

You should see the AI respond. **If it does — congratulations, V1Claw is working on your phone!** 🎉

#### Step 8: Start chatting

```bash
./build/v1claw-android-arm64 agent
```

This opens an interactive chat. Type anything and press Enter. Type `exit` or press `Ctrl+C` to quit.

#### Step 9: Enable phone hardware (optional)

Want V1Claw to use your mic, camera, or read notifications? Edit the config again:

```bash
nano "${V1CLAW_HOME:-$HOME/.v1claw}/config.json"
```

Add a `"permissions"` section (you can turn each feature on or off individually):

```json
{
  "permissions": {
    "microphone": true,
    "camera": true,
    "clipboard": true,
    "notifications": true,
    "location": false,
    "sms": false,
    "phone_calls": false,
    "sensors": false,
    "shell_hardware": true
  }
}
```

> 🔒 **Every feature is OFF by default.** Only turn on what you need. You can change these anytime by editing the config and restarting.

#### Step 10: Run V1Claw 24/7 in the background (optional)

```bash
nohup ./build/v1claw-android-arm64 gateway > v1claw.log 2>&1 & echo $! > v1claw.pid
```

This runs V1Claw as a background service that keeps working even if you close Termux.

To check if it's running:

```bash
curl http://127.0.0.1:18790/health
```

To see what it's doing:

```bash
tail -f v1claw.log
```

To stop it:

```bash
kill $(cat v1claw.pid 2>/dev/null || pgrep v1claw)
```

</details>

---

### 🍎 macOS

<details>
<summary><b>Click to expand — full step-by-step macOS setup</b></summary>

#### Option A: Release binary or installer

Open **Terminal** (`Cmd+Space` → "Terminal") and run:

```bash
curl -fsSL https://raw.githubusercontent.com/amit-vikramaditya/v1claw/main/install.sh | bash
```

The installer will use the latest release when one is published. If you already have Go installed and no release exists yet, it can build from source automatically.

Then run the setup wizard:

```bash
v1claw onboard
```

Verify the install:

```bash
v1claw doctor
```

That's it. Skip to **Step 5: Test it** below.

---

#### Option B: Build from source (for developers)

#### Step 1: Install Go (if you don't have it)

Open **Terminal** (press `Cmd+Space`, type "Terminal", press Enter).

Check if Go is installed:

```bash
go version
```

If it says "command not found", install it:

```bash
# Using Homebrew (recommended)
brew install go

# Or download from https://go.dev/dl/
```

Also make sure you have Git and Make (these come pre-installed on most Macs):

```bash
git --version
make --version
```

#### Step 2: Download V1Claw

```bash
git clone https://github.com/amit-vikramaditya/v1claw.git
cd v1claw
```

#### Step 3: Build

```bash
make build
```

The binary will appear at `build/v1claw-darwin-arm64` (Apple Silicon) or `build/v1claw-darwin-amd64` (Intel Mac).

#### Step 4: Run first-time setup

```bash
./build/v1claw-darwin-* onboard --auto --provider gemini --api-key "YOUR_API_KEY"
```

*(Alternatively, run just `onboard` to use the interactive wizard, which will also let you name your Agent and configure its Memory Soul).*

#### Step 5: Test it

```bash
./build/v1claw-darwin-* agent -m "Hello! Tell me a fun fact."
```

If you see a response — **it's working!** 🎉

#### Step 6: Interactive chat

```bash
./build/v1claw-darwin-* agent
```

#### Step 7: Run as a 24/7 service (optional)

```bash
# Quick background mode
nohup ./build/v1claw-darwin-* gateway > v1claw.log 2>&1 &

# Or install to your PATH and use it anywhere
make install
v1claw gateway
```

</details>

---

### 🐧 Linux

<details>
<summary><b>Click to expand — full step-by-step Linux setup</b></summary>

#### Option A: Release binary or installer

```bash
curl -fsSL https://raw.githubusercontent.com/amit-vikramaditya/v1claw/main/install.sh | bash
```

The installer will use the latest release when one is published. If you already have Go installed and no release exists yet, it can build from source automatically.

Then run the setup wizard:

```bash
v1claw onboard
```

Verify the install:

```bash
v1claw doctor
```

That's it. Skip to **Step 5: Test it** below.

---

#### Option B: Build from source (for developers)

#### Step 1: Install Go, Git, and Make

**Ubuntu/Debian:**
```bash
sudo apt update && sudo apt install -y golang git make
```

**Fedora:**
```bash
sudo dnf install -y golang git make
```

**Arch Linux:**
```bash
sudo pacman -S go git make
```

Verify Go is installed:
```bash
go version
```

#### Step 2: Download V1Claw

```bash
git clone https://github.com/amit-vikramaditya/v1claw.git
cd v1claw
```

#### Step 3: Build

```bash
make build
```

The binary will appear at `build/v1claw-linux-amd64` or `build/v1claw-linux-arm64`.

#### Step 4: Run first-time setup

```bash
./build/v1claw-linux-* onboard --auto --provider gemini --api-key "YOUR_API_KEY"
```

*(Alternatively, run just `onboard` to use the interactive wizard, which will also let you name your Agent and configure its Memory Soul).*

#### Step 5: Test it

```bash
./build/v1claw-linux-* agent -m "Hello! What can you do?"
```

#### Step 6: Run as a 24/7 system service (optional)

```bash
# Install the binary
make install

# Create a systemd service
sudo tee /etc/systemd/system/v1claw.service << 'EOF'
[Unit]
Description=V1Claw AI Assistant
After=network.target

[Service]
Type=simple
User=YOUR_USERNAME
ExecStart=/home/YOUR_USERNAME/.local/bin/v1claw gateway
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# Replace YOUR_USERNAME with your actual username, then:
sudo systemctl daemon-reload
sudo systemctl enable --now v1claw

# Check status
sudo systemctl status v1claw
```

</details>

---

### 🪟 Windows

<details>
<summary><b>Click to expand — full step-by-step Windows setup</b></summary>

#### Option A: PowerShell installer

Open **PowerShell** and run:

```powershell
$installer = Join-Path $env:TEMP "v1claw-install.ps1"
Invoke-WebRequest "https://raw.githubusercontent.com/amit-vikramaditya/v1claw/main/install.ps1" -OutFile $installer
powershell -ExecutionPolicy Bypass -File $installer
```

The installer prefers the latest release when one is published. If no release exists yet and Go is already installed, it falls back to building from source automatically.

Then run:

```powershell
v1claw onboard
```

Then verify:

```powershell
v1claw doctor
```

#### Option B: Build from source manually

##### Step 1: Install Go and Git

1. Download and install **Go** from [go.dev/dl](https://go.dev/dl/) — pick the Windows `.msi` installer
2. Download and install **Git** from [git-scm.com](https://git-scm.com/download/win)
3. Download and install **Make** via [GnuWin32](http://gnuwin32.sourceforge.net/packages/make.htm) or use `choco install make` if you have Chocolatey

Open **Command Prompt** or **PowerShell** and verify:

```bash
go version
git --version
```

##### Step 2: Download V1Claw

```bash
git clone https://github.com/amit-vikramaditya/v1claw.git
cd v1claw
```

##### Step 3: Build

```bash
make build
```

Or if Make doesn't work on Windows:

```bash
go build -o build/v1claw.exe ./cmd/v1claw
```

##### Step 4: Run first-time setup

```bash
build\v1claw.exe onboard
```

The setup wizard will ask you to pick a provider, configure credentials or a local endpoint, and test the connection. Your config is ready immediately.

##### Step 5: Test it

```bash
build\v1claw.exe agent -m "Hello! What can you do?"
```

##### Step 6: Interactive chat

```bash
build\v1claw.exe agent
```

</details>

---

### 🐳 Docker (any platform)

<details>
<summary><b>Click to expand — Docker setup</b></summary>

If you have Docker installed, this is the fastest way:

```bash
# Clone the project
git clone https://github.com/amit-vikramaditya/v1claw.git
cd v1claw

# Copy the example config and edit it
cp config/config.example.json config/config.json
nano config/config.json   # Add provider credentials or endpoint settings

# Run a one-shot query
docker compose run --rm v1claw-agent -m "Hello V1Claw!"

# Or run as a 24/7 background service
docker compose --profile gateway up -d
```

</details>

---

### 🔗 Cross-Compile for Another Device

Have a fast PC and want to build V1Claw for a different device (e.g., build on your Mac for your Android phone)?

```bash
# On your PC:
git clone https://github.com/amit-vikramaditya/v1claw.git
cd v1claw

# Build for Android (ARM64)
GOOS=linux GOARCH=arm64 make build

# Build for Linux server (x86_64)
GOOS=linux GOARCH=amd64 make build

# Build for Windows
GOOS=windows GOARCH=amd64 make build

# Build for Raspberry Pi
GOOS=linux GOARCH=arm GOARM=7 make build
```

Then transfer the binary to your target device (via USB, `scp`, `adb push`, or any file sharing method) and follow from **Step 4** of the relevant guide above.

---

### 🌐 Multi-Device Setup (Tailscale)

Want V1Claw on multiple devices sharing one brain? Use [Tailscale](https://tailscale.com/) (free for personal use):

1. Install Tailscale on all your devices
2. Run V1Claw gateway on your main machine (server/desktop):
   ```bash
   # Edit config — set gateway.host to "0.0.0.0" and enable the API:
   # "v1_api": {"enabled": true, "addr": ":18791", "api_key": "your-secret-key"}
   v1claw gateway
   ```
3. From any other device on your Tailscale network:
   ```bash
   # Interactive mode — full chat with the gateway's brain
   v1claw client --server your-server.tail1234.ts.net:18791 --api-key your-secret-key

   # HTTPS / reverse-proxy deployment
   v1claw client --server https://gateway.example.com --api-key your-secret-key

   # One-shot message
   v1claw client -s your-server.tail1234.ts.net:18791 -k your-secret-key -m "Hello from my phone"

   # If auto-discovery picks the wrong LAN/Tailscale address, override it
   v1claw client -s https://gateway.example.com -k your-secret-key --advertise-host phone.local

   # Or use the REST API directly
   curl http://your-server.tail1234.ts.net:18791/api/v1/chat \
     -H "Authorization: Bearer your-secret-key" \
     -d '{"message": "Hello from my phone"}'
   ```

The client auto-detects local hardware (camera, mic, screen) and registers it with the gateway. When the AI needs to take a photo but the server has no camera, it routes the request to your phone's camera automatically.

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
v1claw --help               # Show help
v1claw onboard              # First-time setup wizard
v1claw onboard --refresh    # Upgrade existing config to new schema (non-destructive)
v1claw onboard --auto \     # Non-interactive setup (CI/scripts)
  --provider gemini \
  --api-key YOUR_KEY \
  --api-base http://localhost:8000/v1 \
  --skip-test              # Optional for CI/offline setup
v1claw agent                # Interactive chat
v1claw agent -m "query"     # One-shot query
v1claw gateway              # Start 24/7 daemon
v1claw client -s host[:port] # Connect to a remote gateway
v1claw client -s https://gateway.example.com # HTTPS / reverse-proxy gateway
v1claw configure            # Change settings interactively
v1claw auth login           # Authenticate
v1claw auth status          # Check auth status
v1claw status               # Show system status
v1claw doctor               # Health check (connectivity, config, workspace)
v1claw cron                 # Manage scheduled tasks
v1claw skills list          # List installed skills
v1claw skills install <github-owner/repo[/path]> # Install a GitHub skill
v1claw version              # Show version
```

---

## Releasing

For maintainers, the local preflight is:

```bash
make release-check
```

Release steps are documented in [RELEASING.md](RELEASING.md).

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
- [x] 15 LLM providers (cloud + local)
- [x] 10 communication channels
- [x] Voice I/O pipeline (mic → STT → agent → TTS → speaker)
- [x] Vision/camera integration
- [x] Android/Termux full hardware access
- [x] Deny-by-default permission system with per-feature toggles
- [x] Security hardening (hostile audit — 13 critical + 30 high vulns fixed)
- [x] Cron scheduling and proactive tasks
- [x] Skills system (installable agent extensions)
- [x] Docker and cross-platform builds
- [x] Multi-device sync — device registration, discovery, and heartbeat
- [x] Client mode — `v1claw client --server host[:port]|url` connects to remote gateway
- [x] Device capability routing — use phone's camera/mic from desktop via WebSocket

### 🚧 In Progress
- [ ] Always-on wake word — persistent low-power listening mode

### 🔮 Future
- [ ] Local STT/TTS — fully offline voice without cloud APIs
- [ ] Smart home integration — control IoT devices
- [ ] Context-aware proactive suggestions — "you have a meeting in 10 minutes"
- [ ] End-to-end encryption for multi-device communication
- [ ] Web dashboard with real-time status and conversation history
- [ ] Plugin marketplace for community-built skills

---

## Troubleshooting

| Problem | Solution |
|---------|----------|
| `command not found: go` | Install Go — see setup guide for your platform above |
| `command not found: make` | Linux: `sudo apt install make` / Mac: `xcode-select --install` / Windows: `choco install make` |
| `permission denied` | Run `chmod +x build/v1claw-*` to make the binary executable |
| Build fails with "out of memory" | Close other apps. On Android, phones have limited RAM — try closing background apps |
| `termux-microphone-record: not found` | Install Termux:API app from F-Droid AND run `pkg install termux-api` in Termux |
| API key error / "unauthorized" | Double-check your API key in your V1Claw config file. Make sure there are no extra spaces |
| `connection refused` on port 18790 | The gateway isn't running. Start it with `v1claw gateway` first |
| AI doesn't use my camera/mic | Enable the permission in config: `"microphone": true` or `"camera": true` and restart |
| `go version` shows old version | V1Claw needs Go 1.25+. Update Go from [go.dev/dl](https://go.dev/dl/) |

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
