# NTX - NEPSE Power Terminal

A fast and intuitive portfolio management terminal application for the Nepal Stock Exchange (NEPSE). Built with Go, NTX helps you track your investments, manage transactions, and analyze portfolio performance directly from your terminal.

![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)
![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macOS%20%7C%20windows-lightgrey.svg)

## Features

### Terminal Interface
- **Multiple Themes**: Choose from Tokyo Night, Rose Pine, Gruvbox, or the default custom theme
- **Responsive Layout**: The interface adapts to your terminal size automatically
- **Keyboard Navigation**: Vim-style shortcuts for efficient navigation

### Portfolio Management
- **Real-time Holdings**: See your portfolio overview with live P/L calculations
- **Transaction Management**: Track all your buy and sell transactions
- **Multiple Portfolios**: Manage several portfolios simultaneously
- **Financial Accuracy**: Uses integer-based calculations to avoid floating-point errors

### Analytics and Insights
- **Performance Metrics**: Track unrealized/realized gains, total returns, and identify your best and worst performers
- **Sector Analysis**: View portfolio allocation and get diversification insights
- **Risk Assessment**: Understand your portfolio's risk profile

### Data Management
- **Meroshare Import**: Import your portfolio directly from Meroshare CSV exports
- **Transaction History**: Comprehensive filtering and search through your trading history
- **Backup System**: Built-in database backup and restore functionality

### Technical Features
- **Type-safe Database**: Uses SQLC for compile-time SQL safety
- **Schema Migrations**: Version-controlled database schema with Goose
- **Configuration System**: Supports CLI flags, environment variables, and config files
- **Cross-platform**: Pure Go implementation that works everywhere

## Getting Started

### What You Need
- Go 1.24 or newer
- A terminal that supports 256 colors (most modern terminals do)

### Installation

```bash
# Clone the repository
git clone https://github.com/your-username/ntx.git
cd ntx

# Build the application
go build -o bin/ntx

# Set up the database
./bin/ntx db init

# Start using NTX
./bin/ntx
```

You can also install directly from source:
```bash
go install github.com/your-username/ntx@latest
```

## How to Use NTX

### Navigation Basics
- **Numbers 1-5**: Jump to different sections (Overview, Holdings, Analysis, History, Market)
- **hjkl or Arrow keys**: Move around within sections
- **Tab/Shift+Tab**: Cycle through sections
- **t**: Switch themes
- **?**: Show help and all available shortcuts
- **q**: Exit the application

### Managing Your Portfolio
1. **Start with a Portfolio**: Go to the Overview section to create your first portfolio
2. **Add Transactions**: Head to the History section and press 'n' to add a new transaction
3. **Import Existing Data**: In the Holdings section, press 'i' to import from Meroshare CSV
4. **Analyze Performance**: Check out the Analysis section for detailed insights

### Configuration
NTX uses a simple configuration hierarchy. Create a config file at `~/.config/ntx/config.toml`:

```toml
[ui]
theme = "tokyo_night"
default_section = "holdings"

[display]
refresh_interval = 30
currency_symbol = "Rs."

[database]
backup_on_exit = true
```

You can override any setting with environment variables:
```bash
export NTX_UI_THEME="rose_pine"
export NTX_DATABASE_PATH="/custom/path/portfolio.db"
```

## Project Structure

```
ntx/
├── cmd/                     # Application entry points
├── internal/
│   ├── app/                 # Main application logic
│   ├── config/              # Configuration management
│   ├── data/                # Database layer
│   │   ├── migrations/      # SQL migrations
│   │   ├── queries/         # SQLC queries
│   │   └── repository/      # Repository pattern
│   ├── database/            # Generated SQLC code
│   ├── portfolio/           # Portfolio business logic
│   │   ├── models/          # Domain models
│   │   └── services/        # Business services
│   └── ui/                  # TUI components
│       ├── components/      # Reusable UI components
│       └── themes/          # Theme system
├── bin/                     # Compiled binaries
└── requirements/            # Project documentation
```

## Development Setup

### Getting the Development Environment Ready

