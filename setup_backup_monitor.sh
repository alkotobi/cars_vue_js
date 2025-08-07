#!/bin/bash

# Setup script for Merhab Cars Backup Monitor
# This script sets up both the initial launcher and the monitoring service

echo "=== Merhab Cars Backup Monitor Setup ==="
echo ""

# Check if we're on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo "❌ This script is for macOS only"
    exit 1
fi

# Create the monitoring application
echo "📦 Creating backup monitor application..."
osacompile -o "Backup Monitor.app" backup_monitor.scpt

# Move to Applications folder
echo "📁 Moving to Applications folder..."
mv "Backup Monitor.app" /Applications/

# Create the initial launcher application
echo "📦 Creating backup launcher application..."
osacompile -o "Backup Launcher.app" open_backup_website.scpt

# Move to Applications folder
echo "📁 Moving to Applications folder..."
mv "Backup Launcher.app" /Applications/

# Add both to login items
echo "🔗 Adding to login items..."
osascript -e 'tell application "System Events" to make login item at end with properties {path:"/Applications/Backup Launcher.app", hidden:false}'
osascript -e 'tell application "System Events" to make login item at end with properties {path:"/Applications/Backup Monitor.app", hidden:true}'

echo ""
echo "✅ Setup Complete!"
echo ""
echo "📋 What's been installed:"
echo "• Backup Launcher.app - Opens Safari with backup website on startup"
echo "• Backup Monitor.app - Monitors Safari and reopens backup website if closed"
echo ""
echo "🔄 How it works:"
echo "1. On startup: Backup Launcher opens Safari with backup website"
echo "2. Every 5 minutes: Backup Monitor checks if backup website is still open"
echo "3. If closed: Backup Monitor reopens the backup website automatically"
echo ""
echo "💡 Features:"
echo "• Automatic monitoring every 5 minutes"
echo "• Notifications when website is reopened"
echo "• Runs in background (hidden from dock)"
echo "• Ensures backup system stays active"
echo ""
echo "🔧 To remove:"
echo "• System Preferences → Users & Groups → Login Items"
echo "• Remove both 'Backup Launcher' and 'Backup Monitor'"
echo ""
echo "🧪 Test the monitor:"
echo "• Open Safari and close the backup website tab"
echo "• Wait up to 5 minutes for it to reopen automatically"
echo "• Or run: open '/Applications/Backup Monitor.app' to test immediately" 