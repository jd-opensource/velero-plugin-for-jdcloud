# Velero Deployment

<h1 align="center">
  <a href="https://github.com/jd-opensource/velero-plugin-for-jdcloud">
    <img src="docs/images/logo.svg" alt="Logo" width="125" height="125">
  </a>
</h1>

<div align="center">
  Velero Plugin For JDCloud
  <br />
  <br />
  <a href="https://github.com/jd-opensource/velero-plugin-for-jdcloud/issues/new?assignees=&labels=bug&template=01_BUG_REPORT.md&title=bug%3A+">Report a Bug</a>
  ·
  <a href="https://github.com/jd-opensource/velero-plugin-for-jdcloud/issues/new?assignees=&labels=enhancement&template=02_FEATURE_REQUEST.md&title=feat%3A+">Request a Feature</a>
  .
  <a href="https://github.com/jd-opensource/velero-plugin-for-jdcloud/discussions">Ask a Question</a>
</div>

<div align="center">
<br />

[![license](https://img.shields.io/github/license/jd-opensource/joylive-agent.svg?style=flat-square)](LICENSE)
[![PRs welcome](https://img.shields.io/badge/PRs-welcome-ff69b4.svg?style=flat-square)](https://github.com/jd-opensource/velero-plugin-for-jdcloud/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22)
[![made with hearth by JD](https://img.shields.io/badge/made%20with%20%E2%99%A5%20by-dec0dOS-ff1414.svg?style=flat-square)](https://github.com/jd-opensource)
[![CodeQL Advanced](https://github.com/jd-opensource/velero-plugin-for-jdcloud/actions/workflows/codeql.yml/badge.svg)](https://github.com/jd-opensource/velero-plugin-for-jdcloud/actions/workflows/codeql.yml)

</div>

## Overview

This plugin is used for saving and retrieving backup data on [JDCloud OSS](https://oss-console.jdcloud.com/) as an object storage plugin. The backup includes metadata files of Kubernetes resources and CSI objects, as well as the progress of asynchronous operations. It is also used to store result data from backups and restores, including log files, warning/error files, and more.

## Prerequisites

- JKE Cluster or other K8S Cluster
- Git
- Docker
- Docker-buildx

## Deployment

1. Log in to a node in the JKE Cluster or other K8S Cluster and connect to the cluster.

2. Install the Velero CLI tool.

   Download `velero-v1.17.1-linux-amd64.tar.gz` from the official website and upload it to the node.

   Extract Velero:

   ```sh
   tar -xvf velero-v1.17.1-linux-amd64.tar.gz
   mv velero-v1.17.1-linux-amd64/velero /usr/local/bin/velero
	```

3. Build the Velero plugin image.

   Clone the Velero plugin repository:

   ```sh
   git clone https://github.com/jd-opensource/velero-plugin-for-jdcloud.git
	```

   Build the image:

   ```sh
   cd velero-plugin-for-jdcloud
   make docker-build IMAGE="my.hub.com/base/velero-plugin-for-jdcloud" VERSION="v1.0.0"
   ```

4. Create a new OSS bucket. The bucket should not contain other directories.

5. Create a credentials configuration file.

   ```sh
   mkdir velero
   vi /home/velero/credentials-velero
   ```

   Add the following content:

   ```sh
   JDCLOUD_OSS_ACCESS_KEY=<Your AK>
   JDCLOUD_OSS_SECRET_KEY=<Your SK>
   ```

6. Install Velero using the CLI tool.

   ```sh
   velero install \
       --provider jdcloud.com/jdcloud \
       --plugins my.hub.com/base/velero-plugin-for-jdcloud:v1.0.0 \
       --bucket <Your Bucket> \
       --prefix "default-prefix/" \
       --secret-file /home/velero/credentials-velero \
       --backup-location-config \
         endpoint=<Your Endpoint>,region="cn-north-1",s3ForcePathStyle="true",bucket="my-bucket",profile="default",insecureSkipTLSVerify="true",credentialsFile="/credentials/cloud"
   ```

   Replace `<Your Bucket>` and `<Your Endpoint>` with your bucket name and endpoint.

   Enter the K8S Cluster, upgrade the Velero workload, replace the image with the one uploaded to JD Image Serve.

7. Check the storage location status.

   ```sh
   velero backup-location get
   ```

   `available` indicates a successful installation.

8. Backup data.

   Backup all resources in the `default` namespace:

   ```sh
   velero backup create default-backup --include-namespaces default
   ```

   Check if the backup was successful:

   ```sh
   velero backup get
   ```

   `completed` indicates the backup was successful.

   View details and logs:

   ```sh
   velero backup describe default-backup
   velero backup logs default-backup
   ```

9. Delete resources in the `default` namespace.

10. Restore resources.

    Update the backup storage location to read-only mode:

    ```sh
    kubectl patch backupstoragelocation default \
        --namespace velero \
        --type merge \
        --patch '{"spec":{"accessMode":"ReadOnly"}}'
    ```

    Restore resources using the created backup:

    ```sh
    velero restore create --from-backup default-backup
    ```

    Check the restore status:

    ```sh
    velero restore get
    ```

    `completed` indicates a successful restore.

    Restore the backup storage location to read-write mode:

    ```sh
    kubectl patch backupstoragelocation default \
        --namespace velero \
        --type merge \
        --patch '{"spec":{"accessMode":"ReadWrite"}}'
    ```

11. Delete a backup.

    ```sh
    velero backup delete default-backup
    ```

12. Uninstall Velero resources.

    ```sh
    kubectl delete namespace/velero clusterrolebinding/velero
    kubectl delete crds -l component=velero
    ```

## Contributors

<a href="https://github.com/jd-opensource/velero-plugin-for-jdcloud/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=jd-opensource/velero-plugin-for-jdcloud" />
</a>

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=jd-opensource/velero-plugin-for-jdcloud&type=date&legend=top-left)](https://www.star-history.com/#jd-opensource/velero-plugin-for-jdcloud&type=date&legend=top-left)