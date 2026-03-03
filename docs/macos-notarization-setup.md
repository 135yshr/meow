# macOS バイナリ署名・公証（Notarization）セットアップガイド

Homebrew でインストールした `meow` コマンドで Gatekeeper 警告を出さないために、Apple Developer ID による署名と公証を設定する手順。

## 前提条件

- [Apple Developer Program](https://developer.apple.com/programs/) に登録済み（年額 $99）
- macOS マシンで Keychain Access が使える環境

## 1. Developer ID Application 証明書の作成

### 1-1. CSR（証明書署名要求）を作成

1. **Keychain Access.app** を開く
2. メニューバー > 証明書アシスタント > **認証局に証明書を要求...**
3. 以下を入力:
   - ユーザのメールアドレス: Apple Developer に登録したメールアドレス
   - 通称: 任意（空欄でも可）
   - 要求の処理: **ディスクに保存**
4. `CertificateSigningRequest.certSigningRequest` を保存

### 1-2. Apple Developer Portal で証明書を発行

1. [Apple Developer - Certificates](https://developer.apple.com/account/resources/certificates/list) にアクセス
2. 「+」ボタンをクリック
3. **Developer ID Application** を選択して「Continue」
4. 先ほどの CSR ファイルをアップロード
5. 証明書（`.cer`）をダウンロード
6. ダウンロードした `.cer` をダブルクリック → Keychain に自動インストール

### 1-3. .p12 ファイルとしてエクスポート

1. **Keychain Access.app** > ログイン > 証明書
2. **「Developer ID Application: あなたの名前 (TEAM_ID)」** を見つける
3. 右クリック > **「"Developer ID Application: ..."を書き出す...」**
4. フォーマット: **個人情報交換 (.p12)** で保存
5. **パスワードを設定**（後で `MACOS_SIGN_PASSWORD` として使用）

### 1-4. Base64 エンコード

```bash
base64 -i /path/to/certificate.p12 | pbcopy
```

クリップボードにコピーされた文字列が `MACOS_SIGN_P12` の値になる。

## 2. App Store Connect API キーの作成

### 2-1. API キーを発行

1. [App Store Connect - Keys](https://appstoreconnect.apple.com/access/integrations/api) にアクセス
2. 「Team Keys」タブ > 「Generate API Key」
3. 名前: `meow-notarization`（任意）
4. アクセス: **Developer**
5. 「Generate」をクリック

### 2-2. 必要な情報を控える

| 情報 | 場所 | GitHub Secret 名 |
|------|------|-------------------|
| **Key ID** | キー一覧の「Key ID」列（10文字英数字） | `MACOS_NOTARY_KEY_ID` |
| **Issuer ID** | ページ上部に表示される UUID | `MACOS_NOTARY_ISSUER_ID` |
| **.p8 ファイル** | 「Download API Key」ボタン | `MACOS_NOTARY_KEY` |

> **注意:** .p8 ファイルは **1回しかダウンロードできない**。安全な場所に保管すること。

### 2-3. .p8 ファイルを Base64 エンコード

```bash
base64 -i /path/to/AuthKey_XXXXXXXXXX.p8 | pbcopy
```

クリップボードにコピーされた文字列が `MACOS_NOTARY_KEY` の値になる。

## 3. GitHub Secrets の登録

リポジトリの **Settings > Secrets and variables > Actions** で以下の5つを登録する。

| Secret 名 | 値 |
|-----------|-----|
| `MACOS_SIGN_P12` | 手順 1-4 の Base64 文字列 |
| `MACOS_SIGN_PASSWORD` | 手順 1-3 で設定した .p12 のパスワード |
| `MACOS_NOTARY_KEY` | 手順 2-3 の Base64 文字列 |
| `MACOS_NOTARY_KEY_ID` | 手順 2-2 の Key ID |
| `MACOS_NOTARY_ISSUER_ID` | 手順 2-2 の Issuer ID |

## 4. 動作確認

### 4-1. リリースを実行

1. PR をマージ
2. マージした PR に `/release` コメント
3. GitHub Actions のログで以下を確認:
   - `signing` ステップが成功
   - `notarizing` ステップが成功

### 4-2. バイナリの検証

リリースされた darwin バイナリをダウンロードして実行:

```bash
# Homebrew 経由
brew update && brew upgrade meow
meow --version

# または直接ダウンロード
curl -LO https://github.com/135yshr/meow/releases/latest/download/meow_darwin_arm64.tar.gz
curl -LO https://github.com/135yshr/meow/releases/latest/download/checksums.txt
shasum -a 256 -c checksums.txt --ignore-missing
tar xzf meow_darwin_arm64.tar.gz
./meow --version
```

Gatekeeper の警告が表示されなければ成功。

### 4-3. 署名の確認（オプション）

```bash
codesign -dv --verbose=4 $(which meow)
```

`Developer ID Application: ...` が表示されていれば正しく署名されている。

## トラブルシューティング

### GoReleaser で notarize がスキップされる

- GitHub Secrets が正しく設定されているか確認
- Secret 名のタイポがないか確認（`MACOS_SIGN_P12` など）

### 署名エラー: "certificate is not valid"

- Developer ID **Application** 証明書であることを確認（Installer ではない）
- 証明書が失効していないか Apple Developer Portal で確認

### 公証エラー: "authentication failed"

- App Store Connect API キーの権限が **Developer** 以上か確認
- Key ID / Issuer ID が正しいか確認
- .p8 ファイルの Base64 エンコードが正しいか確認（改行が含まれていないこと）

### fork リポジトリで失敗する

Secrets が設定されていない環境では `isEnvSet` により自動でスキップされるため、正常動作する。失敗する場合は GoReleaser のバージョンが v2 以上であることを確認。
