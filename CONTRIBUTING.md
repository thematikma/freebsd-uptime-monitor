# Contributing to FreeBSD Uptime Monitor

Thank you for your interest in contributing to the FreeBSD Uptime Monitor project! This document provides guidelines for contributing.

## Development Setup

### Prerequisites
- FreeBSD system (recommended) or Linux/macOS for development
- Go 1.21 or later
- Node.js 18 or later
- Git

### Local Development

1. **Clone the repository**:
   ```bash
   git clone https://github.com/[username]/uptime-monitor.git
   cd uptime-monitor
   ```

2. **Backend development**:
   ```bash
   cd backend
   go mod tidy
   go run cmd/main.go
   ```

3. **Frontend development** (in a separate terminal):
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

## Contributing Guidelines

### Code Style
- **Go**: Follow standard Go conventions and use `gofmt`
- **TypeScript/Svelte**: Use the included ESLint and Prettier configurations
- **Comments**: Write clear, concise comments for complex logic

### Commit Messages
Use conventional commit format:
- `feat: add new monitoring type`
- `fix: resolve websocket connection issue`
- `docs: update installation instructions`
- `refactor: improve database connection handling`

### Pull Request Process

1. **Fork the repository** and create your feature branch from `main`
2. **Test your changes** thoroughly on FreeBSD if possible
3. **Update documentation** as needed
4. **Ensure the build passes**:
   ```bash
   ./deploy/freebsd/build.sh
   ```
5. **Submit a pull request** with a clear description of changes

### Testing

- Test on FreeBSD when possible (VM or physical hardware)
- Ensure all monitor types work correctly
- Test both SQLite and PostgreSQL database options
- Verify the web interface functions properly

### FreeBSD Compatibility

This project prioritizes FreeBSD compatibility. When contributing:

- Use cross-platform Go libraries
- Avoid Linux-specific system calls
- Test networking features work with FreeBSD's network stack
- Ensure the rc.d service script follows FreeBSD conventions

### Areas for Contribution

- **Monitor Types**: Additional monitoring capabilities
- **Alerting**: Email, Slack, webhook notifications
- **UI/UX**: Frontend improvements and new features
- **Documentation**: Installation guides, configuration examples
- **Testing**: Unit tests, integration tests
- **Performance**: Optimization and resource usage improvements

### Reporting Issues

When reporting bugs:

1. **Check existing issues** first
2. **Provide system information** (FreeBSD version, architecture)
3. **Include logs** when relevant
4. **Provide steps to reproduce** the issue
5. **Include configuration** (remove sensitive data)

### Security Issues

For security vulnerabilities, please email directly instead of opening a public issue.

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow
- Keep discussions relevant to the project

## Questions?

Feel free to open an issue for questions about contributing or FreeBSD-specific implementation details.

Thank you for contributing to making uptime monitoring better on FreeBSD!