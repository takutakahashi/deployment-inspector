# deployment-inspector

A Helm chart for deployment-inspector - tool to inspect Kubernetes deployments and run jobs on their nodes

## Installation

```bash
helm install deployment-inspector oci://ghcr.io/takutakahashi/charts/deployment-inspector --version v0.1.0
```

## Configuration

The following table lists the configurable parameters of the deployment-inspector chart and their default values.

| Parameter | Description | Default |
| --------- | ----------- | ------- |
| `cronjob.enabled` | Enable/disable the CronJob | `true` |
| `cronjob.schedule` | Schedule for the CronJob | `"0 * * * *"` |
| `cronjob.ttlSecondsAfterFinished` | TTL for automatic job cleanup | `3600` |
| `deploymentInspector.command` | Command to run: "list" or "run-job" | `"list"` |
| `deploymentInspector.deploymentName` | Target deployment name | `""` |
| `deploymentInspector.namespace` | Target deployment namespace | `"default"` |
| `deploymentInspector.job.namePrefix` | Job name prefix | `"inspector-job"` |
| `deploymentInspector.job.namespace` | Job namespace | `""` |
| `deploymentInspector.job.image` | Job container image | `"busybox"` |
| `deploymentInspector.job.command` | Job command | `[]` |

## Examples

### List pods from a deployment every hour

```yaml
deploymentInspector:
  command: "list"
  deploymentName: "my-app"
  namespace: "production"
```

### Run a job on nodes where deployment pods are running

```yaml
deploymentInspector:
  command: "run-job"
  deploymentName: "my-app"
  namespace: "production"
  job:
    namePrefix: "node-inspector"
    image: "alpine"
    command: ["sh", "-c", "echo 'Running on node: $HOSTNAME'"]
```