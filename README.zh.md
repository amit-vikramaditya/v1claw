<div align="center">
  <img src="assets/logo.jpg" alt="V1Claw" width="512">

  <h1>V1Claw: Go语言超高效24/7 AI助手</h1>

  <h3>"Hello V1" · 24/7助手 · 语音 · 视觉 · 智能家居 · 跨设备</h3>

  <p>
    <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go&logoColor=white" alt="Go">
    <img src="https://img.shields.io/badge/Arch-x86__64%2C%20ARM64%2C%20RISC--V-blue" alt="Hardware">
    <img src="https://img.shields.io/badge/license-MIT-green" alt="License">
    <br>
    <a href="https://v1claw.io"><img src="https://img.shields.io/badge/Website-v1claw.io-blue?style=flat&logo=google-chrome&logoColor=white" alt="Website"></a>
    <a href="https://x.com/SipeedIO"><img src="https://img.shields.io/badge/X_(Twitter)-SipeedIO-black?style=flat&logo=x&logoColor=white" alt="Twitter"></a>
  </p>

 **中文** | [日本語](README.ja.md) | [English](README.md)
</div>

---

🤖 V1Claw是一个24/7个人AI助手（"Hello V1"）——超轻量、事件驱动、语音支持、跨设备。受[PicoClaw](https://github.com/sipeed/picoclaw)和[nanobot](https://github.com/HKUDS/nanobot)启发，从零开始重构，新增8大功能层：事件路由、语音（TTS+唤醒词）、RAG知识引擎、智能家居/日历/邮件集成、视觉、跨设备同步、Web仪表板和主动智能。

⚡️ 在$10硬件上运行，内存不到10MB：比OpenClaw少99%的内存，比Mac mini便宜98%！

<table align="center">
  <tr align="center">
    <td align="center" valign="top">
      <p align="center">
        <img src="assets/v1claw_mem.gif" width="360" height="240">
      </p>
    </td>
    <td align="center" valign="top">
      <p align="center">
        <img src="assets/licheervnano.png" width="400" height="240">
      </p>
    </td>
  </tr>
</table>

> [!CAUTION]
> **🚨 安全声明与官方渠道**
>
> * **无加密货币：** V1Claw**没有**官方代币/币。`pump.fun`或其他交易平台上的所有声明都是**骗局**。
> * **官方域名：** 唯一的官方网站是**[v1claw.io](https://v1claw.io)**，公司官网是**[sipeed.com](https://sipeed.com)**
> * **警告：** 许多`.ai/.org/.com/.net/...`域名由第三方注册。
> * **警告：** v1claw目前处于早期开发阶段，可能存在未解决的网络安全问题。在v1.0发布之前请勿部署到生产环境。
> * **注意：** v1claw最近合并了大量PR，最新版本可能导致更大的内存占用（10-20MB）。我们计划在当前功能集稳定后优先进行资源优化。


## 📢 新闻
2026-02-16 🎉 V1Claw hit 12K stars in one week! Thank you all for your support! V1Claw is growing faster than we ever imagined. Given the high volume of PRs, we urgently need community maintainers. Our volunteer roles and roadmap are officially posted [here](docs/v1claw_community_roadmap_260216.md) —we can’t wait to have you on board!

2026-02-13 🎉 V1Claw hit 5000 stars in 4days! Thank you for the community! There are so many PRs&issues come in (during Chinese New Year holidays), we are finalizing the Project Roadmap and setting up the Developer Group to accelerate V1Claw's development.  
🚀 Call to Action: Please submit your feature requests in GitHub Discussions. We will review and prioritize them during our upcoming weekly meeting.

2026-02-09 🎉 V1Claw Launched! Built in 1 day to bring AI Agents to $10 hardware with <10MB RAM. 🦐 V1Claw，Let's Go！

## ✨ 特性

🪶 **超轻量**: 内存占用不到10MB — 比Clawdbot核心功能小99%。

💰 **最低成本**: 可在$10硬件上高效运行 — 比Mac mini便宜98%。

⚡️ **极速启动**: 启动速度快400倍，0.6GHz单核也能1秒启动。

🌍 **真正便携**: 跨RISC-V、ARM和x86的单一独立二进制文件，一键运行！

🤖 **AI自举**: 自主Go原生实现 — 95% Agent-generated core with human-in-the-loop refinement.

|                               | OpenClaw      | NanoBot                  | **V1Claw**                              |
| ----------------------------- | ------------- | ------------------------ | ----------------------------------------- |
| **语言**                  | TypeScript    | Python                   | **Go**                                    |
| **RAM**                       | >1GB          | >100MB                   | **< 10MB**                                |
| **启动时间**</br>(0.8GHz core) | >500s         | >30s                     | **<1s**                                   |
| **成本**                      | Mac Mini 599$ | 大多数Linux SBC </br>~50$ | **任何Linux开发板**</br>**最低$10** |

<img src="assets/compare.jpg" alt="V1Claw" width="512">

## 🦾 演示

### 🛠️ 标准助手工作流

<table align="center">
  <tr align="center">
    <th><p align="center">🧩 全栈工程师</p></th>
    <th><p align="center">🗂️ 日志与规划管理</p></th>
    <th><p align="center">🔎 网络搜索与学习</p></th>
  </tr>
  <tr>
    <td align="center"><p align="center"><img src="assets/v1claw_code.gif" width="240" height="180"></p></td>
    <td align="center"><p align="center"><img src="assets/v1claw_memory.gif" width="240" height="180"></p></td>
    <td align="center"><p align="center"><img src="assets/v1claw_search.gif" width="240" height="180"></p></td>
  </tr>
  <tr>
    <td align="center">开发 • 部署 • 扩展</td>
    <td align="center">计划 • 自动化 • 记忆</td>
    <td align="center">发现 • 洞察 · 趋势</td>
  </tr>
</table>

### 📱 在旧Android手机上运行
让你的旧手机焕发新生！用V1Claw将其变成智能AI助手。快速开始：
1. **安装Termux**（可从F-Droid或Google Play获取）。
2. **执行命令**
```bash
# Note: Replace v0.1.1 with the latest version from the Releases page
wget https://github.com/amit-vikramaditya/v1claw/releases/download/v0.1.1/v1claw-linux-arm64
chmod +x v1claw-linux-arm64
pkg install proot
termux-chroot ./v1claw-linux-arm64 onboard
```
然后按照「快速开始」部分的说明完成配置！
<img src="assets/termux.jpg" alt="V1Claw" width="512">

### 🐜 创新低占用部署

V1Claw可以部署在几乎任何Linux设备上！

- $9.9 [LicheeRV-Nano](https://www.aliexpress.com/item/1005006519668532.html) E(以太网)或W(WiFi6)版本，用于最小化家庭助手
- $30~50 [NanoKVM](https://www.aliexpress.com/item/1005007369816019.html), or $100 [NanoKVM-Pro](https://www.aliexpress.com/item/1005010048471263.html) 用于自动化服务器维护
- $50 [MaixCAM](https://www.aliexpress.com/item/1005008053333693.html) or $100 [MaixCAM2](https://www.kickstarter.com/projects/zepan/maixcam2-build-your-next-gen-4k-ai-camera) 用于智能监控

<https://private-user-images.githubusercontent.com/83055338/547056448-e7b031ff-d6f5-4468-bcca-5726b6fecb5c.mp4>

🌟 More Deployment Cases Await！

## 📦 安装

### 使用预编译二进制文件安装

从[发布](https://github.com/amit-vikramaditya/v1claw/releases)页面下载适合您平台的固件。

### 从源码安装（最新功能，推荐开发使用）

```bash
git clone https://github.com/amit-vikramaditya/v1claw.git

cd v1claw
make deps

# Build, no need to install
make build

# Build for multiple platforms
make build-all

# Build And Install
make install
```

## 🐳 Docker Compose

您也可以使用Docker Compose运行V1Claw，无需在本地安装任何内容。

```bash
# 1. Clone this repo
git clone https://github.com/amit-vikramaditya/v1claw.git
cd v1claw

# 2. Set your API keys
cp config/config.example.json config/config.json
vim config/config.json      # Set DISCORD_BOT_TOKEN, API keys, etc.

# 3. Build & Start
docker compose --profile gateway up -d

# 4. Check logs
docker compose logs -f v1claw-gateway

# 5. Stop
docker compose --profile gateway down
```

### 智能体模式（单次）

```bash
# Ask a question
docker compose run --rm v1claw-agent -m "What is 2+2?"

# Interactive mode
docker compose run --rm v1claw-agent
```

### 重新构建

```bash
docker compose --profile gateway build --no-cache
docker compose --profile gateway up -d
```

### 🚀 快速开始

> [!TIP]
> Set your API key in `~/.v1claw/config.json`.
> Get API keys: [OpenRouter](https://openrouter.ai/keys) (LLM) · [Zhipu](https://open.bigmodel.cn/usercenter/proj-mgmt/apikeys) (LLM)
> Web search is **optional** - get free [Brave Search API](https://brave.com/search/api) (2000 free queries/month) or use built-in auto fallback.

**1. 初始化**

```bash
v1claw onboard
```

**2. 配置** (`~/.v1claw/config.json`)

```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.v1claw/workspace",
      "model": "glm-4.7",
      "max_tokens": 8192,
      "temperature": 0.7,
      "max_tool_iterations": 20
    }
  },
  "providers": {
    "openrouter": {
      "api_key": "xxx",
      "api_base": "https://openrouter.ai/api/v1"
    }
  },
  "tools": {
    "web": {
      "brave": {
        "enabled": false,
        "api_key": "YOUR_BRAVE_API_KEY",
        "max_results": 5
      },
      "duckduckgo": {
        "enabled": true,
        "max_results": 5
      }
    }
  }
}
```

**3. 获取API密钥**

* **LLM Provider**: [OpenRouter](https://openrouter.ai/keys) · [Zhipu](https://open.bigmodel.cn/usercenter/proj-mgmt/apikeys) · [Anthropic](https://console.anthropic.com) · [OpenAI](https://platform.openai.com) · [Gemini](https://aistudio.google.com/api-keys)
* **Web Search** (optional): [Brave Search](https://brave.com/search/api) - 免费层可用 (2000 requests/month)

> **Note**: See `config.example.json` for a complete configuration template.

**4. 聊天**

```bash
v1claw agent -m "What is 2+2?"
```

就这样！2分钟内您就有了一个可用的AI助手。

---

## 💬 聊天应用

通过Telegram、Discord、钉钉或LINE与您的v1claw对话

| Channel      | Setup                              |
| ------------ | ---------------------------------- |
| **Telegram** | 简单（仅需令牌）                |
| **Discord**  | 简单（机器人令牌+意图）         |
| **QQ**       | 简单（AppID+AppSecret）           |
| **DingTalk** | 中等（应用凭证）           |
| **LINE**     | 中等（凭证+Webhook URL） |

<details>
<summary><b>Telegram</b> （推荐）</summary>

**1. Create a bot**

* Open Telegram, search `@BotFather`
* Send `/newbot`, follow prompts
* Copy the token

**2. 配置**

```json
{
  "channels": {
    "telegram": {
      "enabled": true,
      "token": "YOUR_BOT_TOKEN",
      "allowFrom": ["YOUR_USER_ID"]
    }
  }
}
```

> Get your user ID from `@userinfobot` on Telegram.

**3. Run**

```bash
v1claw gateway
```

</details>

<details>
<summary><b>Discord</b></summary>

**1. Create a bot**

* Go to <https://discord.com/developers/applications>
* Create an application → Bot → Add Bot
* Copy the bot token

**2. Enable intents**

* In the Bot settings, enable **MESSAGE CONTENT INTENT**
* (选项al) Enable **SERVER MEMBERS INTENT** if you plan to use allow lists based on member data

**3. Get your User ID**

* Discord Settings → Advanced → enable **Developer Mode**
* Right-click your avatar → **Copy User ID**

**4. Configure**

```json
{
  "channels": {
    "discord": {
      "enabled": true,
      "token": "YOUR_BOT_TOKEN",
      "allowFrom": ["YOUR_USER_ID"]
    }
  }
}
```

**5. Invite the bot**

* OAuth2 → URL Generator
* Scopes: `bot`
* Bot Permissions: `Send Messages`, `Read Message History`
* Open the generated invite URL and add the bot to your server

**6. Run**

```bash
v1claw gateway
```

</details>

<details>
<summary><b>QQ</b></summary>

**1. Create a bot**

- Go to [QQ Open Platform](https://q.qq.com/#)
- Create an application → Get **AppID** and **AppSecret**

**2. 配置**

```json
{
  "channels": {
    "qq": {
      "enabled": true,
      "app_id": "YOUR_APP_ID",
      "app_secret": "YOUR_APP_SECRET",
      "allow_from": []
    }
  }
}
```

> Set `allow_from` to empty to allow all users, or specify QQ numbers to restrict access.

**3. Run**

```bash
v1claw gateway
```

</details>

<details>
<summary><b>DingTalk</b></summary>

**1. Create a bot**

* Go to [Open Platform](https://open.dingtalk.com/)
* Create an internal app
* Copy Client ID and Client Secret

**2. 配置**

```json
{
  "channels": {
    "dingtalk": {
      "enabled": true,
      "client_id": "YOUR_CLIENT_ID",
      "client_secret": "YOUR_CLIENT_SECRET",
      "allow_from": []
    }
  }
}
```

> Set `allow_from` to empty to allow all users, or specify QQ numbers to restrict access.

**3. Run**

```bash
v1claw gateway
```

</details>

<details>
<summary><b>LINE</b></summary>

**1. Create a LINE Official Account**

- Go to [LINE Developers Console](https://developers.line.biz/)
- Create a provider → Create a Messaging API channel
- Copy **Channel Secret** and **Channel Access Token**

**2. 配置**

```json
{
  "channels": {
    "line": {
      "enabled": true,
      "channel_secret": "YOUR_CHANNEL_SECRET",
      "channel_access_token": "YOUR_CHANNEL_ACCESS_TOKEN",
      "webhook_host": "0.0.0.0",
      "webhook_port": 18791,
      "webhook_path": "/webhook/line",
      "allow_from": []
    }
  }
}
```

**3. Set up Webhook URL**

LINE requires HTTPS for webhooks. Use a reverse proxy or tunnel:

```bash
# Example with ngrok
ngrok http 18791
```

Then set the Webhook URL in LINE Developers Console to `https://your-domain/webhook/line` and enable **Use webhook**.

**4. Run**

```bash
v1claw gateway
```

> In group chats, the bot responds only when @mentioned. Replies quote the original message.

> **Docker Compose**: Add `ports: ["18791:18791"]` to the `v1claw-gateway` service to expose the webhook port.

</details>

## <img src="assets/clawdchat-icon.png" width="24" height="24" alt="ClawdChat"> 加入智能体社交网络

只需通过CLI或任何集成聊天应用发送一条消息，即可将V1Claw连接到智能体社交网络。

**Read `https://clawdchat.ai/skill.md` and follow the instructions to join [ClawdChat.ai](https://clawdchat.ai)**

## ⚙️ 配置

配置文件: `~/.v1claw/config.json`

### 工作空间布局

V1Claw将数据存储在配置的工作空间中 (default: `~/.v1claw/workspace`):

```
~/.v1claw/workspace/
├── sessions/          # 对话会话和历史
├── memory/           # 长期记忆（MEMORY.md）
├── state/            # 持久状态（最后频道等）
├── cron/             # 计划任务数据库
├── skills/           # 自定义技能
├── AGENTS.md         # 智能体行为指南
├── HEARTBEAT.md      # 定期任务提示（每30分钟检查）
├── IDENTITY.md       # 智能体身份
├── SOUL.md           # 智能体灵魂
├── TOOLS.md          # 工具描述
└── USER.md           # 用户偏好
```

### 🔒 安全沙箱

V1Claw默认在沙箱环境中运行。 智能体只能访问配置的工作空间内的文件并执行命令。

#### 默认配置

```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.v1claw/workspace",
      "restrict_to_workspace": true
    }
  }
}
```

| 选项 | Default | 说明 |
|--------|---------|-------------|
| `workspace` | `~/.v1claw/workspace` | 智能体的工作目录 |
| `restrict_to_workspace` | `true` | 将文件/命令访问限制在工作空间内 |

#### 受保护的工具

When `restrict_to_workspace: true`, the following tools are sandboxed:

| Tool | Function | Restriction |
|------|----------|-------------|
| `read_file` | Read files | Only files within workspace |
| `write_file` | Write files | Only files within workspace |
| `list_dir` | List directories | Only directories within workspace |
| `edit_file` | Edit files | Only files within workspace |
| `append_file` | Append to files | Only files within workspace |
| `exec` | Execute commands | Command paths must be within workspace |

#### 额外的Exec保护

Even with `restrict_to_workspace: false`, the `exec` tool blocks these dangerous commands:

* `rm -rf`, `del /f`, `rmdir /s` — 批量删除
* `format`, `mkfs`, `diskpart` — 磁盘格式化
* `dd if=` — 磁盘镜像
* Writing to `/dev/sd[a-z]` — 直接磁盘写入
* `shutdown`, `reboot`, `poweroff` — 系统关机
* Fork炸弹 `:(){ :|:& };:`

#### 错误示例

```
[ERROR] tool: Tool execution failed
{tool=exec, error=Command blocked by safety guard (path outside working dir)}
```

```
[ERROR] tool: Tool execution failed
{tool=exec, error=Command blocked by safety guard (dangerous pattern detected)}
```

#### 禁用限制（安全风险）

如果需要智能体访问工作空间外的路径：

**方法1: 配置文件**

```json
{
  "agents": {
    "defaults": {
      "restrict_to_workspace": false
    }
  }
}
```

**方法2: 环境变量**

```bash
export V1CLAW_AGENTS_DEFAULTS_RESTRICT_TO_WORKSPACE=false
```

> ⚠️ **警告**: 禁用此限制将允许智能体访问系统上的任何路径。仅在受控环境中谨慎使用。

#### 安全边界一致性

The `restrict_to_workspace` setting applies consistently across all execution paths:

| 执行路径 | 安全边界 |
|----------------|-------------------|
| Main Agent | `restrict_to_workspace` ✅ |
| Subagent / Spawn | 继承相同限制 ✅ |
| Heartbeat tasks | 继承相同限制 ✅ |

All paths share the same workspace restriction — there's no way to bypass the security boundary through subagents or scheduled tasks.

### 心跳（定期任务）

V1Claw可以自动执行定期任务。 Create a `HEARTBEAT.md` file in your workspace:

```markdown
# Periodic Tasks

- Check my email for important messages
- Review my calendar for upcoming events
- Check the weather forecast
```

The agent will read this file every 30 minutes (configurable) and execute any tasks using available tools.

#### 使用Spawn的异步任务

For long-running tasks (web search, API calls), use the `spawn` tool to create a **subagent**:

```markdown
# Periodic Tasks

## Quick Tasks (respond directly)
- Report current time

## Long Tasks (use spawn for async)
- Search the web for AI news and summarize
- Check email and report important messages
```

**Key behaviors:**

| Feature | 说明 |
|---------|-------------|
| **spawn** | Creates async subagent, doesn't block heartbeat |
| **Independent context** | Subagent has its own context, no session history |
| **message tool** | Subagent communicates with user directly via message tool |
| **Non-blocking** | After spawning, heartbeat continues to next task |

#### 子智能体通信工作原理

```
Heartbeat triggers
    ↓
Agent reads HEARTBEAT.md
    ↓
For long task: spawn subagent
    ↓                           ↓
Continue to next task      Subagent works independently
    ↓                           ↓
All tasks done            Subagent uses "message" tool
    ↓                           ↓
Respond HEARTBEAT_OK      User receives result directly
```

The subagent has access to tools (message, web_search, etc.) and can communicate with the user independently without going through the main agent.

**Configuration:**

```json
{
  "heartbeat": {
    "enabled": true,
    "interval": 30
  }
}
```

| 选项 | Default | 说明 |
|--------|---------|-------------|
| `enabled` | `true` | 启用/禁用心跳 |
| `interval` | `30` | 检查间隔（分钟）（最小: 5） |

**Environment variables:**

* `V1CLAW_HEARTBEAT_ENABLED=false` to disable
* `V1CLAW_HEARTBEAT_INTERVAL=60` to change interval

### 提供商

> [!NOTE]
> Groq provides free voice transcription via Whisper. If configured, Telegram voice messages will be automatically transcribed.

| Provider                   | 用途                                 | 获取API密钥                                            |
| -------------------------- | --------------------------------------- | ------------------------------------------------------ |
| `gemini`                   | LLM (Gemini direct)                     | [aistudio.google.com](https://aistudio.google.com)     |
| `zhipu`                    | LLM (Zhipu direct)                      | [bigmodel.cn](bigmodel.cn)                             |
| `openrouter(To be tested)` | LLM（推荐，访问所有模型） | [openrouter.ai](https://openrouter.ai)                 |
| `anthropic(To be tested)`  | LLM (Claude direct)                     | [console.anthropic.com](https://console.anthropic.com) |
| `openai(To be tested)`     | LLM (GPT direct)                        | [platform.openai.com](https://platform.openai.com)     |
| `deepseek(To be tested)`   | LLM (DeepSeek direct)                   | [platform.deepseek.com](https://platform.deepseek.com) |
| `groq`                     | LLM + **Voice transcription** (Whisper) | [console.groq.com](https://console.groq.com)           |

<details>
<summary><b>Zhipu</b></summary>

**1. Get API key and base URL**

* Get [API key](https://bigmodel.cn/usercenter/proj-mgmt/apikeys)

**2. 配置**

```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.v1claw/workspace",
      "model": "glm-4.7",
      "max_tokens": 8192,
      "temperature": 0.7,
      "max_tool_iterations": 20
    }
  },
  "providers": {
    "zhipu": {
      "api_key": "Your API Key",
      "api_base": "https://open.bigmodel.cn/api/paas/v4"
    }
  }
}
```

**3. Run**

```bash
v1claw agent -m "Hello"
```

</details>

<details>
<summary><b>Full config example</b></summary>

```json
{
  "agents": {
    "defaults": {
      "model": "anthropic/claude-opus-4-5"
    }
  },
  "providers": {
    "openrouter": {
      "api_key": "sk-or-v1-xxx"
    },
    "groq": {
      "api_key": "gsk_xxx"
    }
  },
  "channels": {
    "telegram": {
      "enabled": true,
      "token": "123456:ABC...",
      "allow_from": ["123456789"]
    },
    "discord": {
      "enabled": true,
      "token": "",
      "allow_from": [""]
    },
    "whatsapp": {
      "enabled": false
    },
    "feishu": {
      "enabled": false,
      "app_id": "cli_xxx",
      "app_secret": "xxx",
      "encrypt_key": "",
      "verification_token": "",
      "allow_from": []
    },
    "qq": {
      "enabled": false,
      "app_id": "",
      "app_secret": "",
      "allow_from": []
    }
  },
  "tools": {
    "web": {
      "brave": {
        "enabled": false,
        "api_key": "BSA...",
        "max_results": 5
      },
      "duckduckgo": {
        "enabled": true,
        "max_results": 5
      }
    },
    "cron": {
      "exec_timeout_minutes": 5
    }
  },
  "heartbeat": {
    "enabled": true,
    "interval": 30
  }
}
```

</details>

## CLI参考

| Command                   | 说明                   |
| ------------------------- | ----------------------------- |
| `v1claw onboard`        | 初始化配置和工作空间 |
| `v1claw agent -m "..."` | 与智能体聊天           |
| `v1claw agent`          | 交互式聊天模式         |
| `v1claw gateway`        | 启动网关             |
| `v1claw status`         | 显示状态                   |
| `v1claw cron list`      | 列出所有计划任务       |
| `v1claw cron add ...`   | 添加计划任务           |

### 计划任务 / 提醒

V1Claw通过`cron`工具支持定时提醒和循环任务：

* **One-time reminders**: "Remind me in 10 minutes" → triggers once after 10min
* **Recurring tasks**: "Remind me every 2 hours" → triggers every 2 hours
* **Cron expressions**: "Remind me at 9am daily" → uses cron expression

Jobs are stored in `~/.v1claw/workspace/cron/` and processed automatically.

## 🤝 贡献与路线图

欢迎PR！代码库刻意保持小巧和可读。🤗

路线图即将公布...

开发者群组建设中，入群要求：至少1个已合并的PR。

用户群组：

discord:  <https://discord.gg/V4sAZ9XWpN>

<img src="assets/wechat.png" alt="V1Claw" width="512">

## 🐛 故障排除

### Web search says "API 配置问题"

如果您尚未配置搜索API密钥，这是正常的。 V1Claw将提供手动搜索的有用链接。

要启用Web搜索：

1. **选项 1 （推荐）**: Get a free API key at [https://brave.com/search/api](https://brave.com/search/api) (2000 free queries/month) for the best results.
2. **选项 2 (No Credit Card)**: If you don't have a key, we automatically fall back to **DuckDuckGo** (no key required).

Add the key to `~/.v1claw/config.json` if using Brave:

```json
{
  "tools": {
    "web": {
      "brave": {
        "enabled": false,
        "api_key": "YOUR_BRAVE_API_KEY",
        "max_results": 5
      },
      "duckduckgo": {
        "enabled": true,
        "max_results": 5
      }
    }
  }
}
```

### 内容过滤错误

某些提供商（如智谱）有内容过滤。尝试改写查询或使用其他模型。

### Telegram bot says "Conflict: terminated by other getUpdates"

当另一个机器人实例正在运行时会发生这种情况。确保同一时间只运行一个`v1claw gateway`。

---

## 📝 API密钥比较

| Service          | 免费额度           | 用例                              |
| ---------------- | ------------------- | ------------------------------------- |
| **OpenRouter**   | 200K tokens/month   | Multiple models (Claude, GPT-4, etc.) |
| **Zhipu**        | 200K tokens/month   | Best for Chinese users                |
| **Brave Search** | 2000 queries/month  | Web搜索功能              |
| **Groq**         | 免费层可用 | Fast inference (Llama, Mixtral)       |
