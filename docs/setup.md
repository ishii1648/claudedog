# セットアップガイド

hitl-metrics を導入して Claude Code の人の介入率を計測・可視化する手順です。

## 前提条件

| ツール | 用途 |
|--------|------|
| Go 1.22+ | CLI ビルド |
| Grafana 11+ | ダッシュボード表示 |
| [frser-sqlite-datasource](https://github.com/fr-ser/grafana-sqlite-datasource) | Grafana の SQLite プラグイン |
| gh CLI | PR URL の自動補完（`backfill` コマンド） |
| Docker（任意） | E2E テスト用の Grafana 環境 |

## 1. CLI のビルド

```fish
git clone https://github.com/ishii1648/hitl-metrics.git
cd hitl-metrics
go build -o ~/.local/bin/hitl-metrics ./cmd/hitl-metrics/
```

`~/.local/bin` が `$PATH` に含まれていることを確認してください。

## 2. Claude Code hook の登録

`~/.claude/settings.json` に以下の hook 設定を追加します。`/path/to/hitl-metrics` はクローン先のパスに置き換えてください。

```json
{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "/path/to/hitl-metrics/hooks/session-index.sh"
          }
        ]
      }
    ],
    "PermissionRequest": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "/path/to/hitl-metrics/hooks/permission-log.sh"
          }
        ]
      }
    ],
    "Stop": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "/path/to/hitl-metrics/hooks/stop.sh"
          }
        ]
      }
    ]
  }
}
```

### hook の役割

| hook | トリガー | 出力先 |
|------|----------|--------|
| `session-index.sh` | セッション開始時 | `~/.claude/session-index.jsonl` |
| `permission-log.sh` | Permission UI 表示時 | `~/.claude/logs/permission.log` |
| `stop.sh` | セッション終了時 | `~/.claude/hitl-metrics.db`（backfill + sync-db） |

登録後、Claude Code で新しいセッションを開始するとデータが記録されます。セッション終了時に PR URL 補完と SQLite DB 同期が自動実行されます。

## 3. データの同期

セッション終了時に Stop hook が `hitl-metrics backfill` → `hitl-metrics sync-db` を自動実行します。

- **backfill**: PR URL 補完・マージ判定・レビューコメント数を `gh` CLI 経由で取得
- **sync-db**: JSONL/log → SQLite 変換（`~/.claude/hitl-metrics.db` を生成）

cursor（`~/.claude/hitl-metrics-state.json`）により前回処理済み以降のエントリのみが走査されるため、高速に完了します。

初回セットアップ時や手動で即時実行する場合:

```fish
hitl-metrics backfill
hitl-metrics sync-db
```

## 4. Grafana ダッシュボードの設定

### 方法 A: ローカル Grafana に手動設定

1. Grafana に [frser-sqlite-datasource](https://github.com/fr-ser/grafana-sqlite-datasource) プラグインをインストール

2. データソースを追加
   - Type: `SQLite`
   - Path: `~/.claude/hitl-metrics.db`（フルパスで指定）

3. ダッシュボードをインポート
   - Grafana の Import 画面で `grafana/dashboards/hitl-metrics.json` をアップロード
   - データソースに上記で作成した SQLite データソースを選択

### 方法 B: プロビジョニングファイルで自動設定

Grafana の設定ディレクトリにプロビジョニングファイルを配置します。

```fish
# データソース設定をコピー（パスを環境に合わせて編集）
cp grafana/provisioning/datasources/hitl-metrics.yaml /etc/grafana/provisioning/datasources/

# ダッシュボード設定をコピー
cp grafana/provisioning/dashboards/hitl-metrics.yaml /etc/grafana/provisioning/dashboards/

# ダッシュボード JSON をコピー
cp -r grafana/dashboards /var/lib/grafana/dashboards/hitl-metrics
```

データソース設定の `path` を自分の環境に合わせて変更してください。

```yaml
# grafana/provisioning/datasources/hitl-metrics.yaml
jsonData:
  path: /Users/<your-username>/.claude/hitl-metrics.db
```

## 5. 日常の運用

Stop hook が登録済みであれば、セッション終了時に `backfill` と `sync-db` が自動実行されます。手動で即時更新する場合:

```fish
hitl-metrics backfill
hitl-metrics sync-db
```

ダッシュボードは `http://localhost:3000`（デフォルト）でアクセスできます。

## E2E テスト環境（開発者向け）

テストデータを使った Grafana 環境を Docker で起動できます。

```fish
make grafana-up          # Grafana + Image Renderer 起動 → http://localhost:13000
make grafana-screenshot  # 全パネルの PNG を取得
make grafana-down        # コンテナ停止
```

## トラブルシューティング

### hook が動作しない

- `~/.claude/settings.json` の hook パスが正しいか確認
- hook スクリプトに実行権限があるか確認: `chmod +x hooks/*.sh`
- デバッグログを確認: `~/.claude/logs/session-index-debug.log`

### sync-db でデータが空になる

- `~/.claude/session-index.jsonl` が存在し、データが記録されているか確認
- `~/.claude/logs/permission.log` が存在するか確認

### Grafana でデータが表示されない

- データソースの Path が `hitl-metrics.db` のフルパスを指しているか確認
- `hitl-metrics sync-db` を再実行して DB を最新化
- Grafana のデータソース設定で「Test」ボタンを押して接続を確認
