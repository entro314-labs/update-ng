#!/bin/bash
# Build script for update-ng
set -e

echo "🚀 Building Update Command NG..."

# Build for current platform
go build -o update-ng main.go
echo "✅ Built update-ng for $(go env GOOS)/$(go env GOARCH)"

# Optional: Build for multiple platforms
if [ "$1" = "all" ]; then
    echo "📦 Building for all platforms..."

    # macOS
    GOOS=darwin GOARCH=amd64 go build -o dist/update-ng-darwin-amd64 main.go
    GOOS=darwin GOARCH=arm64 go build -o dist/update-ng-darwin-arm64 main.go

    # Linux
    GOOS=linux GOARCH=amd64 go build -o dist/update-ng-linux-amd64 main.go
    GOOS=linux GOARCH=arm64 go build -o dist/update-ng-linux-arm64 main.go

    # Windows
    GOOS=windows GOARCH=amd64 go build -o dist/update-ng-windows-amd64.exe main.go

    echo "✅ Built for all platforms in dist/"
fi

echo "🎉 Build complete! Run with: ./update-ng"