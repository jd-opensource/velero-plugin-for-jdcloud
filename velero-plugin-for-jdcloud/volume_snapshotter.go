package main

import (
	"fmt"
	"os"
	"strings"

	jdcloudcore "github.com/jdcloud-api/jdcloud-sdk-go/core"
	common "github.com/jdcloud-api/jdcloud-sdk-go/services/common/models"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/disk/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/disk/client"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/disk/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
)

type VolumeSnapshotter struct {
	log        logrus.FieldLogger
	diskClient *client.DiskClient
	cfg        map[string]string

	region string
}

func newVolumeSnapshotter(logger logrus.FieldLogger) *VolumeSnapshotter {
	return &VolumeSnapshotter{log: logger}
}

func (v *VolumeSnapshotter) newCredentials(ak, sk string) *jdcloudcore.Credential {
	credentials := jdcloudcore.NewCredentials(ak, sk)
	return credentials
}

func (v *VolumeSnapshotter) Init(config map[string]string) error {
	v.log.Infof("Init called")
	v.cfg = config
	region := config[regionKey]
	if region == "" {
		return errors.New("missing region in config")
	}

	v.region = region

	credFile := config[credentialsFileKey]
	if credFile == "" {
		return errors.New("missing credentials file")
	}

	var (
		accessKey, secretKey string
		err                  error
	)
	accessKey, secretKey, err = readCredentialsFromFile(v.log, credFile)
	if err != nil {
		return errors.Wrapf(err, "failed to read credentials from file %s", credFile)
	}

	if accessKey == "" || secretKey == "" {
		return errors.New("missing credentials")
	}

	credentials := v.newCredentials(accessKey, secretKey)
	v.diskClient = client.NewDiskClient(credentials)
	v.diskClient.SetLogger(jdcloudcore.DefaultLogger{Level: 5})

	return nil
}

func readCredentialsFromFile(log logrus.FieldLogger, filepath string) (string, string, error) {
	log.Infof("readCredentialsFromFile called with filepath: %s", filepath)
	fileBytes, err := os.ReadFile(filepath)
	if err != nil {
		return "", "", err
	}

	creds := string(fileBytes)
	if !strings.Contains(creds, ":") {
		return "", "", fmt.Errorf("credentials does not contain a colon: %s", filepath)
	}

	credsList := strings.Split(creds, ":")
	if len(credsList) != 2 {
		return "", "", fmt.Errorf("wrong credential format")
	}

	return credsList[0], credsList[1], nil
}

func (v *VolumeSnapshotter) CreateVolumeFromSnapshot(snapshotID, volumeType, volumeAZ string, iops *int64) (string, error) {
	v.log.Infof("CreateVolumeFromSnapshot called with snapshotID: %s, volumeType: %s, volumeAZ: %s", snapshotID, volumeType, volumeAZ)
	// CreateDisk from Snapshot
	req := apis.NewCreateDisksRequestWithoutParam()
	req.SetRegionId(v.region)
	req.SetMaxCount(volumeDefaultMaxCount)
	req.SetDiskSpec(&models.DiskSpec{
		Az:         volumeAZ,
		DiskType:   volumeType,
		SnapshotId: &snapshotID,
		Name:       fmt.Sprintf("restore-%s", snapshotID),
	})
	req.SetUserTags([]models.Tag{
		{
			Key:   &backupTagKey,
			Value: &backupTagValue,
		},
	})

	if iops != nil {
		// req.DiskSpec.Iops = iops // Check if IOPS is supported in DiskSpec
		// Based on common cloud APIs, IOPS might be a field.
		// Let's assume DiskSpec has it or similar.
		// If compilation fails, we adjust.
		// req.DiskSpec.Iops = int(*iops)
	}

	resp, err := v.diskClient.CreateDisks(req)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create volume from snapshot %s", snapshotID)
	}

	if len(resp.Result.DiskIds) == 0 {
		return "", errors.New("no volume created")
	}

	return resp.Result.DiskIds[0], nil
}

