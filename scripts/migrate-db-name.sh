#!/usr/bin/env bash
# 使い捨てスクリプト: ~/.claude/claudedog.db → ~/.claude/hitl-metrics.db にリネーム
# 実行後に削除してOK
set -euo pipefail

SRC="$HOME/.claude/claudedog.db"
DST="$HOME/.claude/hitl-metrics.db"

if [ ! -f "$SRC" ]; then
  echo "Source not found: $SRC (nothing to migrate)"
  exit 0
fi

if [ -f "$DST" ]; then
  echo "Destination already exists: $DST (aborting to avoid overwrite)"
  exit 1
fi

mv "$SRC" "$DST"
[ -f "${SRC}-shm" ] && mv "${SRC}-shm" "${DST}-shm"
[ -f "${SRC}-wal" ] && mv "${SRC}-wal" "${DST}-wal"

echo "Migrated: $SRC → $DST"
