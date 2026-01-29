#!/bin/sh

# FreeBSD Uptime Monitor Installation Script
# Usage: ./install.sh

set -e

INSTALL_DIR="/usr/local/uptime-monitor"
USER="uptime"
GROUP="uptime"
SERVICE_NAME="uptime_monitor"

echo "Installing Uptime Monitor for FreeBSD..."

# Check if running as root
if [ "$(id -u)" != "0" ]; then
   echo "This script must be run as root" 1>&2
   exit 1
fi

# Create user and group
if ! pw group show "$GROUP" >/dev/null 2>&1; then
    echo "Creating group: $GROUP"
    pw groupadd "$GROUP"
fi

if ! pw user show "$USER" >/dev/null 2>&1; then
    echo "Creating user: $USER"
    pw useradd "$USER" -g "$GROUP" -d "$INSTALL_DIR" -s /usr/sbin/nologin -c "Uptime Monitor User"
fi

# Create directories
echo "Creating directories..."
mkdir -p "$INSTALL_DIR/bin"
mkdir -p "$INSTALL_DIR/frontend/dist"
mkdir -p "$INSTALL_DIR/data"
mkdir -p "/var/log/uptime-monitor"

# Copy files
echo "Copying application files..."
cp backend/build/uptime-monitor "$INSTALL_DIR/bin/"
cp -r frontend/dist/* "$INSTALL_DIR/frontend/dist/"

# Set permissions
echo "Setting permissions..."
chown -R "$USER:$GROUP" "$INSTALL_DIR"
chown -R "$USER:$GROUP" "/var/log/uptime-monitor"
chmod +x "$INSTALL_DIR/bin/uptime-monitor"

# Create configuration file
echo "Creating configuration..."
cat > "$INSTALL_DIR/uptime-monitor.conf" << 'EOF'
# Uptime Monitor Configuration
PORT=8080
HOST=0.0.0.0
DB_TYPE=sqlite
DB_NAME=/usr/local/uptime-monitor/data/uptime.db
JWT_SECRET=change-this-secret-key
MONITOR_INTERVAL=60
MONITOR_TIMEOUT=30
MONITOR_RETRIES=3
EOF

chown "$USER:$GROUP" "$INSTALL_DIR/uptime-monitor.conf"

# Create rc.d service script
echo "Creating service script..."
cat > "/usr/local/etc/rc.d/$SERVICE_NAME" << 'EOF'
#!/bin/sh
# PROVIDE: uptime_monitor
# REQUIRE: LOGIN
# KEYWORD: shutdown

. /etc/rc.subr

name="uptime_monitor"
rcvar="uptime_monitor_enable"

command="/usr/local/uptime-monitor/bin/uptime-monitor"
pidfile="/var/run/uptime-monitor.pid"
uptime_monitor_user="uptime"
uptime_monitor_group="uptime"
uptime_monitor_chdir="/usr/local/uptime-monitor"

# Set environment variables from config file
if [ -f "/usr/local/uptime-monitor/uptime-monitor.conf" ]; then
    . "/usr/local/uptime-monitor/uptime-monitor.conf"
fi

command_args="&"

load_rc_config $name
run_rc_command "$1"
EOF

chmod +x "/usr/local/etc/rc.d/$SERVICE_NAME"

# Create newsyslog configuration for log rotation
echo "Configuring log rotation..."
cat >> "/etc/newsyslog.conf" << EOF
/var/log/uptime-monitor/uptime-monitor.log    $USER:$GROUP   644  7     *    @T00  JC
EOF

echo "Installation completed!"
echo ""
echo "To start the service:"
echo "  sysrc ${SERVICE_NAME}_enable=YES"
echo "  service $SERVICE_NAME start"
echo ""
echo "Configuration file: $INSTALL_DIR/uptime-monitor.conf"
echo "Log files: /var/log/uptime-monitor/"
echo "Web interface will be available at: http://localhost:8080"