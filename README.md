# This is AI Slop. I wanted to see what a Coding Agent can do. If you found this, do not use in any environments outside your home :)
# Uptime Monitor - FreeBSD Edition

A comprehensive uptime monitoring solution similar to Uptime Kuma, designed with full FreeBSD support.

**Primary Target**: FreeBSD 15+ with native Node.js 24 support

## Features

- **Multiple Monitor Types**: HTTP/HTTPS, TCP, and Ping monitoring
- **Real-time Dashboard**: Live status updates with WebSocket connections
- **FreeBSD Native**: Built specifically for FreeBSD with native service integration
- **Lightweight**: Go backend and SvelteKit frontend for minimal resource usage
- **Database Support**: SQLite (default) and PostgreSQL
- **Alert System**: Configurable notifications (planned)
- **Authentication**: Secure user management with JWT tokens

## Technology Stack

- **Backend**: Go 1.21+ with Gin web framework
- **Frontend**: SvelteKit with TypeScript
- **Database**: SQLite (default) or PostgreSQL
- **Monitoring**: Custom Go monitoring engine
- **WebSockets**: Real-time status updates
- **Deployment**: FreeBSD rc.d service integration

## Quick Start

### Prerequisites

FreeBSD 15+ system with:
- Go 1.21 or later: `pkg install go`
- Node.js 18+ (FreeBSD 15 includes Node.js 24): `pkg install node npm`
- Git: `pkg install git`

**Note**: This application is designed for FreeBSD 15+ and requires Node.js 18+. For development on other systems, ensure Node.js 18+ is installed.

### Installation

1. **Clone the repository**:
   ```bash
   git clone https://github.com/thematikma/freebsd-uptime-monitor.git
   cd freebsd-uptime-monitor/
   ```

2. **Build the application**:
   ```bash
   chmod +x deploy/freebsd/build.sh
   ./deploy/freebsd/build.sh
   ```

3. **Install system-wide** (as root):
   ```bash
   chmod +x deploy/freebsd/install.sh
   sudo ./deploy/freebsd/install.sh
   ```

4. **Start the service**:
   ```bash
   sudo sysrc uptime_monitor_enable=YES
   sudo service uptime_monitor start
   ```

5. **Access the web interface**:
   Open http://localhost:8080 in your browser

## Configuration

Edit `/usr/local/uptime-monitor/uptime-monitor.conf`:

```bash
PORT=8080
HOST=0.0.0.0
DB_TYPE=sqlite
DB_NAME=/usr/local/uptime-monitor/data/uptime.db
JWT_SECRET=your-secure-secret-key
MONITOR_INTERVAL=60
MONITOR_TIMEOUT=30
MONITOR_RETRIES=3
```

## Development

**Important**: Development requires Node.js 18+. On FreeBSD 15+, this is automatically satisfied. On other systems, install Node.js 18+ before proceeding.

### Backend Development

```bash
cd backend
go mod tidy
go run cmd/main.go
```

### Frontend Development

```bash
cd frontend
npm install
npm run dev
```

The development server will proxy API requests to the Go backend running on port 8080.

### Development on Non-FreeBSD Systems

If developing on systems with older Node.js versions (like Node.js 16), you'll need to upgrade:
- **Ubuntu/Debian**: Use NodeSource repository for Node.js 18+
- **Fedora**: `dnf install nodejs npm` (usually provides 18+)
- **macOS**: Use Homebrew: `brew install node`

## Project Structure

```
├── backend/                 # Go backend application
│   ├── cmd/                # Application entry point
│   ├── internal/           # Internal packages
│   │   ├── auth/          # Authentication system
│   │   ├── config/        # Configuration management
│   │   ├── database/      # Database layer
│   │   ├── handlers/      # HTTP handlers
│   │   ├── models/        # Data models
│   │   ├── monitoring/    # Monitoring engine
│   │   └── websocket/     # WebSocket hub
│   └── go.mod             # Go dependencies
├── frontend/               # SvelteKit frontend
│   ├── src/               # Source code
│   │   ├── lib/          # Shared components and stores
│   │   └── routes/       # Page routes
│   ├── package.json      # Node.js dependencies
│   └── svelte.config.js  # Svelte configuration
├── deploy/                # Deployment scripts
│   └── freebsd/          # FreeBSD-specific deployment
└── README.md             # This file
```

## Monitor Types

### HTTP/HTTPS Monitoring
- Monitors web endpoints
- Checks response codes and response times
- Configurable timeout and retry settings

### TCP Monitoring
- Tests TCP port connectivity
- Useful for database servers, mail servers, etc.
- Format: `tcp://hostname:port`

### Ping Monitoring
- ICMP ping tests
- Measures packet loss and response times
- Format: `ping://hostname`

## Service Management

The application installs as a FreeBSD service:

```bash
# Start/stop/restart
sudo service uptime_monitor start
sudo service uptime_monitor stop
sudo service uptime_monitor restart

# Check status
sudo service uptime_monitor status

# View logs
tail -f /var/log/uptime-monitor/uptime-monitor.log
```

## Database

### SQLite (Default)
- No additional setup required
- Database file: `/usr/local/uptime-monitor/data/uptime.db`
- Suitable for most installations

### PostgreSQL
Update configuration:
```bash
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=uptime_monitor
DB_USER=uptime
DB_PASS=password
DB_SSLMODE=disable
```

## Security Considerations

1. **Change JWT Secret**: Update `JWT_SECRET` in the configuration file
2. **Firewall**: Restrict access to port 8080 as needed
3. **Database**: Use PostgreSQL with proper authentication for production
4. **HTTPS**: Use a reverse proxy (nginx) for HTTPS termination

## Contributing

1. Fork the repository
2. Create a feature branch
3. Test on FreeBSD
4. Submit a pull request

## License

MIT License

## Support

For FreeBSD-specific issues or general support, please open an issue in the repository.