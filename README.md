# MacDevTUI

A Terminal User Interface (TUI) application for setting up and reinstalling Mac development environments with consistent theming and configuration.

## Features

- Interactive terminal-based installation interface
- Homebrew package management with Brewfile support
- Shell environment setup (zsh, oh-my-posh, atuin, zoxide)
- Development tools installation (Rust, Python, Go, Node.js)
- Dotfiles restoration and configuration
- Terminal configuration with Catppuccin Mocha theme
- Responsive UI that adapts to terminal size
- Keyboard layout support (QWERTY, Colemak-DH)
- Safe command validation to prevent destructive operations

## Installation

1. Clone or download this repository
2. Make sure you have Go installed (1.19+)
3. Build the application:
   ```bash
   go build -o MacDevTUI
   ```
4. Run the installer:
   ```bash
   ./MacDevTUI
   ```

## Configuration

The application looks for configuration files in the following locations:

- `./install-config.json` (current directory)
- `./config/install-config.json` (config directory)
- `~/.config/install-config.json` (user config)

See `/config/install-config.json` for the configuration format and available options.

## Usage

### Navigation

- **↑/↓ or j/k**: Navigate between installation steps
- **Space/Enter**: Toggle step enabled/disabled
- **Tab**: Start installation of enabled steps
- **?**: Show help screen
- **q/Esc**: Quit application

### Keyboard Layouts

The application supports both QWERTY and Colemak-DH keyboard layouts with appropriate key bindings.

## Safety Features

- Configuration validation prevents dangerous commands
- Bounds checking prevents runtime crashes
- Graceful shutdown handling
- Input sanitization for security

## Architecture

- `main.go`: TUI interface and application logic
- `config.go`: Configuration loading and validation
- `installer.go`: Installation step implementations
- `models.go`: Data structures and setup steps
- `theme.go`: UI styling and themes

## License

This project is for personal use in setting up development environments.
