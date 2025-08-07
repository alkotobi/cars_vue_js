-- AppleScript to open Safari with Merhab Cars Backup Website
-- This script opens Safari and navigates to the backup page
-- It also monitors and ensures the backup website stays open

-- Function to check if backup website is open
on checkBackupWebsite()
	tell application "Safari"
		set backupUrl to "https://www.merhab.com/api/backup_simple_web.php"
		set backupWebsiteOpen to false
		
		-- Check all Safari windows
		repeat with currentWindow in windows
			repeat with currentTab in tabs of currentWindow
				if URL of currentTab contains backupUrl then
					set backupWebsiteOpen to true
					exit repeat
				end if
			end repeat
		end repeat
		
		-- If backup website is not open, open it
		if not backupWebsiteOpen then
			activate
			open location backupUrl
			delay 2
			display notification "Backup website reopened" with title "Merhab Cars Backup" subtitle "Ensuring backup system is active"
		end if
	end tell
end checkBackupWebsite

-- Initial setup
tell application "Safari"
	activate
	-- Open the backup website in a new window
	open location "https://www.merhab.com/api/backup_simple_web.php"
	
	-- Wait a moment for the page to load
	delay 2
end tell

-- Display initial notification
display notification "Backup system is ready" with title "Merhab Cars Backup" subtitle "Safari opened with backup interface"

-- Start monitoring loop (check every 5 minutes)
repeat
	checkBackupWebsite()
	delay 300 -- Wait 5 minutes before next check
end repeat 