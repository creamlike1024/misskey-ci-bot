#!/bin/bash
# $1: backup name
# $2: Is a upgrade backup? (0 or 1)
if [ "$2" = "1" ]; then
  DBFILE=/tmp/misskey_db-$(date +%Y-%m-%d)-before-$1.sql
else
  DBFILE=/tmp/misskey_db-$(date +%Y-%m-%d)-$1.sql
fi
cd /opt/misskey && docker compose exec db pg_dump -U misskey misskey>"$DBFILE" && rsync -e ssh "$DBFILE" claire@205.185.117.85:/mnt/sda1/misskey-backup/db/ && rm "$DBFILE"