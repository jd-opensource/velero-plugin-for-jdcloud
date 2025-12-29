/*
Copyright the Velero contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bufio"
	"context"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	endpointKey         = "endpoint"
	regionKey           = "region"
	s3ForcePathStyleKey = "s3ForcePathStyle"
	bucketKey           = "bucket"
	credentialsFileKey  = "credentialsFile"
)

type ObjectStore struct {
	log       logrus.FieldLogger
	s3Client  *s3.S3
	s3Manager *s3manager.Uploader
}

func newObjectStore(logger logrus.FieldLogger) *ObjectStore {
	return &ObjectStore{log: logger}
}

func getMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (o *ObjectStore) Init(config map[string]string) error {
	o.log.Infof("ObjectStore Init called with config keys: %v", getMapKeys(config))
	o.log.Infof("Config values: region=%s, endpoint=%s, bucket=%s",
		config[regionKey], config[endpointKey], config[bucketKey])

	var (
		region              = config[regionKey]
		endpoint            = config[endpointKey]
		s3ForcePathStyleVal = config[s3ForcePathStyleKey]
		s3ForcePathStyle    bool
		err                 error
	)

	if s3ForcePathStyleVal != "" {
		if s3ForcePathStyle, err = strconv.ParseBool(s3ForcePathStyleVal); err != nil {
			return errors.Wrapf(err, "could not parse %s (expected bool)", s3ForcePathStyleKey)
		}
	} else {
		// Default to true for many S3 compatible stores if not specified, but let's stick to standard false default or user input
		// JD Cloud usually works well with path style
		s3ForcePathStyle = true
	}

	cfg, err := newConfigBuilder(o.log).
		WithRegion(region).
		WithEndpoint(endpoint).
		WithMaxRetries(5).
		WithTLSSettings(s3ForcePathStyle).
		Build()
	if err != nil {
		return errors.WithStack(err)
	}

	// For JD Cloud, we usually expect an endpoint.
	// If standard AWS region handling is desired, we can keep the AWS logic,
	// but mostly we rely on the custom endpoint for S3 compatible plugins.
	if endpoint == "" && region == "" {
		return errors.New("region or s3Url must be specified")
	}

	accessKey := os.Getenv("JDCLOUD_OSS_ACCESS_KEY")
	secretKey := os.Getenv("JDCLOUD_OSS_SECRET_KEY")

	if accessKey == "" || secretKey == "" {
		// Try to read credentials from file if provided in config
		credFile := config[credentialsFileKey]
		if credFile != "" {
			var err error
			accessKey, secretKey, err = readOssCredentialsFromFile(credFile)
			if err != nil {
				return errors.Wrapf(err, "failed to read credentials from file %s", credFile)
			}
		}
	}

	if accessKey == "" || secretKey == "" {
		return errors.New("missing credentials (JDCLOUD_OSS_ACCESS_KEY, JDCLOUD_OSS_SECRET_KEY) or valid credentials file")
	}

	// Create client
	// Log the URL to verify protocol
	o.log.Infof("Creating S3 client with URL: %s, Region: %s, PathStyle: %v", endpoint, region, s3ForcePathStyle)

	client, s3Manager, err := newS3(accessKey, secretKey, cfg)
	if err != nil {
		return errors.WithStack(err)
	}
	o.s3Client = client
	o.s3Manager = s3Manager

	return nil
}

func (o *ObjectStore) PutObject(bucket, key string, body io.Reader) error {
	o.log.Infof("Starting PutObject for bucket: %s, key: %s", bucket, key)
	if _, ok := body.(io.ReadSeeker); !ok {
		o.log.Debugf("Body for %s is not an io.ReadSeeker, using s3manager Uploader", key)
	}

	input := &s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   body,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Minute) // Increased timeout for large files
	defer cancel()

	_, err := o.s3Manager.UploadWithContext(ctx, input)
	if err != nil {
		return errors.Wrapf(err, "error putting object %s", key)
	}

	o.log.Infof("Successfully uploaded object %s", key)
	return nil
}

func readOssCredentialsFromFile(filepath string) (string, string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	var accessKey, secretKey string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Strip quotes if present
		value = strings.Trim(value, `"'`)

		switch key {
		case "JDCLOUD_OSS_ACCESS_KEY":
			accessKey = value
		case "JDCLOUD_OSS_SECRET_KEY":
			secretKey = value
		}
	}

	if err := scanner.Err(); err != nil {
		return "", "", err
	}

	return accessKey, secretKey, nil
}

func (o *ObjectStore) ObjectExists(bucket, key string) (bool, error) {
	log := o.log.WithFields(
		logrus.Fields{
			"bucket": bucket,
			"key":    key,
		},
	)

	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	log.Debug("Checking if object exists")
	if _, err := o.s3Client.HeadObject(input); err != nil {
		log.Debug("Checking for JDCloud S3 specific error information")
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket, s3.ErrCodeNoSuchKey, "NotFound":
				return false, nil
			}
		}

		// S3 compatible might return 404 which sdk translates to NotFound
		// or generic error.
		// For now assume if HeadObject fails with 404 it's not found.
		return false, errors.WithStack(err)
	}

	log.Debug("Object exists")
	return true, nil
}

func (o *ObjectStore) GetObject(bucket, key string) (io.ReadCloser, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	output, err := o.s3Client.GetObject(input)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting object %s", key)
	}

	return output.Body, nil
}

func (o *ObjectStore) ListCommonPrefixes(bucket, prefix, delimiter string) ([]string, error) {
	req := &s3.ListObjectsV2Input{
		Bucket:    &bucket,
		Prefix:    &prefix,
		Delimiter: &delimiter,
	}

	var ret []string
	err := o.s3Client.ListObjectsV2Pages(req, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, prefix := range page.CommonPrefixes {
			ret = append(ret, *prefix.Prefix)
		}
		return !lastPage
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return ret, nil
}

func (o *ObjectStore) ListObjects(bucket, prefix string) ([]string, error) {
	req := &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: &prefix,
	}

	var ret []string
	err := o.s3Client.ListObjectsV2Pages(req, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			ret = append(ret, *obj.Key)
		}
		return !lastPage
	})

	if err != nil {
		return nil, errors.WithStack(err)
	}

	// ensure that returned objects are in a consistent order so that the deletion logic deletes the objects before
	// the pseudo-folder prefix object for s3 providers (such as Quobyte) that return the pseudo-folder as an object.
	// See https://github.com/vmware-tanzu/velero/pull/999
	sort.Sort(sort.Reverse(sort.StringSlice(ret)))

	return ret, nil
}

func (o *ObjectStore) DeleteObject(bucket, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := o.s3Client.DeleteObject(input)

	return errors.Wrapf(err, "error deleting object %s", key)
}

func (o *ObjectStore) CreateSignedURL(bucket, key string, ttl time.Duration) (string, error) {
	input := &s3.GetObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)}

	req, _ := o.s3Client.GetObjectRequest(input)

	urlStr, err := req.Presign(ttl)
	if err != nil {
		return "", errors.WithStack(err)
	}

	o.log.Infof("Successfully created signed URL: %s", urlStr)

	return urlStr, nil
}
