#!/bin/sh

# FreeBSD Build Script for Uptime Monitor
# This script builds both backend and frontend on FreeBSD

set -e

echo "Building Uptime Monitor for FreeBSD..."

# Check if Go is installed
if ! command -v go >/dev/null 2>&1; then
    echo "Go is not installed. Install with: pkg install go"
    exit 1
fi

# Check if Node.js is installed
if ! command -v node >/dev/null 2>&1; then
    echo "Node.js is not installed. Install with: pkg install node"
    exit 1
fi

# Check if npm is available
if ! command -v npm >/dev/null 2>&1; then
    echo "npm is not installed. Install with: pkg install npm"
    exit 1
fi

# Check if npm is available
if ! command -v npm >/dev/null 2>&1; then
    echo "npm is not installed. Install with: pkg install npm"
    exit 1
fi

# Build backend
echo "Building Go backend..."
cd backend
go mod tidy
mkdir -p build

# Build for FreeBSD
CGO_ENABLED=1 GOOS=freebsd go build -o build/uptime-monitor ./cmd/main.go

echo "Backend built successfully: backend/build/uptime-monitor"
cd ..

# Build frontend
echo "Building SvelteKit frontend..."
cd frontend

# Install dependencies
npm install

# Build for production
npm run build

echo "Frontend built successfully: frontend/dist/"
cd ..

echo ""
echo "Build completed!"
echo "Backend binary: backend/build/uptime-monitor"
echo "Frontend files: frontend/dist/"
echo ""
echo "Run ./deploy/freebsd/install.sh as root to install the application."