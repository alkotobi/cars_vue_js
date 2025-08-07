-- Backup Website Monitor
-- This script monitors Safari and ensures the backup website stays open
-- Run this as a background service

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

-- Start monitoring loop (check every 5 minutes)
repeat
	checkBackupWebsite()
	delay 300 -- Wait 5 minutes before next check
end repeat 