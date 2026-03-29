# TODO

hitl-metrics の開発タスクを管理する。完了したら CHANGELOG.md に記録して削除する。

## 実装待ち

（なし）

## 未着手

## 検討中

- Bash コマンドのコンテキスト消費監視
  - `PostToolUse` Hook（`posttooluse-track.sh`）で Bash コマンドの stdout サイズを記録
  - redirect-to-tools をすり抜けた正当な Bash コマンドのうち、出力が大きいものを特定
  - 定期集計で「常連犯」コマンドを可視化し、対策要否を判断する
- retro-pr との連携
  - PR の下位・上位10%ずつは自動で retro-pr 実行
  - 結果を PR と関連付けて表示

## 進行中

- ADR-021: hooks の Shell スクリプトを Go サブコマンドに統一
  - 関連 ADR: [ADR-021](docs/adr/021-migrate-shell-hooks-to-go-subcommands.md)
  - [x] 全 hook が `hitl-metrics hook <event-name>` サブコマンドとして実行できる
  - [x] `permission-log` と `pretooluse-track` の tool annotation ロジックが共通関数に統合される
  - [x] `todo-cleanup-check` の TODO パース処理に Go テストが存在する
  - [x] `hitl-metrics install` が Go サブコマンド形式で `settings.json` に登録できる
  - [x] `internal/install/embed.go` と Shell スクリプトファイルが削除される
  - [x] `docs/architecture.md` が更新される
