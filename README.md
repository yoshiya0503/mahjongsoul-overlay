# mahjongsoul-overlay

[![Test](https://github.com/yoshiya0503/mahjongsoul-overlay/actions/workflows/test.yml/badge.svg)](https://github.com/yoshiya0503/mahjongsoul-overlay/actions/workflows/test.yml)
[![Deploy](https://github.com/yoshiya0503/mahjongsoul-overlay/actions/workflows/fly-deploy.yml/badge.svg)](https://github.com/yoshiya0503/mahjongsoul-overlay/actions/workflows/fly-deploy.yml)
![Go Version](https://img.shields.io/github/go-mod/go-version/yoshiya0503/mahjongsoul-overlay)
[![License](https://img.shields.io/github/license/yoshiya0503/mahjongsoul-overlay)](LICENSE)

雀魂 (Mahjongsoul) の対局情報をリアルタイムに表示する配信用オーバーレイです。Chrome拡張がゲーム内WebSocket通信を傍受し、サーバー経由でOBSなどに表示できるブラウザソースとして動作します。

## 機能

- 現在の局・本場の表示
- プレイヤー名・スコア・順位のリアルタイム表示
- 和了/流局時の点数変動トースト
- 段位ポイントのプログレスバー表示
- セッション成績（対局数・平均順位・1位率・pt増減）

## アーキテクチャ

```
雀魂 (ブラウザ)
  │  Chrome拡張がprotobuf.jsのdecodeをフック
  │  WebSocket
  ▼
mahjongsoul-overlay サーバー (Go / Fiber)
  │  WebSocket
  ▼
OBS ブラウザソース (overlay)
```

## セットアップ

### サーバー

```bash
go build -o mahjongsoul-overlay .
./mahjongsoul-overlay
```

デフォルトで `http://localhost:8787` で起動します。

### Chrome拡張

1. `chrome://extensions` を開き、デベロッパーモードを有効にする
2. 「パッケージ化されていない拡張機能を読み込む」で `extension/` ディレクトリを選択
3. 雀魂のページを開く（またはリロードする）

### OBS

ブラウザソースに `http://localhost:8787` を追加してください。

## 設定

[Viper](https://github.com/spf13/viper) による設定管理に対応しています。設定ファイル、環境変数のいずれからでも設定可能です。

**優先順位**: 環境変数 > config.yaml > デフォルト値

### 設定ファイル

`./config.yaml` または `~/.mahjongsoul-overlay/config.yaml` に配置します。

```yaml
server:
  addr: ":8787"
game:
  session_file: "session.json"
  initial_score: 25000
```

### 環境変数

プレフィックス `MSO_` を付けて、キーの `.` を `_` に置換します。

| 設定キー | 環境変数 | デフォルト値 | 説明 |
|---|---|---|---|
| `server.addr` | `MSO_SERVER_ADDR` | `:8787` | サーバーのリッスンアドレス |
| `game.session_file` | `MSO_GAME_SESSION_FILE` | `session.json` | セッションデータの保存先 |
| `game.initial_score` | `MSO_GAME_INITIAL_SCORE` | `25000` | 対局開始時の初期持ち点 |

## デプロイ

[Fly.io](https://fly.io) へのデプロイに対応しています。`main` ブランチへのpushで自動デプロイされます。

```bash
flyctl deploy
```

## 開発

```bash
# テスト
go test ./... -v -race

# ビルド
go build -o mahjongsoul-overlay .
```

## プロジェクト構成

```
├── main.go                 # エントリーポイント
├── pkg/
│   ├── config/             # 設定管理 (Viper)
│   ├── game/               # ゲーム状態管理
│   ├── handler/            # WebSocket / API ハンドラ
│   └── models/             # データモデル
├── extension/              # Chrome拡張 (WebSocketフック)
├── public/                 # オーバーレイUI (HTML/CSS/JS)
├── Dockerfile
└── fly.toml
```
