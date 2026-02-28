# LLM / multimodal model IDs for top providers (early 2026)

Names are taken from each provider’s public docs or reliable aggregators and focus on **base models** (not every fine‑tune). Where providers expose many variants only via authenticated `list models` APIs, you should still call those APIs in your own code for a truly exhaustive list.[web:88][web:89][web:110][web:118][web:121][web:122][web:134][web:136]

---

## OpenAI – main public models

From OpenAI’s "Models" catalog page.[web:88]

- gpt-5.2  
- gpt-5.2-pro  
- gpt-5  
- gpt-5-mini  
- gpt-5-nano  
- gpt-4.1  
- gpt-oss-120b  
- gpt-oss-20b  
- sora-2  
- sora-2-pro  
- o3-deep-research  
- o4-mini-deep-research  
- gpt-image-1.5  
- chatgpt-image-latest  
- gpt-image-1  
- gpt-image-1-mini  
- gpt-4o-mini-tts  
- gpt-4o-transcribe  
- gpt-4o-mini-transcribe  
- gpt-realtime  
- gpt-audio  
- gpt-realtime-mini  
- gpt-audio-mini  

> For *all* live IDs (including legacy models and fine‑tunes), you must still call `GET https://api.openai.com/v1/models` in your own environment.[web:99][web:103][web:105]

---

## Google DeepMind – Gemini

Based on Gemini API model docs and Vertex AI docs.[web:89][web:104][web:106][web:137][web:140][web:141][web:145][web:146]

Core Gemini 3 / 2.5 text & multimodal models:

- gemini-3.1-pro-preview  
- gemini-3.1-pro-preview-customtools  
- gemini-3-flash-preview  
- gemini-3-pro-preview (deprecated; shuts down March 2026)  
- gemini-2.5-pro  
- gemini-2.5-flash-preview-09-25 (deprecated)  

Image / specialized Gemini 3 models:

- gemini-3.1-flash-image-preview  

> The Gemini models page also lists families "Gemini 3", "Gemini 2.5 Flash", "Gemini 2.5 Flash‑Lite", etc., but not every concrete model ID is spelled out without calling `models.list`.[web:89][web:91][web:96][web:104][web:140]

---

## Anthropic – Claude

Anthropic’s public pages name model *families* (Opus, Sonnet, Haiku) and versions, but do not expose the full set of raw API IDs without logging into the dashboard or using cloud partners.[web:55][web:58][web:87][web:92][web:130]

Main current Claude 4‑series models (marketing names):

- Claude Opus 4.6  
- Claude Sonnet 4.6  
- Claude Haiku 4.5  

Older but still referenced families:

- Claude 4.1 Opus / Sonnet / Haiku  
- Claude 4 Opus / Sonnet / Haiku  
- Claude 3.7 / 3.5 / 3 families (Opus, Sonnet, Haiku)  

> For exact API IDs (for example, historical `claude-3-opus-20240229`‑style strings and the latest 4.x IDs) you’ll need to read Anthropic’s API docs or the specific cloud provider (Bedrock / Vertex) you are using.[web:87][web:100][web:125][web:130]

---

## Meta – Llama (high level)

Meta mostly distributes Llama as open(-weight) checkpoints; individual hosting providers choose their own IDs. Current families:[web:59][web:62][web:65][web:68][web:71]

- Llama 4 Behemoth (frontier teacher model)  
- Llama 4 deployable variants (provider‑specific IDs)  
- Llama 3.3 70B  
- Llama 3.2  
- Llama 3.1  
- Llama 3 (various sizes)  

> For concrete IDs you’ll use host‑specific strings (for example, `meta-llama/Meta-Llama-3-70B-Instruct` on Hugging Face, or provider‑specific labels in Bedrock, Azure, etc.).[web:132]

---

## Microsoft – Phi‑4 family

From Microsoft Phi‑4 announcements and docs.[web:60][web:63][web:66][web:69][web:72]

- phi-4-reasoning  
- phi-4-mini  
- phi-4-multimodal  

> Exact IDs differ slightly between Azure AI Studio, local ONNX/DirectML builds, and open‑weight checkpoints, but the above strings reflect the canonical naming.

---

## Amazon (AWS) – Nova models

From Bedrock and community docs.[web:64][web:67][web:109][web:115][web:118]

Core Nova model IDs used on Bedrock:

- amazon.nova-micro-v1:0  
- amazon.nova-lite-v1:0  
- amazon.nova-pro-v1:0  
- amazon.nova-premier-v1:0  
- amazon.nova-canvas-v1:0  
- amazon.nova-reel-v1:0  

> Some docs and blog posts also mention Nova 2 Lite / Pro / Omni; those map to the same or successor IDs above depending on region and rollout phase.[web:64]

---

## xAI – Grok models

From xAI/Grok provider docs and model lists.[web:122][web:124][web:127][web:129]

Chat / reasoning models:

- grok-4  
- grok-4-0709  
- grok-4-latest  
- grok-3  
- grok-3-latest  
- grok-3-mini  
- grok-2-1212  
- grok-beta  

Vision / image variants:

- grok-4-vision (provider‑specific)  
- grok-2-vision-latest  
- grok-2-vision-1212  
- grok-vision-beta  
- grok-2-image-1212  

> Exact availability differs between xAI’s own API and aggregators like OpenRouter or AI SDK; all of these IDs are live in at least one documented integration.[web:122][web:124][web:127][web:129]

---

## Mistral AI – selected model IDs

From Mistral’s official "Models" page and Bedrock docs.[web:110][web:113][web:116][web:119]

Frontier & generalist models (latest marketing names):

- mistral-large-3  
- mistral-medium-3.1  
- mistral-small-3.2  
- ministral-3-14b  
- ministral-3-8b  
- ministral-3-3b  
- magistral-medium-1.2  
- magistral-small-1.2  

Specialist models (selection):

- codestral (latest)  
- devstral-2  
- devstral-small-2  
- devstral-medium-1.0  
- devstral-small-1.1  
- mistral-embed  
- codestral-embed  
- mistral-ocr-3  
- mistral-ocr-2  
- voxtral-mini  
- voxtral-small  
- voxtral-mini-transcribe  
- voxtral-mini-transcribe-2  
- voxtral-mini-transcribe-realtime  
- mistral-moderation  

Important legacy / versioned IDs (still seen in many APIs):

- magistral-medium-2507  
- magistral-small-2507  
- magistral-medium-2506  
- magistral-small-2506  
- devstral-small-2505  
- mistral-small-2503  
- mistral-ocr-2503  
- mistral-saba-2502  
- mistral-small-2501  
- codestral-2501  
- ministral-3b-2410  
- ministral-8b-2410  
- mistral-small-2409  
- pixtral-12b-2409  
- mistral-large-2407  
- codestral-2405  
- mistral-small-2402  
- mistral-large-2402  
- mistral-medium-2312  
- open-mistral-7b  
- open-mixtral-8x22b  
- open-mixtral-8x7b  

---

## Cohere – Command family (high level)

Cohere’s public docs list many Command and Embed models; some are region/provider‑specific.[web:35][web:79][web:81][web:111][web:114][web:117][web:120]

Representative LLM IDs:

- command-a-03-2025  
- command-a-reasoning-08-2025  
- command-a-vision-08-2025  
- command-r-plus (command-r-plus-latest on some platforms)  
- command-r  
- command (legacy)  

Representative embedding / reranker model IDs (often used alongside LLMs):

- embed-english-v3.0  
- embed-multilingual-v3.0  
- rerank-english-v3.0  

> For the exact current list, Cohere exposes a `List Models` endpoint in their API.[web:114]

---

## Alibaba / Qwen – main API model IDs

From Qwen model lists on third‑party directories and Qwen’s own overview.[web:121][web:123][web:126][web:128][web:131]

General LLMs:

- qwen-max-latest  
- qwen-plus-latest  
- qwen-turbo-latest  
- qwen-long  

Multimodal (vision-language):

- qwen-vl-plus-latest  
- qwen-vl-max-latest  

Math and code specialists:

- qwen-math-turbo-latest  
- qwen-math-plus-latest  
- qwen-coder-turbo-latest  
- qwen-coder-plus-latest  

Research / reasoning:

- qwq-32b-preview  

Open‑weight / instruct variants (selection):

- qwen2.5-7b-instruct  
- qwen2.5-72b-instruct  
- qwen2.5-32b-instruct  
- qwen3-max  
- qwen3.5-plus  
- qwen3.5-turbo  
- qwen3-omni  

> Alibaba Cloud Model Studio also lists newer Qwen‑3.x and Qwen‑3.5 variants (for example `qwen-turbo-2025-04-28`) with date‑stamped IDs; you’d normally read those straight from their console or docs for exact naming.[web:128][web:131]