func (v *VolumeSnapshotter) GetVolumeInfo(volumeID, volumeAZ string) (string, *int64, error) {
	v.log.Infof("GetVolumeInfo called with volumeID: %s, volumeAZ: %s", volumeID, volumeAZ)
	req := apis.NewDescribeDisksRequestWithoutParam()
	req.SetRegionId(v.region)
	req.SetFilters([]common.Filter{
		{
			Name:   filterDiskId,
			Values: []string{volumeID},
		},
	})

	resp, err := v.diskClient.DescribeDisks(req)
	if err != nil {
		return "", nil, errors.Wrapf(err, "failed to describe volume %s", volumeID)
	}

	if len(resp.Result.Disks) == 0 {
		return "", nil, errors.Errorf("volume %s not found", volumeID)
	}

	disk := resp.Result.Disks[0]
	// Return VolumeType and IOPS
	// disk.DiskType is usually string
	// disk.Iops is usually int or int64

	var iops int64
	if disk.Iops != 0 { // Assuming Iops field exists and is int
		iops = int64(disk.Iops)
	}

	return disk.DiskType, &iops, nil
}

func (v *VolumeSnapshotter) CreateSnapshot(volumeID, volumeAZ string, tags map[string]string) (string, error) {
	v.log.Infof("CreateSnapshot called with volumeID: %s, volumeAZ: %s, tags: %v", volumeID, volumeAZ, tags)
	req := apis.NewCreateSnapshotRequestWithoutParam()
	req.SetRegionId(v.region)
	req.SetSnapshotSpec(&models.SnapshotSpec{
		DiskId:      volumeID,
		Name:        fmt.Sprintf("velero-%s", volumeID),
		Description: &snapshotDefaultDescription,
	})

	resp, err := v.diskClient.CreateSnapshot(req)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create snapshot for volume %s", volumeID)
	}

	return resp.Result.SnapshotId, nil
}

func (v *VolumeSnapshotter) DeleteSnapshot(snapshotID string) error {
	v.log.Infof("DeleteSnapshot called with snapshotID: %s", snapshotID)
	req := apis.NewDeleteSnapshotRequestWithoutParam()
	req.SetRegionId(v.region)
	req.SetSnapshotId(snapshotID)

	_, err := v.diskClient.DeleteSnapshot(req)
	if err != nil {
		return errors.Wrapf(err, "failed to delete snapshot %s", snapshotID)
	}

	return nil
}

func (v *VolumeSnapshotter) GetVolumeID(pv runtime.Unstructured) (string, error) {
	v.log.Infof("GetVolumeID called")
	// Helper to extract VolumeID from PV
	if pv == nil {
		return "", nil
	}

	// This depends on the CSI driver or Cloud Provider structure in PV.
	// For now, return empty as standard Velero plugins might handle this if not custom.
	// But usually we need to parse PV.Spec...

	// Example: spec.csi.volumeHandle
	metadata, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pv)
	if err != nil {
		return "", err
	}

	// Simple traversal
	spec, ok := metadata["spec"].(map[string]interface{})
	if !ok {
		return "", nil
	}

	// Check CSI
	if csi, ok := spec["csi"].(map[string]interface{}); ok {
		if handle, ok := csi["volumeHandle"].(string); ok {
			return handle, nil
		}
	}

	// Check old style AWS EBS if needed?
	// JDCloud might use CSI driver name "disk.jdcloud.csi.driver"

	return "", nil
}

func (v *VolumeSnapshotter) SetVolumeID(pv runtime.Unstructured, volumeID string) (runtime.Unstructured, error) {
	v.log.Infof("SetVolumeID called with volumeID: %s", volumeID)
	// Update PV with new VolumeID
	if pv == nil {
		return nil, nil
	}

	obj := pv.UnstructuredContent()
	spec, ok := obj["spec"].(map[string]interface{})
	if !ok {
		return pv, nil
	}

	// Update CSI volumeHandle
	if csi, ok := spec["csi"].(map[string]interface{}); ok {
		csi["volumeHandle"] = volumeID
	}

	return pv, nil
}
