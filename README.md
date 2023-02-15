# elon-bot
## Elon Musk AI Replica Discord Chat Bot.

## Simply add it to your server: http://discord.me/elonmusk
Join [73labs Discord](https://discord.gg/YMdtBaSj)

- Based on OpenAI's GPT-3 Davinvi language model.
- Uses the OpenAI completion API to get the responses.
- Keeps conversation context in sessions.

## How to run it?

### Step 1.
Clone the repository wherever you like.
Follow the installation instructions of GO for your system on https://go.dev.

### Step 2.
Setup a new bot on https://discord.com/developers/applications.
  - Bot -> Generate Bot Token
 
Sign up for a new account on https://openai.com
  - Profile Icon -> View API Keys -> Generate API Keys

### Step 3.
Create .env file in the project root. Replace the placeholders with your tokens and save.
```env
DISCORD_BOT_TOKEN="<YOUR_BOT_TOKEN>"
OPENAI_API_KEY="<YOUR_API_KEY>"
```

### Step 4.
From the project root run
```
go run cmd/chatbot/main.go
