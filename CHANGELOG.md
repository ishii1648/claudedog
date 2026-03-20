# Changelog

claudedog の変更履歴。新しいものが上。

## 2026-03-21

- Python バッチ2本（session-index-update.py, session-index-backfill-batch.py）を Go に移植
- dashboard/server.py を削除し、SQLite + Grafana に可視化を移行（ADR-015）
- `claudedog sync-db` サブコマンドを新規追加（JSONL/log → SQLite 変換）
- `claudedog update` / `claudedog backfill` サブコマンドで Python 版を置換
- Grafana ダッシュボード定義を追加（grafana/）
- bash ラッパー `claudedog` を削除（Go バイナリが代替）
- Python 依存をゼロに

## 2026-03-06

- `configs/claude/scripts/` からトップレベル `claudedog/` に移動し、ディレクトリ隔離を実施（ADR-052）
- 開発プロセスを ADR 駆動から TODO.md + CHANGELOG.md ベースに移行
