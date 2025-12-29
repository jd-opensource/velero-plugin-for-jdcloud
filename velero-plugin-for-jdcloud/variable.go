package main

const (
	volumeDefaultMaxCount = 1
	filterDiskId          = "diskId"
)

var (
	backupTagKey               = "velero.io/backup-name"
	backupTagValue             = "velero"
	snapshotDefaultDescription = "Created by Velero"
)