```bash
# Get the code
git clone https://github.com/your-username/ntx.git
cd ntx

# Install dependencies
go mod download

# Install development tools
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/kyleconroy/sqlc/cmd/sqlc@latest

# Run the tests
go test ./...

# Regenerate SQLC code after changing queries
sqlc generate

# Create a new database migration
goose -dir internal/data/migrations create add_new_table sql
```

### Working with the Database
```bash
# Set up a fresh database
go run . db init

# Apply migrations
go run . db migrate

# Check what migrations have been applied
go run . db status

# Create a backup
go run . db backup
```

### Building and Testing
```bash
# Build for your current system
go build -o bin/ntx

# Cross-compile for other platforms
./scripts/build.sh

# Run with race detection enabled
go run -race .

# Get test coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Contributing

We'd love your help making NTX better! Whether you're fixing bugs, adding features, improving docs, or just sharing ideas, contributions are welcome.

### How to Contribute

1. **Fork and Clone**
   ```bash
   # Fork the repo on GitHub, then:
   git clone https://github.com/your-username/ntx.git
   cd ntx
   ```

2. **Create a Branch**
   ```bash
   git checkout -b feature/your-feature-name
   # or for bug fixes:
   git checkout -b fix/issue-description
   ```

3. **Make Your Changes**
   - Follow Go best practices and the existing code style
   - Add tests for new functionality
   - Update documentation when needed
   - Make sure your code is well-commented

4. **Test Everything**
   ```bash
   # Run all tests
   go test ./...
   
   # If you have golangci-lint installed
   golangci-lint run
   
   # Actually try the app
   go run .
   ```

5. **Commit and Push**
   ```bash
   git add .
   git commit -m "feat: describe what you added"
   # or for fixes:
   git commit -m "fix: describe what you fixed"
   git push origin feature/your-feature-name
   ```

6. **Open a Pull Request**
   - Target the `main` branch
   - Write a clear description of what you changed and why
   - Reference any related issues

### Code Guidelines

**Go Standards**
- Use Go 1.24+ features when they make sense
- Format with `gofmt` and `goimports`
- Follow the project's commenting style:
  - Start each file with a comment explaining its purpose
  - Document all functions with their inputs and outputs
  - Add inline comments for complex logic
- Handle errors properly with context
- Write tests for new features

**Commit Messages**
We use [Conventional Commits](https://www.conventionalcommits.org/):
- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `refactor:` for code improvements
- `test:` for adding tests
- `chore:` for maintenance

**Database Changes**
- Always create migrations for schema changes
- Update SQLC queries if you modify them
- Test both up and down migrations
- Document any breaking changes

**UI Changes**
- Test on different terminal sizes
- Make sure all themes still work properly
- Keep keyboard navigation consistent
- Follow the existing design patterns

### What to Work On

**Good Starting Points**
- Improve existing themes or create new ones
- Write more documentation
- Add unit tests
- Small UI improvements
- Add new configuration options

**High Impact Work**
- Real-time price integration
- Better portfolio analytics
- Export features
- Performance improvements
- Better mobile terminal support

**Advanced Projects**
- Plugin system for custom features
- Integration with NEPSE or other financial APIs
- Machine learning for portfolio insights
- Multiple language support
- Cloud synchronization

### Reporting Problems

When you find a bug, please include:
- **Your setup**: Operating system, terminal, Go version
- **How to reproduce**: Step-by-step instructions
- **What should happen**: Expected behavior
- **What actually happens**: Actual behavior
- **Screenshots**: Especially helpful for UI issues
- **Error messages**: Any logs or error output

### Getting Help

- **GitHub Discussions**: For questions and brainstorming
- **Issues**: For specific bugs and feature requests
- **Code Reviews**: We review all pull requests thoroughly

### Recognition

Contributors are recognized in the project's `CONTRIBUTORS.md` file and in release notes for significant contributions.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Support

If you find NTX useful:
- Star the repository to show your support
- Report bugs and suggest features
- Contribute code or documentation
- Tell other NEPSE investors about it

---

Built for the NEPSE trading community. Happy investing!

