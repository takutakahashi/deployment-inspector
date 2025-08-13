# K8s Pod Node Job

Kubernetesクラスター内でDeploymentが作成したPodを一覧表示し、それらのPodが実行されているノード上でJobを起動するツールです。

## ビルド

```bash
mise exec -- go build -o k8s-pod-node-job main.go
```

## 使い方

### 1. Podとノードの一覧表示

```bash
./k8s-pod-node-job list <deployment-name> [-n namespace]
```

例:
```bash
./k8s-pod-node-job list nginx-deployment -n production
```

### 2. 全ノードでJobを起動

```bash
./k8s-pod-node-job run-job <deployment-name> <job-name> [-n namespace] [-i image] [-c command]
```

例:
```bash
# デフォルトコマンドでJobを実行
./k8s-pod-node-job run-job nginx-deployment cleanup-job -n production

# カスタムイメージとコマンドでJobを実行
./k8s-pod-node-job run-job nginx-deployment cleanup-job -n production -i alpine:latest -c "ls,-la,/tmp"
```

## 認証

- クラスター内で実行する場合: InClusterConfigを自動的に使用
- クラスター外で実行する場合: `~/.kube/config`を使用

## オプション

- `-n, --namespace`: Kubernetesネームスペース (デフォルト: default)
- `-i, --image`: Jobで使用するコンテナイメージ (デフォルト: busybox)
- `-c, --command`: Jobで実行するコマンド (カンマ区切り)