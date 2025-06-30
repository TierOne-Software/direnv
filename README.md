# direnv - Directory Environment Manager

Automatically setup your shell environment based on the current directory. 

## Features

- 🚀 **Auto-apply**: Automatically load environment when entering directories (per-session control)
- 🔧 **Rich configuration**: Environment variables, aliases, embedded scripts, and hooks
- 📁 **Hierarchical**: Searches parent directories for configuration
- 🔄 **Multi-shell safe**: Per-process state management prevents conflicts between terminals
- 🐚 **Shell support**: Native integration for bash and zsh with completion
- 📝 **TOML format**: Clean, commented configuration files
- 🎯 **Local overrides**: Personal `.direnv.local.toml` for developer-specific settings
- 📊 **Environment diff**: Preview changes before applying
- 🩺 **Diagnostics**: Built-in doctor command for troubleshooting
- 🪝 **Hooks**: Pre-apply, post-apply, and on-leave automation

## Installation

```bash
go install github.com/TierOne-Software/direnv@latest
```

Or build from source:

```bash
git clone https://github.com/TierOne-Software/direnv
cd direnv
go build -o direnv
sudo mv direnv /usr/local/bin/
```

## Quick Start

1. Initialize shell integration:
   ```bash
   # Automatically detects your shell and outputs the appropriate script
   direnv init >> ~/.bashrc   # if using bash
   direnv init >> ~/.zshrc    # if using zsh
   ```

2. Reload your shell:
   ```bash
   source ~/.bashrc  # or ~/.zshrc
   ```

3. Create a `.direnv.toml` in your project:
   ```toml
   auto_apply = true

   [environment]
   PROJECT_NAME = "myproject"
   NODE_ENV = "development"

   [aliases]
   build = "npm run build"
   test = "npm test"
   ```

4. Enter the directory and watch the magic happen!

## Usage

### Commands

- `direnv apply` - Output shell commands to apply the environment (use with eval)
- `direnv diff` - Show what changes would be applied
- `direnv info` - Show current status and configuration
- `direnv enable` - Enable auto-apply globally
- `direnv disable` - Disable auto-apply globally
- `direnv init` - Print shell integration script
- `direnv completion` - Generate shell completions
- `direnv doctor` - Diagnose configuration issues
- `direnv restore` - Restore the previous environment
- `direnv run <script>` - Run a script defined in the configuration

### Shell Functions

After shell integration, you'll have these functions available:
- `direnv-apply` - Apply the current directory's environment
- `direnv-restore` - Restore the previous environment
- `direnv-info` - Show current status
- `direnv-enable` - Enable auto-apply
- `direnv-disable` - Disable auto-apply

### Configuration Format

Create a `.direnv.toml` file in your project directory:

```toml
# Enable automatic loading for this directory
auto_apply = true

# Environment variables
[environment]
CC = "gcc-11"
PATH = "$PATH:$PROJECT_ROOT/bin"  # $PROJECT_ROOT is the config directory

# Shell aliases
[aliases]
ll = "ls -la"
gs = "git status"

# Embedded scripts - these become shell functions you can call directly!
[scripts]
setup = """
echo "Setting up project..."
npm install
"""

# After applying, you can just run: setup
```

### Environment Variable Expansion

- `$PROJECT_ROOT` - Expands to the directory containing `.direnv.toml`
- Standard variables like `$PATH`, `$HOME` are expanded
- Use single quotes to prevent expansion

## Advanced Features

### Local Overrides

Create a `.direnv.local.toml` file for personal settings that override team configuration:

```toml
# .direnv.local.toml (add to .gitignore)
[environment]
DATABASE_URL = "postgresql://localhost:5432/my_local_db"
EDITOR = "nvim"  # Override team's vim setting

[aliases]
test = "go test -v ./..."  # Add verbose flag to team's test alias
```

### Hooks

Automate tasks at specific points in the environment lifecycle:

```toml
[hooks]
pre_apply = """
echo "Loading environment for $PROJECT_NAME..."
"""

post_apply = """
echo "Environment ready! Available commands:"
echo "  build - Build the project"
echo "  test  - Run tests"
"""

on_leave = """
echo "Leaving $PROJECT_NAME environment"
rm -rf tmp/*.log 2>/dev/null || true
"""
```

### Auto-Apply Control

Enable auto-apply per shell session:

```bash
# Enable for current shell session
export DIRENV_AUTO_APPLY=1

# Disable for a specific command
DIRENV_AUTO_APPLY=0 some-command

# Add to shell config for permanent enable
echo 'export DIRENV_AUTO_APPLY=1' >> ~/.zshrc
```

### Multi-Terminal Safety

Each shell session maintains independent state:
- State files stored in: `~/.config/direnv/state_12345.json`
- No conflicts between multiple terminals
- Automatic cleanup of orphaned state files

### Diagnostics

```bash
# Check configuration and setup
direnv doctor

# Preview changes before applying
direnv diff

# See detailed status
direnv info

# Clean up orphaned state files
direnv cleanup
```

### Examples

See the `example/` directory for real-world configuration examples:
- Basic project setup
- Node.js development  
- Go development
- Linux kernel cross-compilation

## How It Works

1. When you `cd` into a directory, direnv checks for `.direnv.toml`
2. If found and `auto_apply` is true, it:
   - Exports environment variables with expansion
   - Creates shell aliases for quick commands
   - **Defines shell functions from scripts that you can call directly**
3. Scripts become first-class commands in your shell:
   ```bash
   # Instead of: direnv run build
   # Just type: build
   
   # Instead of: direnv run test  
   # Just type: test
   ```

## Testing

Run the test suite:

```bash
go test ./...

# Run integration tests
go test -v ./... -run Integration
```

## License

Licensed under the Apache License, Version 2.0. See [LICENSE-2.0.txt](LICENSE-2.0.txt) for the full license text.

## Contributing

We welcome contributions to TierOne direnv! Here's how you can help:

### Submitting Changes

1. Fork the repository on GitHub
2. Create a feature branch from `master`
3. Make your changes with appropriate tests
4. Ensure all tests pass and code follows the existing style
5. Submit a pull request with a clear description of your changes

### Development Guidelines

- Follow the existing code style and naming conventions
- Add tests for new features or bug fixes
- Update documentation for changes

### Reporting Issues

Please report bugs, feature requests, or questions by opening an issue on GitHub.

## Copyright

Copyright 2025 TierOne Software. All rights reserved.