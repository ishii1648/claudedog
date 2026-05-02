# hitl-metrics

Claude Code の PR 単位のトークン消費効率を追跡・可視化する計測ツール（hook・CLI・ダッシュボード）。

## アーキテクチャ

3層構成:

1. **データ収集層** (`hooks/`) — Claude Code hook で session イベントを JSONL に記録
2. **データ変換層** (`cmd/hitl-metrics/`, `internal/`) — Go CLI で JSONL/transcript → SQLite 変換・PR URL 補完
3. **可視化層** (`grafana/`) — Grafana ダッシュボードで PR 単位の token 効率を表示

データフロー: `hooks → ~/.claude/session-index.jsonl + transcript JSONL → hitl-metrics sync-db → SQLite → Grafana`

## データモデル（SQLite）

- **sessions** — セッション単位の基本情報（session_id, repo, branch, pr_url, is_subagent 等）
- **transcript_stats** — トランスクリプトから抽出した統計（tool_use_total, mid_session_msgs, ask_user_question, token usage, is_ghost）
- **pr_metrics**（VIEW） — PR 単位で total_tokens, tokens_per_tool_use, pr_per_million_tokens を集約

## CLI コマンド

```
hitl-metrics update <session_id> <url>...          # PR URL を追加
hitl-metrics update --mark-checked <session_id>... # backfill_checked をセット
hitl-metrics update --by-branch <repo> <branch> <url>  # ブランチ全セッションに URL 追加
hitl-metrics backfill [--recheck]                  # PR URL の一括補完
hitl-metrics sync-db                               # JSONL/transcript → SQLite 変換
hitl-metrics install                               # セットアップ案内（hook 登録は dotfiles または手動）
hitl-metrics install --uninstall-hooks             # 旧 install で書き込んだ hook を settings.json から削除
hitl-metrics doctor                                # binary / data dir / hook 登録の検証
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

### ダッシュボード変更時の必須作業

- `grafana/dashboards/hitl-metrics.json` の表示を変更した場合は、必ず `make grafana-screenshot` を実行して README 用スクリーンショット（`docs/images/dashboard-*.png`）も同じ変更に合わせて更新する。
- スクリーンショット生成でポート競合が起きる場合は `GRAFANA_PORT=<unused-port> make grafana-screenshot` を使う。
