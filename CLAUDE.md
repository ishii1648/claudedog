# hitl-metrics

Claude Code の人の介入率を追跡・可視化する計測ツール（hook・CLI・ダッシュボード）。

## アーキテクチャ

3層構成:

1. **データ収集層** (`hooks/`) — Claude Code hook で session/permission/tool-use イベントを JSONL/log に記録
2. **データ変換層** (`cmd/hitl-metrics/`, `internal/`) — Go CLI で JSONL/log → SQLite 変換・PR URL 補完
3. **可視化層** (`grafana/`) — Grafana ダッシュボードで介入率・ツール分布を表示

データフロー: `hooks → ~/.claude/*.jsonl|log → hitl-metrics sync-db → SQLite → Grafana`

## データモデル（SQLite）

- **sessions** — セッション単位の基本情報（session_id, repo, branch, pr_url, is_subagent 等）
- **permission_events** — permission UI 発生イベント（timestamp, session_id, tool）
- **transcript_stats** — トランスクリプトから抽出した統計（tool_use_total, mid_session_msgs, ask_user_question, is_ghost）
- **pr_metrics**（VIEW） — PR 単位で session_count, perm_count, perm_rate を集約

## CLI コマンド

```
hitl-metrics update <session_id> <url>...          # PR URL を追加
hitl-metrics update --mark-checked <session_id>... # backfill_checked をセット
hitl-metrics update --by-branch <repo> <branch> <url>  # ブランチ全セッションに URL 追加
hitl-metrics backfill [--recheck]                  # PR URL の一括補完
hitl-metrics sync-db                               # JSONL/log → SQLite 変換
```

## 開発規約

### 意思決定の記録方針

- 複数コミットにまたがる設計判断 → `docs/adr/` に ADR を作成
- 1コミット内で完結する判断 → Contextual Commits のアクション行で記録
- chore/リファクタなど意思決定を伴わない変更 → アクション行不要

### コミット

Contextual Commits を使用。Conventional Commits プレフィックス + 構造化されたアクション行でコミットの意図を記録する。

### ブランチ命名

`feat/`, `fix/`, `docs/`, `chore/` + kebab-case（例: `feat/add-sync-db`）

### テスト

```fish
go test ./...                          # 全テスト
make grafana-screenshot                # E2E: Grafana スクリーンショット検証
```
