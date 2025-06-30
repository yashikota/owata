# Owata - Discord通知ツール

🔔 クロスプラットフォーム対応のGoツールでDiscord通知を送信します。Claude CodeやGemini CLIなどのLLMが完了通知を送信するのに最適です。  

## 特徴

- 🖥️ **クロスプラットフォーム**: Windows、macOS、Linuxで動作
- 📨 **Discord webhooks**: リッチな埋め込み通知を送信
- ⚙️ **設定可能**: JSONの設定ファイルまたはコマンドライン引数
- 🚀 **ゼロ依存**: Goの標準ライブラリのみを使用
- 🤖 **LLMフレンドリー**: 自動通知のためのシンプルなコマンドラインインターフェイス

## インストール

### ソースからビルド

```bash
git clone https://github.com/yashikota/owata
cd owata
go build -o owata
```

### バイナリをダウンロード

[リリースページ](https://github.com/yashikota/owata/releases)から最新版をダウンロードしてください。

## 使い方

### 基本的な使い方

```bash
# webhook URLで通知を送信
owata "Claude Codeセッション完了" https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN

# タスク完了について通知
owata "タスクが正常に完了しました" https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN

# ソースを指定して通知を送信
owata "タスクが完了しました" --source="Claude Code"
```

### AI/LLMツールでの使用

AIツールやLLMエージェントは、コマンドを実行することで直接CLIを使用できます：

```bash
# どのプログラミング言語からでも
exec("owata 'AIタスクが完了しました' --source='Claude Code'");
```

### 設定ファイルを使用

1. サンプル設定ファイルをコピー:
   ```bash
   cp owata-config.json.example owata-config.json
   ```

2. `owata-config.json`を編集:
   ```json
   {
     "webhook_url": "https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN",
     "username": "Owata",
     "avatar_url": "https://example.com/avatar.png"
   }
   ```

3. 監視を開始:
   ```bash
   owata claude
   ```

## 設定

### 設定ファイルのオプション

- `webhook_url`: Discord Webhook URL（必須）
- `username`: ボットのカスタムユーザー名（オプション、デフォルト: "Owata Monitor"）
- `avatar_url`: ボットのカスタムアバターURL（オプション）

### コマンドライン

```bash
owata <message> [webhook-url] [--source=<source>]
```

- `message`: 送信するメッセージ（必須）
- `webhook-url`: Discord webhook URL（設定ファイルを使用する場合はオプション）
- `--source`: 通知のソースを指定（例: "Claude Code"、"Gemini"など）

## Discord Webhookの設定

1. Discordサーバーの設定画面を開く
2. 連携サービス → Webhookに移動
3. 「新しいWebhook」をクリック
4. チャンネルを選択してWebhook URLをコピー
5. このURLを設定ファイルまたはコマンドラインで使用

## 使用例

```bash
owata "Claude Codeによるタスクが完了しました" --source="Claude Code"
```

```bash
owata "Geminiによるタスクが完了しました" --source="Gemini CLI"
```

```bash
owata "カスタムメッセージ" --source="任意のソース"
```

## 通知フォーマット

Owataは以下の情報を含むDiscord埋め込みメッセージを送信します：
- メッセージテキスト
- 作業ディレクトリ
- ソース（指定されていれば）
- タイムスタンプ
