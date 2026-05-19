![LLM Expense Tracker](resources/banner.svg)

<p align="center">
  <img alt="Go" src="https://img.shields.io/badge/Backend-Go-00ADD8?style=for-the-badge&logo=go&logoColor=white">
  <img alt="React" src="https://img.shields.io/badge/Frontend-React-61DAFB?style=for-the-badge&logo=react&logoColor=111827">
  <img alt="MySQL" src="https://img.shields.io/badge/Database-MySQL-4479A1?style=for-the-badge&logo=mysql&logoColor=white">
  <img alt="Telegram Bot" src="https://img.shields.io/badge/Bot-Telegram-26A5E4?style=for-the-badge&logo=telegram&logoColor=white">
  <img alt="Anthropic" src="https://img.shields.io/badge/LLM-Anthropic-D97706?style=for-the-badge">
  <img alt="OpenAI" src="https://img.shields.io/badge/LLM-OpenAI-412991?style=for-the-badge&logo=openai&logoColor=white">
</p>


An AI-powered personal finance tracker with expense parsing, wallets, budgets, loans[planned], transfers, receipts, web dashboard, and Telegram bot support.

**Describe an expense in plain text → AI extracts the transaction → wallet balance updates automatically.**

---

**Stack:** Go · MySQL · React · Telegram Bot API · Anthropic / OpenAI


> This is an early-stage project that I continue developing in my free time. More features, improvements, and bug fixes will be added gradually.



## Features

- **AI chat (web & Telegram)** — describe transactions in plain text on either interface; LLM extracts amount, category, merchant
- **Receipt parsing** — upload photo or PDF on web or send to bot; AI extracts items automatically
- **Dashboard** — monthly overview, category trends, budget tracking
- **Accounts & transfers** — cash, bank, credit card with balance tracking
- **Budgets** — per-category monthly limits with usage alerts
- **CSV export** — filtered expense export


## Getting Started

```bash
cp .env.example .env   # fill in DB, LLM, and Telegram credentials
make db-create         # create the MySQL database
make migrate-up        # run schema migrations
make dev               # start API (air) + React dev server in parallel
```

**Key `.env` values:**

| Key | Description |
|-----|-------------|
| `DB_*` | MySQL connection |
| `LLM_PROVIDER` | `anthropic` or `openai` |
| `ANTHROPIC_API_KEY` / `OPENAI_API_KEY` | LLM credentials |
| `TELEGRAM_BOT_TOKEN` / `TELEGRAM_CHAT_ID` | Bot credentials |

**Production build:**

```bash
make build             # compiles React into binary, builds Go server → bin/server
make migrate-up
./bin/server
```

## Architecture

```
cmd/
  server/               HTTP server entry point
  migrate/              DB migration runner
internal/
  config/               environment config
  domain/               shared entity types
  llm/                  LLM provider abstraction (Anthropic / OpenAI)
  repository/mysql/     data access layer
  service/
    chat.go             centralized AI decision engine (shared by all transports)
    account|budget|...  business logic per domain
  transport/
    http/
      handler/          REST API handlers
      middleware/       auth, logging
      router.go         route registration
    telegram/
      handler.go        bot polling & infrastructure
      chat.go           message → ChatProcessor → Telegram UI
  worker/               background jobs (budget alerts, scheduler)
  web/                  embedded React build (served by Go)
web/src/                React + Vite source
migrations/             SQL schema files
```

## Screenshots

### Web dashboard

![Web dashboard](resources/1.png)

### Transactions

![Transactions](resources/2.png)

<table>
  <tr>
    <td align="center" width="33%">
      <img src="resources/tg1.jpeg" width="95%"><br>
      <sub>Telegram bot expense flow</sub>
    </td>
    <td align="center" width="33%">
      <img src="resources/tg3.jpeg" width="95%"><br>
      <sub>Receipt image reading</sub>
    </td>
    <td align="center" width="33%">
      <img src="resources/tg2.jpeg" width="95%"><br>
      <sub>Telegram bot wallet summary</sub>
    </td>
  </tr>
</table>
