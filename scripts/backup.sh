#!/bin/bash
# Backup PostgreSQL Tuléh Server — pg_dump ter-gzip, ber-timestamp, dgn rotasi.
# Dipanggil systemd timer harian (deploy/tuleh-backup.timer) atau manual.
#
# Data merchant sungguhan hidup di sini — backup adalah jaring pengaman utama.
set -euo pipefail

DIR="${TULEH_BACKUP_DIR:-/home/dede/tuleh-server/backups}"
SIMPAN_HARI="${TULEH_BACKUP_KEEP:-14}"
CONTAINER="tuleh-postgres"
DB="tuleh_pos"
USER="tuleh"

mkdir -p "$DIR"
STAMP=$(date +%Y%m%d-%H%M%S)
TARGET="$DIR/tuleh_pos-$STAMP.sql.gz"

# --clean --if-exists → dump bisa di-restore ke DB berisi tanpa error duplikat.
docker exec "$CONTAINER" pg_dump -U "$USER" --clean --if-exists "$DB" | gzip -9 > "$TARGET"

UKURAN=$(du -h "$TARGET" | cut -f1)
echo "[$(date '+%F %T')] backup OK: $TARGET ($UKURAN)"

# Rotasi: hapus backup lebih tua dari SIMPAN_HARI (default 14).
find "$DIR" -name 'tuleh_pos-*.sql.gz' -type f -mtime "+$SIMPAN_HARI" -delete
echo "[$(date '+%F %T')] rotasi selesai (simpan $SIMPAN_HARI hari) — tersisa $(find "$DIR" -name 'tuleh_pos-*.sql.gz' | wc -l) berkas"
