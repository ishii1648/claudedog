# TODO

hitl-metrics の開発タスクを管理する。完了したタスクは削除する。変更履歴は git log と GitHub Release を参照し、設計の経緯は `docs/history.md` に集約する。

## 実装タスク

- Codex CLI 対応 — Grafana ダッシュボード
  - [ ] Agent 別比較 stat パネルを追加（avg tokens / PR と PR / 1M tokens）
  - [ ] `make grafana-screenshot` を実行して `docs/images/dashboard-*.png` を更新する
- sync-db を incremental UPSERT に切り替えて Grafana との race condition を解消する
  - [ ] `internal/syncdb/schema.sql` を新規作成し、現行 `schema.go` の DDL（テーブル/INDEX/VIEW、`permission_events` の legacy DROP も含む）を全て移管する
  - [ ] `go:embed` で `schema.sql` を読み込み、`go:generate` で SHA256 を生成して埋め込む
  - [ ] `schema_meta` テーブル（最後に適用したスキーマハッシュを保存）を `schema.sql` に追加する
  - [ ] sync-db 起動時に埋め込みハッシュと `schema_meta` のハッシュを比較する（`schema_meta` 不在も不一致扱い）
  - [ ] ハッシュ一致時は DDL を実行せず `sessions` / `transcript_stats` への INSERT OR REPLACE のみ実行する
  - [ ] ハッシュ不一致時は `schema.sql` を実行（既存テーブル/VIEW を DROP & CREATE）し、INSERT OR REPLACE 後に新ハッシュを `schema_meta` に書き込む
  - [ ] GitHub Actions ワークフローを新規追加し、`go generate ./... && git diff --exit-code` でハッシュ生成漏れを検出する
  - [ ] 既存の `internal/syncdb/syncdb_test.go` が通る
  - [ ] 同一 session-index で 2 回 Run しても行数が増えず最新値で上書きされるテストを追加する
  - [ ] スキーマハッシュが一致しない既存 DB に対して Run しても自動 fallback で成功するテストを追加する
  - [ ] `docs/design.md`「既知の制約」から DROP & CREATE による Grafana race condition の記述を削除する

## 検討中

- Stop hook の `hitl-metrics` PATH 依存をなくす — 解決方針を決める
  - 候補 A: `backfill` / `sync-db` を `internal/` 関数として直接呼ぶ（同一プロセス、PATH 非依存）
  - 候補 B: `setup` 時に hook コマンドの絶対パスを案内する（`settings.json` / `config.toml` 側で絶対パスを書く）
  - 候補 C: hook 内で binary を `os.Executable()` で解決し PATH にフォールバックしない
  - 失敗時ログの設計（PATH 不在 / 内部エラーの切り分け）も方針に含める
  - 方針確定後に受け入れ条件を整えて実装タスクへ昇格させる

- ローカル検証環境と CI の再現性 — 完了条件を具体化する
  - SQLite テストの不安定要因（macOS arm64 で `modernc.org/sqlite` 使用時に発生する事象）を特定し、安定化の具体条件を決める
  - `go test -race` がローカルで実行不能な事例を整理し、代替手順（CI に委ねる / Docker で回す等）の方針を決める
  - 制約整理の記録先（`docs/setup.md` か別 docs か）を決める

- Bash コマンドのコンテキスト消費監視
  - `PostToolUse` hook で Bash コマンドの stdout サイズを記録する想定
  - redirect-to-tools をすり抜けた正当な Bash コマンドのうち、出力が大きいものを特定する
  - 定期集計で「常連犯」コマンドを可視化し、対策要否を判断する
  - 受け入れ条件（記録先・閾値・集計方法）が未確定

- retro-pr との連携
  - PR の下位・上位 10% ずつは自動で retro-pr 実行する想定
  - 結果を PR と関連付けて表示する想定
  - 受け入れ条件（連携方式・表示先・自動化対象）が未確定
