# Claude Switch ⏻

CLI tool to easily switch between different Claude Code / AI provider profiles. Perfect when you hit rate limits on one provider and need to switch to another.

![Coverage](https://img.shields.io/badge/coverage-97.5%25-brightgreen)
![License](https://img.shields.io/badge/license-MIT-blue)
![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)

## ✨ Features

- **Auto-detect provider** from profile name (`minimax` → Minimax, `openai-prod` → OpenAI, etc.)
- **Quick add** - name your profile after the provider and it auto-detects everything
- **Multiple providers** - Anthropic, OpenAI, GitHub Copilot, Minimax, Azure, Vertex
- **Easy switch** - One command to switch profiles
- **Export env vars** - Ready for Claude Code or any AI tool
- **Docker support** - Run as container

## 🚀 Quick Start

```bash
# Quick add - reads API key from env automatically
# (reads ANTHROPIC_API_KEY, MINIMAX_API_KEY, OPENAI_API_KEY, etc.)
claude-switch quick --name anthropic
claude-switch quick --name minimax

# Or pass key explicitly
claude-switch quick --name openai --key sk-xxx

# Or add with explicit provider (when name doesn't match a provider)
claude-switch add --name my-custom --key xxx --provider minimax

# Switch to a profile
claude-switch use --name minimax

# Export env vars
eval $(claude-switch export)
```

## 📦 Installation

### From Release

```bash
# Linux
curl -L -o claude-switch https://github.com/dablon/claude-switch/releases/latest/download/claude-switch-linux-amd64
chmod +x claude-switch

# Windows
# Download claude-switch-windows-amd64.exe
```

### From Source

```bash
git clone https://github.com/dablon/claude-switch
cd claude-switch
go build -o claude-switch ./cmd/claude-switch
```

### Docker

```bash
docker build -t claude-switch .
docker run --rm -it -v ~/.claude-switch:/root/.claude-switch claude-switch --help
```

## 📖 Commands

| Command | Description |
|---------|-------------|
| `claude-switch quick -n <name> -k <key>` | Quick add with auto-detection |
| `claude-switch add -n <name> -k <key> -p <provider> -m <model>` | Add profile |
| `claude-switch list` | List all profiles |
| `claude-switch use -n <name>` | Switch to profile |
| `claude-switch show` | Show current profile |
| `claude-switch export` | Export env vars |
| `claude-switch detect -k <name>` | Detect provider from name |
| `claude-switch providers` | List supported providers |
| `claude-switch remove -n <name>` | Remove profile |

### Options

```bash
-n, --name      Profile name (required for add/quick/use/remove)
-k, --key       API key (required for add/quick)
-p, --provider  Provider: anthropic, openai, minimax, azure, vertex
-m, --model     Model name (auto-detected if omitted)
-e, --endpoint  Custom endpoint (optional)
```

## 🔍 Auto-Detection

The CLI automatically detects the provider from your **profile name**. If the name contains a known provider, it's auto-assigned:

| Profile Name | Detected Provider |
|-------------|-------------------|
| `anthropic`, `my-anthropic` | Anthropic |
| `openai`, `openai-prod` | OpenAI |
| `minimax`, `minimax-test` | Minimax |
| `azure`, `azure-eastus` | Azure |
| `vertex`, `vertex-us` | Vertex |
| `github-copilot` | GitHub Copilot |
| anything else | Custom |

If the name doesn't match a provider, use `--provider` explicitly with `add`.

## 🏗️ Project Structure

```
claude-switch/
├── cmd/
│   └── claude-switch/
│       └── main.go           # CLI entry point
├── internal/
│   ├── config/               # Config management
│   │   ├── config.go
│   │   └── config_test.go
│   └── provider/            # Provider detection
│       ├── provider.go
│       └── provider_test.go
├── tests/
│   └── e2e/                # E2E tests
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
```

## 🧪 Testing

```bash
# Run unit tests
go test -v ./internal/...

# Run E2E tests
go test -v ./tests/e2e/...

# Run all tests with coverage
go test -cover ./...
```

### Coverage

| Package | Coverage |
|---------|----------|
| internal/config | 98.5% |
| internal/provider | 96.1% |
| **Total** | **97.5%** |

## 🔄 Example Workflow

```bash
# 1. Set your env vars (or add them to your shell profile)
export ANTHROPIC_API_KEY=sk-ant-api03-xxx
export MINIMAX_API_KEY=sk-cp-your_key

# 2. Quick add profiles (keys read from env automatically)
claude-switch quick --name anthropic
claude-switch quick --name minimax

# 3. When Anthropic hits limit...
claude-switch use --name minimax

# 4. Apply the new profile
eval $(claude-switch export)

# 5. Claude Code now uses Minimax!
```

## ⚙️ Config

Profiles are stored in: `~/.claude-switch/config.json`

```json
{
  "profiles": [
    {
      "name": "claude-pro",
      "provider": "anthropic",
      "model": "claude-opus-4-6",
      "api_key": "sk-ant-xxx"
    }
  ],
  "current_profile": "claude-pro"
}
```

## 🐳 Docker Usage

```bash
# Build
docker build -t claude-switch .

# Add profile
docker run --rm -it -v $HOME/.claude-switch:/root/.claude-switch claude-switch quick -n minimax -k sk-cp-xxx

# List profiles
docker run --rm -it -v $HOME/.claude-switch:/root/.claude-switch claude-switch list
```

## 📝 License

MIT - See [LICENSE](LICENSE) for details.

---

Made with ❤️ for developers
