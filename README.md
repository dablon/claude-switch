# Claude Switch

CLI tool to easily switch between different Claude Code profiles (API providers).

## Installation

```bash
# Download the latest release for your platform
# Linux/macOS
curl -L -o claude-switch https://github.com/dablon/claude-switch/releases/latest/download/claude-switch-linux-amd64
chmod +x claude-switch

# Windows
# Download claude-switch.exe from releases
```

Or build from source:

```bash
git clone https://github.com/dablon/claude-switch
cd claude-switch
go build -o claude-switch .
```

## Usage

### Add a profile

```bash
claude-switch add \
  --name claude-pro \
  --provider anthropic \
  --model claude-opus-4-6 \
  --key sk-ant-your-api-key

claude-switch add \
  --name minimax \
  --provider minimax \
  --model minimax-M2.5 \
  --key your-minimax-key

claude-switch add \
  --name openai \
  --provider openai \
  --model gpt-5 \
  --key sk-your-openai-key
```

### List profiles

```bash
claude-switch list
```

### Switch profile

```bash
claude-switch use --name minimax
```

### Show current profile

```bash
claude-switch show
```

### Export environment variables

```bash
# Export for current shell
eval $(claude-switch export)

# Or just print them
claude-switch export
```

### Apply and see instructions

```bash
claude-switch apply --name minimax
```

### Remove a profile

```bash
claude-switch remove --name old-profile
```

## Supported Providers

- `anthropic` / `claude` - Anthropic Claude API
- `openai` / `github-copilot` - OpenAI / GitHub Copilot
- `minimax` - Minimax API
- Any other provider (generic export)

## Example Workflow

```bash
# 1. Add your profiles
claude-switch add --name claude-pro --provider anthropic --model claude-opus-4-6 --key sk-ant-xxx
claude-switch add --name minimax --provider minimax --model minimax-M2.5 --key mm-xxx

# 2. Switch to minimax when you hit Claude Pro limit
claude-switch use --name minimax

# 3. Apply the environment variables
eval $(claude-switch export)

# 4. Now Claude Code will use minimax!
```

## Config Location

Profiles are stored in: `~/.claude-switch/config.json`

## License

MIT
