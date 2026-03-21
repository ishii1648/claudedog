# Changelog

hitl-metrics の変更履歴。新しいものが上。

## 2026-03-21

- ADR-017: 設計/実装セッション分離の自動ディスパッチ
- Python バッチ2本（session-index-update.py, session-index-backfill-batch.py）を Go に移植
- dashboard/server.py を削除し、SQLite + Grafana に可視化を移行（ADR-015）
- `hitl-metrics sync-db` サブコマンドを新規追加（JSONL/log → SQLite 変換）
- `hitl-metrics update` / `hitl-metrics backfill` サブコマンドで Python 版を置換
- Grafana ダッシュボード定義を追加（grafana/）
- bash ラッパー `hitl-metrics` を削除（Go バイナリが代替）
- Python 依存をゼロに

## 2026-03-06

- `configs/claude/scripts/` からトップレベル `hitl-metrics/` に移動し、ディレクトリ隔離を実施（ADR-052）
- 開発プロセスを ADR 駆動から TODO.md + CHANGELOG.md ベースに移行
