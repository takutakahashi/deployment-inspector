# Deployment Inspector

Kubernetesクラスター内でDeploymentが作成したPodを一覧表示し、それらのPodが実行されているノード上でJobを起動するツールです。

## パッケージ構造

```
.
├── cmd/
│   └── deployment-inspector/
│       └── main.go          # CLIエントリーポイント
├── pkg/
│   └── k8s/
│       ├── client.go        # Kubernetesクライアント管理
│       ├── client_test.go   
│       ├── deployment.go    # Deployment操作
│       ├── deployment_test.go
│       ├── job.go          # Job操作
│       └── job_test.go
└── go.mod
```

## ビルド

```bash
mise exec -- go build -o deployment-inspector ./cmd/deployment-inspector
```

## テスト

```bash
# 全テストを実行
mise exec -- go test ./...

# 詳細な出力付き
mise exec -- go test -v ./...
```

## 使い方

### 1. Podとノードの一覧表示

```bash
./deployment-inspector list <deployment-name> [-n namespace]
```

例:
```bash
./deployment-inspector list nginx-deployment -n production
```

### 2. 全ノードでJobを起動

```bash
./deployment-inspector run-job <deployment-name> <job-name> [-n namespace] [-i image] [-c command]
```

例:
```bash
# デフォルトコマンドでJobを実行
./deployment-inspector run-job nginx-deployment cleanup-job -n production

# カスタムイメージとコマンドでJobを実行
./deployment-inspector run-job nginx-deployment cleanup-job -n production -i alpine:latest -c "ls,-la,/tmp"
```

## 認証

- クラスター内で実行する場合: InClusterConfigを自動的に使用
- クラスター外で実行する場合: `~/.kube/config`を使用

## オプション

- `-n, --namespace`: Kubernetesネームスペース (デフォルト: default)
- `-i, --image`: Jobで使用するコンテナイメージ (デフォルト: busybox)
- `-c, --command`: Jobで実行するコマンド (カンマ区切り)