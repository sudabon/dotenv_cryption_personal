# dotenv_cryption_personal

個人開発向けの `.env` 暗号化・復号 CLI です。AES-256-GCM のエンベロープ暗号化と ENVC バイナリフォーマットは既存の `dotenv_cryption` と互換を保ちつつ、マスターキーの保管先を AWS Systems Manager Parameter Store に限定しています。

## 特徴

- AES-256-GCM による認証付き暗号化
- ファイルごとにランダムな 32 バイトのデータキーを生成するエンベロープ暗号化
- AWS Systems Manager Parameter Store の `SecureString` によるマスターキー管理
- `dotenv.yaml` による AWS 専用のシンプルな設定
- `encrypt`, `decrypt`, `create master`, `delete master` コマンド

## インストール

### ローカルビルド

`dotenv_cryption` と同様に、ルートでビルドしてインストールすれば `envcrypt` をコマンドラインアプリケーションとして扱えます。

```bash
go build -o envcrypt .
install -m 0755 envcrypt /usr/local/bin/envcrypt
envcrypt version
```

### GitHub Releases からインストール

タグ付きリリース以降は、GitHub Releases から利用している環境に対応した tarball をダウンロードしてインストールできます。

- Apple Silicon Mac: `envcrypt_<version>_darwin_arm64.tar.gz`
- Intel Mac: `envcrypt_<version>_darwin_amd64.tar.gz`
- Linux x86_64: `envcrypt_<version>_linux_amd64.tar.gz`

Releases: `https://github.com/sudabon/dotenv_cryption_personal/releases`

```bash
VERSION=v0.1.0
OS=darwin
ARCH=arm64

curl -LO "https://github.com/sudabon/dotenv_cryption_personal/releases/download/${VERSION}/envcrypt_${VERSION}_${OS}_${ARCH}.tar.gz"
tar -xzf "envcrypt_${VERSION}_${OS}_${ARCH}.tar.gz"
install -m 0755 envcrypt /usr/local/bin/envcrypt
envcrypt version
```

## セットアップ

### 1. 設定ファイルを作成

プロジェクトルートに `dotenv.yaml` を作成します。サンプルは `dotenv.example.yaml` にあります。

```yaml
aws:
  region: ap-northeast-1
  parameter_name: /personal/envcrypt/master-key

crypto:
  algorithm: aes-256-gcm

files:
  encrypted_prefix: ""
```

設定項目:

| フィールド | 説明 | 必須 |
|---|---|---|
| `aws.region` | AWS リージョン | Yes |
| `aws.parameter_name` | マスターキーを保存する Parameter Store 名 | Yes |
| `crypto.algorithm` | 暗号アルゴリズム (`aes-256-gcm` 固定) | No |
| `files.encrypted_prefix` | 暗号化ファイル名に付けるプレフィックス | No |

### 2. AWS 認証を設定

CLI は標準の AWS SDK 認証チェーンを使います。最低でも次のどちらかを設定してください。

```bash
aws configure
```

または:

```bash
export AWS_PROFILE=your-profile
```

必要な IAM 権限:

- `ssm:GetParameter`
- `ssm:PutParameter`
- `ssm:DeleteParameter`

`SecureString` はデフォルトの AWS マネージド KMS キーを前提にしています。

### 3. マスターキーを作成

```bash
envcrypt create master
```

指定した Parameter Store 名に、base64 エンコード済み 32 バイト鍵を `SecureString` として新規作成します。既存パラメータは上書きしません。

### 4. `.gitignore` を確認

平文の `.env` は Git に含めない運用を推奨します。

```gitignore
.env
```

## 使い方

### 暗号化

```bash
# デフォルト (.env -> .env.enc)
envcrypt encrypt

# ファイル指定
envcrypt encrypt --file .env.production
```

### 復号

```bash
# デフォルト (.env.enc -> .env)
envcrypt decrypt

# ファイル指定
envcrypt decrypt --file .env.production.enc
```

### マスターキーの作成

```bash
envcrypt create master
```

### マスターキーの削除

```bash
envcrypt delete master
```

## 出力ファイル名

- `files.encrypted_prefix` 未設定: `.env` -> `.env.enc`
- `files.encrypted_prefix: "enc."`: `.env` -> `enc..env`

復号時は、設定したプレフィックス形式か `.enc` サフィックス形式のどちらかに一致するファイル名のみ自動復元できます。

## ENVC フォーマット

暗号化ファイルは次のバイナリ構造で保存されます。

```text
MAGIC(4B)  VERSION(1B)  NONCE_LEN(1B)  WRAPPED_KEY_LEN(2B)  NONCE  WRAPPED_KEY  CIPHERTEXT
ENVC       0x01         12             variable              ...    ...          ...
```

## 既存 `dotenv_cryption` からの移行

Secrets Manager ベースの `dotenv_cryption` から移行する方法は 2 つあります。

1. Parameter Store に新しいマスターキーを作成し、既存の暗号化ファイルをすべて再暗号化する
2. 既存ツールで使っている 32 バイトのマスターキーそのものを base64 化し、同じ鍵素材として Parameter Store の `SecureString` に登録してから切り替える

後者を選ぶ場合は、両ツールで同じ 32 バイト鍵を使っている限り ENVC ファイルを相互に復号できます。

## 開発

```bash
gofmt -w $(rg --files -g '*.go')
env GOTOOLCHAIN=local go mod tidy
env GOTOOLCHAIN=go1.25.8 go test ./...
env GOTOOLCHAIN=go1.25.8 go vet ./...
```

ビルド:

```bash
go build -o envcrypt .
```

リリース tarball のローカル生成:

```bash
make release-snapshot
ls dist/
```

`v0.1.0` のようなタグを push すると、`.github/workflows/release.yaml` が GitHub Release と tarball 一式を作成します。

## ライセンス

MIT
