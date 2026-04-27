# ReadBud

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](./LICENSE)

[中文](./README.zh-CN.md) | English

ReadBud is an end-to-end content automation tool for WeChat Official Accounts. Given a keyword or topic, it researches the web, drafts an article with an LLM, generates a cover image, applies a typographic style, and ships the result to a WeChat draft — either through the official Draft API or through a companion browser extension that fills in the WeChat web editor.

## Features

- **Research → Draft → Cover → Publish** in a single pipeline
- **Three publish paths** to fit different account types:
  - **API mode** — direct delivery via WeChat Draft API (requires a verified service account)
  - **Extension mode** — a Chrome extension auto-fills the WeChat web editor (works with personal accounts)
  - **Manual mode** — packaged content for copy-paste delivery
- **Pluggable providers** for LLMs, image generation, web search, and crawling — switch vendors from the settings UI
- **Three built-in typographic styles** (minimal, magazine, stitch), each with a matching cover-image prompt set
- **Brand and style profiles** so multiple accounts can share or diverge in tone and look
- **Draft version history** with rollback
- **Pipeline observability** — every stage is a tracked job with retries and timing

## Architecture

- **Backend** — Go 1.26, Gin, GORM, Asynq, PostgreSQL, Redis
- **Frontend** — Vue 3, Pinia, Element Plus, Vite
- **Browser extension** — Chrome Manifest V3 (vanilla JS)
- **Provider integrations** — adapter layer for LLM / image / search / crawler vendors, configurable per workspace

## Quick Start

Requires Docker and Docker Compose.

```bash
git clone https://github.com/Runcoor/ReadBud.git
cd ReadBud
cp backend/configs/config.example.yaml backend/configs/config.yaml
# edit backend/configs/config.yaml
./start.sh
```

Once the stack is up:

- Frontend: <http://localhost:19880>
- API: <http://localhost:19881>

### First-run checklist

1. Register an account (open registration is on by default — disable it before going to production).
2. **Settings → Providers** — add API keys for your LLM, search, and image-generation services.
3. **Settings → WeChat Accounts** — bind your AppID and AppSecret, then choose a delivery mode (API / Extension / Manual).
4. **Settings → Extension** — issue a token and install the bundled extension from `wechat-extension/` via `chrome://extensions` (developer mode → Load unpacked).

## Typographic Styles

| Style    | Palette                          | Feel                       |
| -------- | -------------------------------- | -------------------------- |
| minimal  | Black, white, a hint of yellow   | Swiss / editorial          |
| magazine | Off-white with red accents       | Print magazine             |
| stitch   | Cream and burnt orange           | Handcrafted / vintage      |

Cover images are generated per style — prompts and palettes stay consistent between cover and body.

## Roadmap

- Round-trip the published article URL after a successful WeChat send
- Wire up the analytics dashboard (UI exists, data plumbing pending)
- Verify the extension on Edge and Firefox (currently tested on Chrome)

## License

Licensed under the [GNU AGPL-3.0](./LICENSE).
