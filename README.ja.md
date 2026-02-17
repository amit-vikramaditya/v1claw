<div align="center">
  <img src="assets/logo.jpg" alt="V1Claw" width="512">

  <h1>V1Claw: Go製 超効率的な24/7 AIアシスタント</h1>

  <h3>"Hello V1" · 24/7アシスタント · 音声 · ビジョン · スマートホーム · クロスデバイス</h3>

  <p>
    <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go&logoColor=white" alt="Go">
    <img src="https://img.shields.io/badge/Arch-x86__64%2C%20ARM64%2C%20RISC--V-blue" alt="Hardware">
    <img src="https://img.shields.io/badge/license-MIT-green" alt="License">
    <br>
    <a href="https://v1claw.io"><img src="https://img.shields.io/badge/Website-v1claw.io-blue?style=flat&logo=google-chrome&logoColor=white" alt="Website"></a>
    <a href="https://x.com/SipeedIO"><img src="https://img.shields.io/badge/X_(Twitter)-SipeedIO-black?style=flat&logo=x&logoColor=white" alt="Twitter"></a>
  </p>

 [中文](README.zh.md) | **日本語** | [English](README.md)
</div>

---

🤖 V1Clawは24/7パーソナルAIアシスタント（「Hello V1」）です。超軽量、イベント駆動、音声対応、クロスデバイス対応。[PicoClaw](https://github.com/sipeed/picoclaw)と[nanobot](https://github.com/HKUDS/nanobot)にインスパイアされ、8つの追加機能レイヤー（イベントルーティング、音声（TTS＋ウェイクワード）、RAGナレッジエンジン、スマートホーム/カレンダー/メール連携、ビジョン、クロスデバイス同期、Webダッシュボード、プロアクティブインテリジェンス）を加えてゼロから再構築しました。

⚡️ $10のハードウェアで動作、RAM10MB未満：OpenClawより99%少ないメモリ、Mac miniより98%安い！

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
> **🚨 セキュリティと公式チャンネル**
>
> * **暗号通貨なし：** V1Clawには公式トークン/コインは**ありません**。`pump.fun`やその他の取引プラットフォームでの主張はすべて**詐欺**です。
> * **公式ドメイン：** 唯一の公式サイトは**[v1claw.io](https://v1claw.io)**、企業サイトは**[sipeed.com](https://sipeed.com)**です
> * **警告：** 多くの`.ai/.org/.com/.net/...`ドメインは第三者によって登録されています。
> * **警告：** v1clawは現在初期開発段階であり、未解決のネットワークセキュリティ問題がある可能性があります。v1.0リリース前に本番環境にデプロイしないでください。
> * **注意：** v1clawは最近多くのPRをマージしたため、最新バージョンではメモリ使用量が増加（10〜20MB）する可能性があります。現在の機能セットが安定した段階で、リソース最適化を優先する予定です。


## 📢 ニュース
2026-02-16 🎉 V1Clawが1週間で12Kスターを達成！皆様のご支援に感謝します！V1Clawは想像以上のスピードで成長しています。PR数が多いため、コミュニティメンテナーを緊急募集しています。ボランティアの役割とロードマップが公開されました [here](docs/v1claw_community_roadmap_260216.md) —we can’t wait to have you on board!

2026-02-13 🎉 V1Clawが4日で5000スターを達成！コミュニティの皆様に感謝します！（旧正月休暇中にもかかわらず）多くのPRやIssueが寄せられています。プロジェクトロードマップの最終調整と開発者グループの設立を進めています。  
🚀 ご協力のお願い：GitHub Discussionsで機能リクエストを提出してください。今後のウィークリーミーティングで確認・優先順位付けを行います。

2026-02-09 🎉 V1Clawローンチ！$10のハードウェアで10MB未満のRAMでAIエージェントを実現するために1日で構築。�� V1Claw、Let's Go！

## ✨ 特徴

🪶 **超軽量**: メモリ使用量10MB未満 — Clawdbotのコア機能より99%小さい。

💰 **最小コスト**: $10のハードウェアで動作可能 — Mac miniより98%安い。

⚡️ **超高速**: 起動時間400倍高速、0.6GHzシングルコアでも1秒で起動。

🌍 **真のポータビリティ**: RISC-V、ARM、x86対応の単一バイナリ、ワンクリックでGo！

🤖 **AIブートストラップ**: 自律的なGoネイティブ実装 — 95%エージェント生成コア、ヒューマンインザループで洗練。

|                               | OpenClaw      | NanoBot                  | **V1Claw**                              |
| ----------------------------- | ------------- | ------------------------ | ----------------------------------------- |
| **言語**                  | TypeScript    | Python                   | **Go**                                    |
| **RAM**                       | >1GB          | >100MB                   | **< 10MB**                                |
| **起動時間**</br>(0.8GHz core) | >500s         | >30s                     | **<1s**                                   |
| **コスト**                      | Mac Mini 599$ | ほとんどのLinux SBC </br>~50$ | **任意のLinuxボード**</br>**最安$10** |

<img src="assets/compare.jpg" alt="V1Claw" width="512">

## 🦾 デモンストレーション

### 🛠️ 標準アシスタントワークフロー

<table align="center">
  <tr align="center">
    <th><p align="center">🧩 フルスタックエンジニア</p></th>
    <th><p align="center">🗂️ ログ＆プランニング管理</p></th>
    <th><p align="center">🔎 Web検索＆学習</p></th>
  </tr>
  <tr>
    <td align="center"><p align="center"><img src="assets/v1claw_code.gif" width="240" height="180"></p></td>
    <td align="center"><p align="center"><img src="assets/v1claw_memory.gif" width="240" height="180"></p></td>
    <td align="center"><p align="center"><img src="assets/v1claw_search.gif" width="240" height="180"></p></td>
  </tr>
  <tr>
    <td align="center">開発 • デプロイ • スケール</td>
    <td align="center">スケジュール • 自動化 • メモリ</td>
    <td align="center">発見 • インサイト • トレンド</td>
  </tr>
</table>

### 📱 古いAndroidスマホで実行
10年前のスマホに第二の人生を！V1ClawでスマートAIアシスタントに変身。クイックスタート：
1. **Termuxをインストール**（F-DroidまたはGoogle Playから入手可能）。
2. **コマンドを実行**
```bash
# Note: Replace v0.1.1 with the latest version from the Releases page
wget https://github.com/amit-vikramaditya/v1claw/releases/download/v0.1.1/v1claw-linux-arm64
chmod +x v1claw-linux-arm64
pkg install proot
termux-chroot ./v1claw-linux-arm64 onboard
```
その後、「クイックスタート」セクションの手順に従って設定を完了してください！
<img src="assets/termux.jpg" alt="V1Claw" width="512">

### 🐜 革新的な省フットプリントデプロイ

V1ClawはほぼすべてのLinuxデバイスにデプロイ可能！

- $9.9 [LicheeRV-Nano](https://www.aliexpress.com/item/1005006519668532.html) E(Ethernet)またはW(WiFi6)バージョン、最小限のホームアシスタント用
- $30~50 [NanoKVM](https://www.aliexpress.com/item/1005007369816019.html), or $100 [NanoKVM-Pro](https://www.aliexpress.com/item/1005010048471263.html) 自動サーバーメンテナンス用
- $50 [MaixCAM](https://www.aliexpress.com/item/1005008053333693.html) or $100 [MaixCAM2](https://www.kickstarter.com/projects/zepan/maixcam2-build-your-next-gen-4k-ai-camera) スマートモニタリング用

<https://private-user-images.githubusercontent.com/83055338/547056448-e7b031ff-d6f5-4468-bcca-5726b6fecb5c.mp4>

🌟 さらなるデプロイ事例をお楽しみに！

## 📦 インストール

### ビルド済みバイナリでインストール

[リリース](https://github.com/amit-vikramaditya/v1claw/releases)ページからプラットフォームに合ったファームウェアをダウンロードしてください。

### ソースからインストール（最新機能、開発推奨）

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

Docker Composeを使えば、ローカルに何もインストールせずにV1Clawを実行できます。

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

### エージェントモード（ワンショット）

```bash
# Ask a question
docker compose run --rm v1claw-agent -m "What is 2+2?"

# Interactive mode
docker compose run --rm v1claw-agent
```

### 再ビルド

```bash
docker compose --profile gateway build --no-cache
docker compose --profile gateway up -d
```

### 🚀 クイックスタート

> [!TIP]
> APIキーを`~/.v1claw/config.json`に設定してください。
> Get API keys: [OpenRouter](https://openrouter.ai/keys) (LLM) · [Zhipu](https://open.bigmodel.cn/usercenter/proj-mgmt/apikeys) (LLM)
> Web検索は**オプション**です - 無料の [Brave Search API](https://brave.com/search/api) （月2000回無料）を取得するか、組み込みの自動フォールバックを使用してください。

**1. 初期化**

```bash
v1claw onboard
```

**2. 設定** (`~/.v1claw/config.json`)

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

**3. APIキーの取得**

* **LLM Provider**: [OpenRouter](https://openrouter.ai/keys) · [Zhipu](https://open.bigmodel.cn/usercenter/proj-mgmt/apikeys) · [Anthropic](https://console.anthropic.com) · [OpenAI](https://platform.openai.com) · [Gemini](https://aistudio.google.com/api-keys)
* **Web検索**（オプション）: [Brave Search](https://brave.com/search/api) - 無料枠あり（月2000リクエスト）

> **Note**: 完全な設定テンプレートは`config.example.json`を参照してください。

**4. チャット**

```bash
v1claw agent -m "What is 2+2?"
```

以上です！2分でAIアシスタントが動作します。

---

## 💬 チャットアプリ

Telegram、Discord、DingTalk、LINEを通じてv1clawと会話

| Channel      | Setup                              |
| ------------ | ---------------------------------- |
| **Telegram** | 簡単（トークンのみ）                |
| **Discord**  | 簡単（ボットトークン＋インテント）         |
| **QQ**       | 簡単（AppID＋AppSecret）           |
| **DingTalk** | 中程度（アプリ認証情報）           |
| **LINE**     | 中程度（認証情報＋Webhook URL） |

<details>
<summary><b>Telegram</b> （推奨）</summary>

**1. ボットを作成**

* Telegramを開き、`@BotFather`を検索
* `/newbot`を送信し、指示に従う
* トークンをコピー

**2. 設定**

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

> Telegramの`@userinfobot`からユーザーIDを取得してください。

**3. 実行**

```bash
v1claw gateway
```

</details>

<details>
<summary><b>Discord</b></summary>

**1. ボットを作成**

* <https://discord.com/developers/applications>にアクセス
* アプリケーションを作成 → Bot → ボットを追加
* ボットトークンをコピー

**2. インテントを有効化**

* Bot設定で**MESSAGE CONTENT INTENT**を有効化
*（オプション）メンバーデータに基づく許可リストを使用する場合は**SERVER MEMBERS INTENT**を有効化

**3. ユーザーIDを取得**

* Discord設定 → 詳細設定 → **開発者モード**を有効化
* アバターを右クリック → **ユーザーIDをコピー**

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

**5. ボットを招待**

* OAuth2 → URLジェネレーター
* スコープ: `bot`
* ボット権限: `Send Messages`, `Read Message History`
* 生成された招待URLを開き、サーバーにボットを追加

**6. Run**

```bash
v1claw gateway
```

</details>

<details>
<summary><b>QQ</b></summary>

**1. ボットを作成**

- [QQオープンプラットフォーム](https://q.qq.com/#)にアクセス
- アプリケーションを作成 → **AppID**と**AppSecret**を取得

**2. 設定**

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

> `allow_from`を空にすると全ユーザーを許可、QQ番号を指定するとアクセスを制限します。

**3. 実行**

```bash
v1claw gateway
```

</details>

<details>
<summary><b>DingTalk</b></summary>

**1. ボットを作成**

* [オープンプラットフォーム](https://open.dingtalk.com/)にアクセス
* 内部アプリを作成
* Client IDとClient Secretをコピー

**2. 設定**

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

> `allow_from`を空にすると全ユーザーを許可、QQ番号を指定するとアクセスを制限します。

**3. 実行**

```bash
v1claw gateway
```

</details>

<details>
<summary><b>LINE</b></summary>

**1. LINE公式アカウントを作成**

- [LINE Developersコンソール](https://developers.line.biz/)にアクセス
- プロバイダーを作成 → Messaging APIチャンネルを作成
- **Channel Secret**と**Channel Access Token**をコピー

**2. 設定**

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

**3. Webhook URLを設定**

LINEのWebhookにはHTTPSが必要です。リバースプロキシまたはトンネルを使用してください：

```bash
# Example with ngrok
ngrok http 18791
```

LINE DevelopersコンソールでWebhook URLを`https://your-domain/webhook/line`に設定し、**Webhookを使用**を有効にしてください。

**4. Run**

```bash
v1claw gateway
```

> グループチャットでは、ボットは@メンションされた場合にのみ応答します。返信は元のメッセージを引用します。

> **Docker Compose**: Add `ports: ["18791:18791"]` to the `v1claw-gateway` service to expose the webhook port.

</details>

## <img src="assets/clawdchat-icon.png" width="24" height="24" alt="ClawdChat"> エージェントソーシャルネットワークに参加

CLIまたは統合チャットアプリからメッセージを1つ送信するだけで、V1Clawをエージェントソーシャルネットワークに接続できます。

**`https://clawdchat.ai/skill.md`を読み、指示に従って[ClawdChat.ai](https://clawdchat.ai)に参加してください**

## ⚙️ 設定

設定ファイル: `~/.v1claw/config.json`

### ワークスペースレイアウト

V1Clawは設定されたワークスペース（デフォルト: `~/.v1claw/workspace`）にデータを保存します：

```
~/.v1claw/workspace/
├── sessions/          # 会話セッションと履歴
├── memory/           # 長期記憶（MEMORY.md）
├── state/            # 永続状態（最後のチャンネルなど）
├── cron/             # スケジュールジョブデータベース
├── skills/           # カスタムスキル
├── AGENTS.md         # エージェント行動ガイド
├── HEARTBEAT.md      # 定期タスクプロンプト（30分ごとにチェック）
├── IDENTITY.md       # エージェントアイデンティティ
├── SOUL.md           # エージェントソウル
├── TOOLS.md          # ツール説明
└── USER.md           # ユーザー設定
```

### 🔒 セキュリティサンドボックス

V1Clawはデフォルトでサンドボックス環境で実行されます。エージェントは設定されたワークスペース内のファイルにのみアクセスし、コマンドを実行できます。

#### デフォルト設定

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

| オプション | Default | 説明 |
|--------|---------|-------------|
| `workspace` | `~/.v1claw/workspace` | エージェントの作業ディレクトリ |
| `restrict_to_workspace` | `true` | ファイル/コマンドアクセスをワークスペースに制限 |

#### 保護されたツール

`restrict_to_workspace: true`の場合、以下のツールがサンドボックス化されます：

| Tool | 機能 | 制限 |
|------|----------|-------------|
| `read_file` | ファイル読み取り | ワークスペース内のファイルのみ |
| `write_file` | ファイル書き込み | ワークスペース内のファイルのみ |
| `list_dir` | ディレクトリ一覧 | ワークスペース内のディレクトリのみ |
| `edit_file` | ファイル編集 | ワークスペース内のファイルのみ |
| `append_file` | ファイル追記 | ワークスペース内のファイルのみ |
| `exec` | コマンド実行 | コマンドパスはワークスペース内である必要あり |

#### 追加のExec保護

`restrict_to_workspace: false`でも、`exec`ツールは以下の危険なコマンドをブロックします：

* `rm -rf`, `del /f`, `rmdir /s` — 一括削除
* `format`, `mkfs`, `diskpart` — ディスクフォーマット
* `dd if=` — ディスクイメージング
* Writing to `/dev/sd[a-z]` — 直接ディスク書き込み
* `shutdown`, `reboot`, `poweroff` — システムシャットダウン
* フォーク爆弾 `:(){ :|:& };:`

#### エラー例

```
[ERROR] tool: Tool execution failed
{tool=exec, error=Command blocked by safety guard (path outside working dir)}
```

```
[ERROR] tool: Tool execution failed
{tool=exec, error=Command blocked by safety guard (dangerous pattern detected)}
```

#### 制限の無効化（セキュリティリスク）

ワークスペース外のパスにアクセスする必要がある場合：

**方法1: 設定ファイル**

```json
{
  "agents": {
    "defaults": {
      "restrict_to_workspace": false
    }
  }
}
```

**方法2: 環境変数**

```bash
export V1CLAW_AGENTS_DEFAULTS_RESTRICT_TO_WORKSPACE=false
```

> ⚠️ **警告**: この制限を無効にすると、エージェントがシステム上の任意のパスにアクセスできるようになります。管理された環境でのみ注意して使用してください。

#### セキュリティ境界の一貫性

`restrict_to_workspace`設定はすべての実行パスに一貫して適用されます：

| 実行パス | セキュリティ境界 |
|----------------|-------------------|
| Main Agent | `restrict_to_workspace` ✅ |
| Subagent / Spawn | 同じ制限を継承 ✅ |
| Heartbeat tasks | 同じ制限を継承 ✅ |

すべてのパスが同じワークスペース制限を共有します。サブエージェントやスケジュールタスクを通じてセキュリティ境界をバイパスする方法はありません。

### ハートビート（定期タスク）

V1Clawは定期タスクを自動的に実行できます。ワークスペースに`HEARTBEAT.md`ファイルを作成してください：

```markdown
# Periodic Tasks

- Check my email for important messages
- Review my calendar for upcoming events
- Check the weather forecast
```

エージェントはこのファイルを30分ごと（設定可能）に読み取り、利用可能なツールを使用してタスクを実行します。

#### Spawnによる非同期タスク

長時間実行タスク（Web検索、API呼び出し）には、`spawn`ツールを使用して**サブエージェント**を作成します：

```markdown
# Periodic Tasks

## Quick Tasks (respond directly)
- Report current time

## Long Tasks (use spawn for async)
- Search the web for AI news and summarize
- Check email and report important messages
```

**主な動作：**

| Feature | 説明 |
|---------|-------------|
| **spawn** | 非同期サブエージェントを作成、ハートビートをブロックしない |
| **Independent context** | サブエージェントは独自のコンテキストを持ち、セッション履歴なし |
| **message tool** | サブエージェントはメッセージツールを通じてユーザーと直接通信 |
| **Non-blocking** | スポーン後、ハートビートは次のタスクに続行 |

#### サブエージェント通信の仕組み

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

サブエージェントはツール（message、web_searchなど）にアクセスでき、メインエージェントを経由せずにユーザーと独立して通信できます。

**設定：**

```json
{
  "heartbeat": {
    "enabled": true,
    "interval": 30
  }
}
```

| オプション | Default | 説明 |
|--------|---------|-------------|
| `enabled` | `true` | ハートビートの有効/無効 |
| `interval` | `30` | チェック間隔（分）（最小: 5） |

**環境変数：**

* `V1CLAW_HEARTBEAT_ENABLED=false` 無効化
* `V1CLAW_HEARTBEAT_INTERVAL=60` 間隔変更

### プロバイダー

> [!NOTE]
> Groqは Whisperを通じて無料の音声文字起こしを提供しています。設定すれば、Telegramの音声メッセージが自動的に文字起こしされます。

| Provider                   | 用途                                 | APIキー取得                                            |
| -------------------------- | --------------------------------------- | ------------------------------------------------------ |
| `gemini`                   | LLM（Geminiダイレクト）                     | [aistudio.google.com](https://aistudio.google.com)     |
| `zhipu`                    | LLM（Zhipuダイレクト）                      | [bigmodel.cn](bigmodel.cn)                             |
| `openrouter(To be tested)` | LLM（推奨、全モデルアクセス） | [openrouter.ai](https://openrouter.ai)                 |
| `anthropic(To be tested)`  | LLM（Claudeダイレクト）                     | [console.anthropic.com](https://console.anthropic.com) |
| `openai(To be tested)`     | LLM（GPTダイレクト）                        | [platform.openai.com](https://platform.openai.com)     |
| `deepseek(To be tested)`   | LLM（DeepSeekダイレクト）                   | [platform.deepseek.com](https://platform.deepseek.com) |
| `groq`                     | LLM + **音声文字起こし**（Whisper） | [console.groq.com](https://console.groq.com)           |

<details>
<summary><b>Zhipu</b></summary>

**1. APIキーとベースURLを取得**

* Get [API key](https://bigmodel.cn/usercenter/proj-mgmt/apikeys)

**2. 設定**

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

**3. 実行**

```bash
v1claw agent -m "Hello"
```

</details>

<details>
<summary><b>完全な設定例</b></summary>

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

## CLIリファレンス

| Command                   | 説明                   |
| ------------------------- | ----------------------------- |
| `v1claw onboard`        | 設定とワークスペースの初期化 |
| `v1claw agent -m "..."` | エージェントとチャット           |
| `v1claw agent`          | インタラクティブチャットモード         |
| `v1claw gateway`        | ゲートウェイを起動             |
| `v1claw status`         | ステータスを表示                   |
| `v1claw cron list`      | スケジュールされたジョブを一覧表示       |
| `v1claw cron add ...`   | スケジュールジョブを追加           |

### スケジュールタスク / リマインダー

V1Clawは`cron`ツールを通じてスケジュールリマインダーと繰り返しタスクをサポートします：

* **One-time reminders**: 「10分後にリマインド」→ 10分後に1回トリガー
* **Recurring tasks**: 「2時間ごとにリマインド」→ 2時間ごとにトリガー
* **Cron expressions**: 「毎日9時にリマインド」→ cron式を使用

ジョブは`~/.v1claw/workspace/cron/`に保存され、自動的に処理されます。

## 🤝 貢献＆ロードマップ

PR歓迎！コードベースは意図的に小さく読みやすくしています。🤗

ロードマップは近日公開...

開発者グループ構築中、参加要件：最低1つのマージ済みPR。

ユーザーグループ：

discord:  <https://discord.gg/V4sAZ9XWpN>

<img src="assets/wechat.png" alt="V1Claw" width="512">

## 🐛 トラブルシューティング

### Web検索で「API 配置问题」と表示される

これは検索APIキーをまだ設定していない場合の正常な動作です。V1Clawは手動検索用のリンクを提供します。

Web検索を有効にするには：

1. **オプション 1 （推奨）**: Get a free API key at [https://brave.com/search/api](https://brave.com/search/api) （月2000回無料）で取得して最良の結果を得る。
2. **オプション 2 (No Credit Card)**: If you don't have a key, we automatically fall back to **DuckDuckGo** (no key required).

Add the key to `~/.v1claw/config.json` Braveを使用する場合：

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

### コンテンツフィルタリングエラーが発生する

一部のプロバイダー（Zhipuなど）にはコンテンツフィルタリングがあります。クエリを言い換えるか、別のモデルを使用してください。

### Telegramボットで「Conflict: terminated by other getUpdates」と表示される

これはボットの別のインスタンスが実行中の場合に発生します。`v1claw gateway`が1つだけ実行されていることを確認してください。

---

## 📝 APIキー比較

| Service          | 無料枠           | ユースケース                              |
| ---------------- | ------------------- | ------------------------------------- |
| **OpenRouter**   | 月200Kトークン   | 複数モデル（Claude、GPT-4など） |
| **Zhipu**        | 月200Kトークン   | 中国ユーザーに最適                |
| **Brave Search** | 月2000クエリ  | Web検索機能              |
| **Groq**         | 無料枠あり | 高速推論（Llama、Mixtral）       |
