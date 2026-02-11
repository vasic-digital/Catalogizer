#\!/bin/bash
# Install Tauri v2 build dependencies for ALT Linux 11
# Run with: sudo bash scripts/install-tauri-deps.sh

set -e

echo "Installing Tauri v2 build dependencies for ALT Linux..."

apt-get install -y     libgtk+3-devel     libatk-devel     libcairo-devel     libpango-devel     libwebkit2gtk4.1-devel     libsoup3.0-devel     librsvg-devel     glib2-devel     libgdk-pixbuf-devel     at-spi2-atk-devel

echo ""
echo "Verifying pkg-config files..."
pkg-config --libs webkit2gtk-4.1 && echo "webkit2gtk-4.1: OK" || echo "webkit2gtk-4.1: FAILED"
pkg-config --libs gtk+-3.0 && echo "gtk+-3.0: OK" || echo "gtk+-3.0: FAILED"
pkg-config --libs gdk-3.0 && echo "gdk-3.0: OK" || echo "gdk-3.0: FAILED"

echo ""
echo "Tauri v2 build dependencies installed successfully\!"
echo "You can now build with: npm run tauri:build"
