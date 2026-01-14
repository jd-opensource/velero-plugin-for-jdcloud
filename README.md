# Velero Plugin For JDCloud

<h1 align="center">
  <a href="https://github.com/jd-opensource/velero-plugin-for-jdcloud">
    <img src="docs/images/logo.svg" alt="Velero Plugin for JDCloud Logo" width="125" height="125">
  </a>
</h1>

<div align="center">
  <h2>Velero Plugin For JDCloud</h2>
  <p><strong>Seamless Kubernetes Backup & Restore with JDCloud Object Storage</strong></p>
  <br />
  <a href="https://github.com/jd-opensource/velero-plugin-for-jdcloud/issues/new?assignees=&labels=bug&template=01_BUG_REPORT.md&title=bug%3A+">🐛 Report a Bug</a>
  ·
  <a href="https://github.com/jd-opensource/velero-plugin-for-jdcloud/issues/new?assignees=&labels=enhancement&template=02_FEATURE_REQUEST.md&title=feat%3A+">✨ Request a Feature</a>
  ·
  <a href="https://github.com/jd-opensource/velero-plugin-for-jdcloud/discussions">💬 Ask a Question</a>
</div>

<div align="center">
<br />

[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/jd-opensource/velero-plugin-for-jdcloud/badge)](https://scorecard.dev/viewer/?uri=github.com/jd-opensource/velero-plugin-for-jdcloud)
[![license](https://img.shields.io/github/license/jd-opensource/joylive-agent.svg?style=flat-square)](LICENSE)
[![PRs welcome](https://img.shields.io/badge/PRs-welcome-ff69b4.svg?style=flat-square)](https://github.com/jd-opensource/velero-plugin-for-jdcloud/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22)
[![made with hearth by JD](https://img.shields.io/badge/made%20with%20%E2%99%A5%20by-JDCloud-ff1414.svg?style=flat-square)](https://github.com/jd-opensource)
[![CodeQL Advanced](https://github.com/jd-opensource/velero-plugin-for-jdcloud/actions/workflows/codeql.yml/badge.svg)](https://github.com/jd-opensource/velero-plugin-for-jdcloud/actions/workflows/codeql.yml)

</div>

## 📑 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Deployment](#deployment)
- [Troubleshooting](#troubleshooting)
- [Contributors](#contributors)
- [Star History](#star-history)

## Overview

This plugin is used for saving and retrieving backup data on [JDCloud OSS](https://oss-console.jdcloud.com/) as an object storage plugin. The backup includes metadata files of Kubernetes resources and CSI objects, as well as the progress of asynchronous operations. It is also used to store result data from backups and restores, including log files, warning/error files, and more.

## Features

- **JDCloud OSS Integration**: Seamless integration with JDCloud Object Storage Service
- **Kubernetes Backup**: Full support for Kubernetes resource backup and restore
- **CSI Support**: Compatible with Container Storage Interface (CSI) snapshots
- **Asynchronous Operations**: Handles long-running backup and restore operations
- **Multi-namespace Support**: Backup and restore across multiple Kubernetes namespaces
- **Secure Credentials**: Secure handling of cloud provider credentials

## Prerequisites

- JKE Cluster or other Kubernetes Cluster
- Git
- Docker
- Docker-buildx
- Velero CLI (v1.17.1 or later)

## Deployment

### Step 1: Prepare Environment

1. Log in to a node in the JKE Cluster or other Kubernetes Cluster and connect to the cluster.

2. Install the Velero CLI tool.

   Download the appropriate Velero CLI binary for your architecture:
   - **Linux AMD64**: [velero-v1.17.1-linux-amd64.tar.gz](https://github.com/vmware-tanzu/velero/releases/download/v1.17.1/velero-v1.17.1-linux-amd64.tar.gz)
   - **Other architectures**: Visit the [official Velero releases page](https://github.com/vmware-tanzu/velero/releases) to find binaries for ARM64, macOS, or Windows
   - Upload the downloaded file to your Kubernetes node

   Extract Velero:

   ```bash
   tar -xvf velero-v1.17.1-linux-amd64.tar.gz
   sudo mv velero-v1.17.1-linux-amd64/velero /usr/local/bin/velero
   ```

### Step 2: Get Plugin Image

You have two options to obtain the plugin image:

#### Option A: Use Pre-built Image (Recommended)

Pull the official image from Docker Hub:

```bash
docker pull jdopensource/velero-plugin-for-jdcloud:dev-v1.0.0
```

> 💡 **Note**: Check the [Docker Hub tags](https://hub.docker.com/r/jdopensource/velero-plugin-for-jdcloud/tags) for the latest version.

#### Option B: Build from Source

1. Clone the Velero plugin repository:

   ```bash
   git clone https://github.com/jd-opensource/velero-plugin-for-jdcloud.git
   cd velero-plugin-for-jdcloud
   ```

2. Build the Docker image:

   ```bash
   make docker-build IMAGE="my.hub.com/base/velero-plugin-for-jdcloud" VERSION="v1.0.0"
   ```

   > 🔧 **Customization**: Replace `my.hub.com/base/velero-plugin-for-jdcloud` with your preferred registry and repository name.

### Step 3: Configure JDCloud OSS

1. **Create a new OSS bucket**:
   - Log in to [JDCloud OSS Console](https://oss-console.jdcloud.com/)
   - Create a new bucket in your preferred region
   - Ensure the bucket is empty (no existing directories or files)

2. **Create a credentials configuration file**:

   ```bash
   mkdir -p /home/velero
   vi /home/velero/credentials-velero
   ```

   Add the following content:

   ```bash
   JDCLOUD_OSS_ACCESS_KEY=<Your Access Key>
   JDCLOUD_OSS_SECRET_KEY=<Your Secret Key>
   ```

### Step 4: Install Velero

Install Velero using the CLI tool:

```bash
velero install \
    --provider jdcloud.com/jdcloud \
    --plugins my.hub.com/base/velero-plugin-for-jdcloud:v1.0.0 \
    --bucket <Your Bucket> \
    --prefix "velero-backups/" \
    --secret-file /home/velero/credentials-velero \
    --backup-location-config \
      endpoint=<Your Endpoint>,region="cn-north-1",s3ForcePathStyle="true",bucket="<Your Bucket>",profile="default",insecureSkipTLSVerify="true",credentialsFile="/credentials/cloud"
```

**Replace the following placeholders**:
- `<Your Bucket>`: Your JDCloud OSS bucket name
- `<Your Endpoint>`: Your JDCloud OSS endpoint (e.g., `https://s3.cn-north-1.jdcloud-oss.com`)

> 📝 **Configuration Notes**:
> - **Private Registry**: If using a private container registry, ensure your Kubernetes cluster can pull images from it. You may need to configure image pull secrets.
> - **Endpoint Format**: Use the correct JDCloud OSS endpoint format, typically `https://s3.<region>.jdcloud-oss.com`
> - **Region Values**: Common regions include `cn-north-1`, `cn-south-1`, `cn-east-1`. Check JDCloud documentation for your specific region.
> - **Bucket Naming**: Bucket names must be globally unique across JDCloud OSS
> - **Prefix**: The prefix `velero-backups/` helps organize backup files in your bucket

### Step 5: Verify Installation

Check the storage location status:

```bash
velero backup-location get
```
The expected output:
```bash
NAME      PROVIDER              BUCKET/PREFIX                   PHASE       LAST VALIDATED                  ACCESS MODE   DEFAULT
default   jdcloud.com/jdcloud   <Your Bucket>/<Your prefix>     Available   2026-01-04 15:48:35 +0800 CST   ReadWrite     true
```

`Available` indicates a successful installation.

### Step 6: Create Your First Backup

1. **Create a test backup** of all resources in the `default` namespace:

   ```bash
   velero backup create default-backup \
     --include-namespaces default \
     --wait
   ```

   The expected output:
   ```bash
   Backup request "default-backup" submitted successfully.
   Waiting for backup to complete. You may safely press ctrl-c to stop waiting - your backup will continue in the background.
   ..........
   Backup completed with status: Completed. You may check for more information using the commands `velero backup describe default-backup` and `velero backup logs default-backup`.
   ```

2. **Check backup status**:

   ```bash
   velero backup get
   ```

   The expected output:
   ```bash
   NAME             STATUS      ERRORS   WARNINGS   CREATED                         EXPIRES   STORAGE LOCATION   SELECTOR
   default-backup   Completed   0        0          2026-01-04 15:52:39 +0800 CST   29d       default            <none>
   ```

   Status meanings:
   - `Completed`: Backup finished successfully
   - `InProgress`: Backup is still running
   - `Failed`: Backup encountered errors

3. **View backup details**:

   ```bash
   # Get detailed information about the backup
   velero backup describe default-backup --details
   
   # View backup logs for troubleshooting
   velero backup logs default-backup
   ```

### Step 7: Restore from Backup

1. **Optional - Test restore**: Delete resources in the `default` namespace to test restore functionality:

   ```bash
   kubectl delete namespace default
   kubectl create namespace default
   ```

2. **Set backup location to read-only mode** (recommended during restore):

   ```bash
   kubectl patch backupstoragelocation default \
       --namespace velero \
       --type merge \
       --patch '{"spec":{"accessMode":"ReadOnly"}}'
   ```

3. **Restore from backup**:

   ```bash
   velero restore create restore-default \
     --from-backup default-backup \
     --wait
   ```

   The expected output:
   ```bash
   Restore request "restore-default" submitted successfully.
   Waiting for restore to complete. You may safely press ctrl-c to stop waiting - your restore will continue in the background.
   ......
   Restore completed with status: Completed. You may check for more information using the commands `velero restore describe restore-default` and `velero restore logs restore-default`.
   ```

4. **Check restore status**:

   ```bash
   velero restore get
   ```

   The expected output:
   ```bash
   NAME              BACKUP           STATUS      STARTED                         COMPLETED                       ERRORS   WARNINGS   CREATED                         SELECTOR
   restore-default   default-backup   Completed   2026-01-04 15:56:38 +0800 CST   2026-01-04 15:56:44 +0800 CST   0        3          2026-01-04 15:56:38 +0800 CST   <none>
   ```

   Status meanings:
   - `Completed`: Restore finished successfully
   - `InProgress`: Restore is still running
   - `Failed`: Restore encountered errors

5. **View restore details**:

   ```bash
   velero restore describe restore-default
   velero restore logs restore-default
   ```

6. **Restore backup location to read-write mode**:

   ```bash
   kubectl patch backupstoragelocation default \
       --namespace velero \
       --type merge \
       --patch '{"spec":{"accessMode":"ReadWrite"}}'
   ```

### Step 8: Manage Backups

#### Delete a Backup

```bash
velero backup delete default-backup
```

> ⚠️ **Warning**: This will permanently delete the backup from storage. Ensure you have other backups or don't need this data.

The expected output:
```bash
Are you sure you want to continue (Y/N)? Y
Request to delete backup "default-backup" submitted successfully.
The backup will be fully deleted after all associated data (disk snapshots, backup files, restores) are removed.
```

#### List All Backups

```bash
velero backup get
```

#### Uninstall Velero

If you need to completely remove Velero from your cluster:

```bash
# Delete Velero namespace and resources
kubectl delete namespace velero
kubectl delete clusterrolebinding velero

# Delete Velero CRDs
kubectl delete crds -l component=velero

# Optional: Clean up remaining resources
kubectl delete clusterroles -l component=velero
```

## Troubleshooting

### Common Issues and Solutions

#### 1. Backup Location Not Available
**Symptoms**: `velero backup-location get` shows status as `Unavailable`

**Solutions**:
- Check credentials file permissions:
  ```bash
  ls -la /home/velero/credentials-velero
  chmod 600 /home/velero/credentials-velero
  ```
- Verify JDCloud OSS credentials are correct and active
- Ensure the bucket exists and is accessible from your cluster
- Check network connectivity to JDCloud OSS endpoint

#### 2. Image Pull Errors
**Symptoms**: Velero pods stuck in `ImagePullBackOff` or `ErrImagePull`

**Solutions**:
- If using a private registry, configure image pull secrets:
  ```bash
  kubectl create secret docker-registry regcred \
    --docker-server=<your-registry-server> \
    --docker-username=<your-username> \
    --docker-password=<your-password> \
    --docker-email=<your-email> \
    -n velero
  ```
- Verify the image name and tag are correct
- Check if the image exists in the registry

#### 3. Permission Issues
**Symptoms**: Backup/restore operations fail with permission errors

**Solutions**:
- Ensure Velero service account has necessary RBAC permissions
- Check if the cluster has sufficient CPU/memory resources
- Verify the node has access to the JDCloud OSS endpoint

#### 4. Backup/Restore Failures
**Symptoms**: Operations fail with various error messages

**Debugging Steps**:
1. Check Velero pod logs:
   ```bash
   kubectl logs -n velero deployment/velero
   ```
2. Check specific backup/restore logs:
   ```bash
   velero backup logs <backup-name>
   velero restore logs <restore-name>
   ```
3. Describe the resource for detailed status:
   ```bash
   velero backup describe <backup-name> --details
   ```

### Getting Help

- 📖 **Documentation**: Check the [official Velero documentation](https://velero.io/docs/)
- 🐛 **Report bugs**: [Create a bug report](https://github.com/jd-opensource/velero-plugin-for-jdcloud/issues/new?assignees=&labels=bug&template=01_BUG_REPORT.md&title=bug%3A+)
- ✨ **Request features**: [Request a new feature](https://github.com/jd-opensource/velero-plugin-for-jdcloud/issues/new?assignees=&labels=enhancement&template=02_FEATURE_REQUEST.md&title=feat%3A+)
- 💬 **Community support**: [Join discussions](https://github.com/jd-opensource/velero-plugin-for-jdcloud/discussions)

## Contributors

<a href="https://github.com/jd-opensource/velero-plugin-for-jdcloud/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=jd-opensource/velero-plugin-for-jdcloud" alt="Contributors" />
</a>

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=jd-opensource/velero-plugin-for-jdcloud&type=date&legend=top-left)](https://www.star-history.com/#jd-opensource/velero-plugin-for-jdcloud&type=date&legend=top-left)
