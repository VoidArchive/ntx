# NTX - NEPSE Power Terminal

A terminal-based portfolio management and analysis tool for the Nepal Stock Exchange (NEPSE). Built with Go and Bubbletea for fast, keyboard-driven financial analysis and portfolio tracking.

## Phase 1 Foundation - Complete ✅

The project foundation has been successfully implemented with:

- ✅ Go module initialized with proper naming
- ✅ Complete folder structure following Go standard project layout
- ✅ Basic main.go with proper error handling and application lifecycle
- ✅ Configuration management (config.toml + encrypted credentials)
- ✅ Structured logging with slog
- ✅ Simple Bubbletea TUI foundation ready for Phase 2

## Quick Start

### Build and Run

```bash
# Build the application
go build -o bin/ntx ./cmd/ntx

# Run the application
./bin/ntx
```

### Configuration

Configuration files are automatically created in `~/.ntx/`:
- `config.toml` - User-editable settings (see `configs/config.toml` for template)
- `credentials` - Encrypted application secrets (auto-generated)

## Architecture

### Project Structure
```
ntx/
├── cmd/ntx/                  # Application entry point
├── internal/                 # Private business logic
│   ├── app/                 # Application orchestration & lifecycle
│   ├── ui/                  # Bubbletea components & views
│   ├── data/                # Repository pattern & SQLite
│   ├── market/              # Web scraping & data parsing
│   ├── analysis/            # Financial calculations
│   └── security/            # Encryption & validation
├── configs/                 # Configuration templates
└── requirements/            # Generated requirement specs
```

### Security Features
- AES-256 encryption for sensitive data
- Secure credential management
- Input validation and sanitization
- Proper file permissions (0600 for credentials, 0644 for config)

## Next Steps (Phase 2)

The foundation is ready for implementing:
- Full configuration loading (TOML parsing with Viper)
- Portfolio management features
- NEPSE data scraping with respectful rate limiting
- Multi-pane dashboard layout
- SQLite database with application-layer encryption

## Development

This is a learning project focused on:
- Go standard library and best practices
- Bubbletea TUI development patterns
- Secure application architecture
- Financial domain modeling
- Web scraping ethics and techniques
