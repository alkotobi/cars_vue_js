# Database Backup System

This system provides both manual and automatic database backup functionality for the Merhab Cars application.

## Features

- **Manual Backup**: Click the "Database Backup" button in the dashboard (admin only)
- **Automatic Backup**: Scheduled backups using cron jobs
- **Backup Management**: Automatic cleanup of old backups
- **Multiple Formats**: SQL dump with both mysqldump and PHP fallback

## Manual Backup

### Dashboard Button
1. Log in as an administrator
2. Go to the Dashboard
3. Click the "Database Backup" button
4. The backup will be downloaded automatically

### API Endpoint
- **URL**: `/api/backup.php`
- **Method**: GET
- **Authentication**: Required (admin only)
- **Parameters**: 
  - `filename` (optional): Custom filename for the backup

## Automatic Backup Setup

### 1. Configure the Backup Script

Edit `backup_script.sh` and update the database configuration:

```bash
DB_HOST="localhost"
DB_USER="your_username"
DB_PASS="your_password"
DB_NAME="merhab_cars"
BACKUP_DIR="./backups"
MAX_BACKUPS=10  # Number of backups to keep
```

### 2. Test the Script

```bash
./backup_script.sh
```

### 3. Set Up Cron Job

Add to your crontab for automatic daily backups:

```bash
# Open crontab editor
crontab -e

# Add one of these lines:
# Daily backup at 2 AM
0 2 * * * /path/to/your/project/backup_script.sh

# Weekly backup on Sunday at 2 AM
0 2 * * 0 /path/to/your/project/backup_script.sh

# Multiple times per day (every 6 hours)
0 */6 * * * /path/to/your/project/backup_script.sh
```

### 4. Verify Cron Job

```bash
# List current cron jobs
crontab -l

# Check cron logs
tail -f /var/log/cron
```

## Backup Locations

- **Manual Backups**: Downloaded to your browser's default download folder
- **Automatic Backups**: Stored in `./backups/` directory
- **Backup Format**: `merhab_cars_backup_YYYY-MM-DD_HH-MM-SS.sql`

## Backup Content

The backup includes:
- Complete database structure (CREATE TABLE statements)
- All data from all tables
- Stored procedures and functions
- Triggers and events
- Database configuration

## Security

- Only administrators can access the backup functionality
- Backups are created with proper file permissions
- Old backups are automatically cleaned up to save space
- Database credentials are protected

## Troubleshooting

### Common Issues

1. **Permission Denied**
   ```bash
   chmod +x backup_script.sh
   ```

2. **mysqldump not found**
   ```bash
   # Install MySQL client tools
   sudo apt-get install mysql-client  # Ubuntu/Debian
   brew install mysql-client          # macOS
   ```

3. **Database connection failed**
   - Check database credentials in `backup_script.sh`
   - Verify database server is running
   - Check firewall settings

4. **Cron job not running**
   ```bash
   # Check cron service
   sudo service cron status
   
   # Check cron logs
   tail -f /var/log/cron
   ```

### Manual Backup Recovery

To restore a backup:

```bash
# Method 1: Using mysql command
mysql -u username -p database_name < backup_file.sql

# Method 2: Using phpMyAdmin
# Import the .sql file through phpMyAdmin interface
```

## Monitoring

### Check Backup Status

```bash
# List recent backups
ls -la backups/

# Check backup sizes
du -h backups/*.sql

# Verify backup integrity
head -20 backups/latest_backup.sql
```

### Log Monitoring

The backup script provides colored output for easy monitoring:
- 🟢 Green: Success messages
- 🟡 Yellow: Warning messages  
- 🔴 Red: Error messages

## Configuration Options

### Backup Script Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | Database host | localhost |
| `DB_USER` | Database username | root |
| `DB_PASS` | Database password | (empty) |
| `DB_NAME` | Database name | merhab_cars |
| `BACKUP_DIR` | Backup storage directory | ./backups |
| `MAX_BACKUPS` | Number of backups to keep | 10 |

### API Configuration

The backup API (`api/backup.php`) includes:
- Authentication checks
- Admin role verification
- Automatic filename generation
- Fallback PHP backup method

## Best Practices

1. **Regular Testing**: Test backup restoration periodically
2. **Multiple Locations**: Store backups in different locations
3. **Monitoring**: Set up alerts for failed backups
4. **Documentation**: Keep backup procedures documented
5. **Security**: Protect backup files with proper permissions

## Support

For issues with the backup system:
1. Check the troubleshooting section above
2. Review the backup logs
3. Test the backup script manually
4. Verify database connectivity 