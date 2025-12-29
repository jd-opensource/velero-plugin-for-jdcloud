# Backup Storage Location

The following sample JDCloud `BackupStorageLocation` YAML shows all of the configurable parameters. The items under `spec.config` can be provided as key-value pairs to the `velero install` command's `--backup-location-config` flag -- for example, `endpoint=https://s3.cn-north-1.jdcloud-oss.com,region=cn-east-1,...`.

```yaml
apiVersion: velero.io/v1
kind: BackupStorageLocation
metadata:
  labels:
    component: velero
  name: default
  namespace: velero
spec:
  config:
    bucket: my-bucket
    credentialsFile: /credentials/cloud
    endpoint: https://s3.cn-north-1.jdcloud-oss.com
    insecureSkipTLSVerify: "true"
    prefix: ""
    profile: default
    region: cn-north-1
    s3ForcePathStyle: "true"
  default: true
  objectStorage:
    bucket: my-bucket
    prefix: default/velero-bak
  provider: jdcloud.com/jdcloud
```