#!/bin/bash

# Advanced Setup script for Merhab Cars Backup Monitor
# This script offers different visibility options for the monitoring service

echo "=== Merhab Cars Backup Monitor - Advanced Setup ==="
echo ""

# Check if we're on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo "❌ This script is for macOS only"
    exit 1
fi

echo "🔍 Visibility Options for Backup Monitor:"
echo ""
echo "1. Dock Icon (Default) - Shows small icon in dock when running"
echo "2. Hidden App - Runs with minimal visibility, no dock icon"
echo "3. Launch Agent - Runs completely in background (requires sudo)"
echo "4. Manual Only - No automatic startup, run manually when needed"
echo ""

read -p "Choose option (1-4): " choice

case $choice in
    1)
        echo "📦 Setting up Dock Icon version..."
        osacompile -o "Backup Monitor.app" backup_monitor.scpt
        mv "Backup Monitor.app" /Applications/
        osascript -e 'tell application "System Events" to make login item at end with properties {path:"/Applications/Backup Monitor.app", hidden:false}'
        echo "✅ Dock Icon version installed - will show small icon in dock"
        ;;
    2)
        echo "📦 Setting up Hidden App version..."
        osacompile -o "Backup Monitor.app" backup_monitor.scpt
        mv "Backup Monitor.app" /Applications/
        osascript -e 'tell application "System Events" to make login item at end with properties {path:"/Applications/Backup Monitor.app", hidden:true}'
        echo "✅ Hidden App version installed - minimal visibility"
        ;;
    3)
        echo "📦 Setting up Launch Agent version (requires sudo)..."
        osacompile -o "Backup Monitor.app" backup_monitor.scpt
        mv "Backup Monitor.app" /Applications/
        
        # Install launch agent
        sudo cp com.merhab.backup.monitor.plist /Library/LaunchAgents/
        sudo launchctl load /Library/LaunchAgents/com.merhab.backup.monitor.plist
        
        echo "✅ Launch Agent version installed - completely invisible background service"
        echo "📝 Logs available at: /tmp/backup_monitor.log"
        ;;
    4)
        echo "📦 Setting up Manual Only version..."
        osacompile -o "Backup Monitor.app" backup_monitor.scpt
        mv "Backup Monitor.app" /Applications/
        echo "✅ Manual Only version installed - run manually when needed"
        echo "💡 To run: open '/Applications/Backup Monitor.app'"
        ;;
    *)
        echo "❌ Invalid choice. Exiting."
        exit 1
        ;;
esac

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
echo "• Ensures backup system stays active"
echo ""
echo "🔧 To remove:"
if [[ $choice == 3 ]]; then
    echo "• Run: sudo launchctl unload /Library/LaunchAgents/com.merhab.backup.monitor.plist"
    echo "• Delete: sudo rm /Library/LaunchAgents/com.merhab.backup.monitor.plist"
fi
echo "• System Preferences → Users & Groups → Login Items"
echo "• Remove 'Backup Launcher' and 'Backup Monitor'"
echo ""
echo "🧪 Test the monitor:"
echo "• Open Safari and close the backup website tab"
echo "• Wait up to 5 minutes for it to reopen automatically"
echo "• Or run: open '/Applications/Backup Monitor.app' to test immediately" 