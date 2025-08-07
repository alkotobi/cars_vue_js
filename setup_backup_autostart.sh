#!/bin/bash

# Setup script for Merhab Cars Backup Auto-Start
# This script helps set up the AppleScript to run on Mac startup

echo "=== Merhab Cars Backup Auto-Start Setup ==="
echo ""

# Check if we're on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo "❌ This script is for macOS only"
    exit 1
fi

# Create the AppleScript if it doesn't exist
if [ ! -f "open_backup_website.scpt" ]; then
    echo "❌ AppleScript file not found. Please make sure 'open_backup_website.scpt' is in the current directory."
    exit 1
fi

echo "📋 Setup Steps:"
echo ""
echo "1. Open Script Editor:"
echo "   - Go to Applications → Utilities → Script Editor"
echo "   - Or use Spotlight (Cmd+Space) and search 'Script Editor'"
echo ""
echo "2. Open the AppleScript:"
echo "   - In Script Editor, go to File → Open"
echo "   - Navigate to this folder and select 'open_backup_website.scpt'"
echo ""
echo "3. Test the script:"
echo "   - Click the 'Run' button (play icon) in Script Editor"
echo "   - Safari should open with the backup website"
echo ""
echo "4. Save as Application:"
echo "   - In Script Editor, go to File → Export"
echo "   - Choose 'Application' as the file format"
echo "   - Save as 'Backup Website.app' in your Applications folder"
echo ""
echo "5. Add to Login Items:"
echo "   - Go to System Preferences → Users & Groups"
echo "   - Select your user account → Login Items tab"
echo "   - Click '+' button and add 'Backup Website.app'"
echo ""
echo "✅ Done! The backup website will now open automatically when you start your Mac."
echo ""
echo "💡 Tips:"
echo "- The script will show a notification when it runs"
echo "- Safari will open with the backup interface ready"
echo "- Automatic backups will start based on your settings"
echo ""
echo "🔧 To remove auto-start later:"
echo "- Go to System Preferences → Users & Groups → Login Items"
echo "- Select 'Backup Website.app' and click '-' to remove it" 