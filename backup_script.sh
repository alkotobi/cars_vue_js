#!/bin/bash

# Database backup script for Merhab Cars
# This script creates automatic backups of the database

# Configuration
DB_HOST="localhost"
DB_USER="root"
DB_PASS=""
DB_NAME="merhab_cars"
BACKUP_DIR="./backups"
DATE_FORMAT="%Y-%m-%d_%H-%M-%S"
MAX_BACKUPS=10  # Keep only last 10 backups

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

# Generate backup filename with timestamp
BACKUP_FILE="merhab_cars_backup_$(date +$DATE_FORMAT).sql"
BACKUP_PATH="$BACKUP_DIR/$BACKUP_FILE"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting database backup...${NC}"

# Create backup using mysqldump
if mysqldump --host="$DB_HOST" --user="$DB_USER" --password="$DB_PASS" "$DB_NAME" > "$BACKUP_PATH"; then
    echo -e "${GREEN}✓ Backup created successfully: $BACKUP_FILE${NC}"
    
    # Get file size
    FILE_SIZE=$(du -h "$BACKUP_PATH" | cut -f1)
    echo -e "${GREEN}✓ Backup size: $FILE_SIZE${NC}"
    
    # Clean up old backups (keep only MAX_BACKUPS)
    echo -e "${YELLOW}Cleaning up old backups...${NC}"
    cd "$BACKUP_DIR"
    ls -t *.sql | tail -n +$((MAX_BACKUPS + 1)) | xargs -r rm
    echo -e "${GREEN}✓ Old backups cleaned up${NC}"
    
    echo -e "${GREEN}✓ Backup completed successfully!${NC}"
    echo -e "${GREEN}Backup location: $BACKUP_PATH${NC}"
else
    echo -e "${RED}✗ Backup failed!${NC}"
    exit 1
fi 