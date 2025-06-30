# NTX - NEPSE Terminal Exchange

A terminal-based portfolio management and analysis tool for the Nepal Stock Exchange (NEPSE). Built with Go and Bubbletea for fast, keyboard-driven financial analysis and portfolio tracking.

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
