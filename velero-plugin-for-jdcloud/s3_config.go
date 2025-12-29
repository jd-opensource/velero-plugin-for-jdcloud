package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type configBuilder struct {
	log    logrus.FieldLogger
	config *aws.Config
}

func newConfigBuilder(logger logrus.FieldLogger) *configBuilder {
	return &configBuilder{
		log:    logger,
		config: &aws.Config{},
	}
}

func (cb *configBuilder) WithRegion(region string) *configBuilder {
	cb.log.Infof("WithRegion called. Region: %s", region)
	if region != "" {
		cb.config.Region = aws.String(region)
	}
	return cb
}

func (cb *configBuilder) WithEndpoint(ep string) *configBuilder {
	cb.log.Infof("WithEndpoint called. Endpoint: %s", ep)
	if ep != "" {
		cb.config.Endpoint = aws.String(ep)
	}
	return cb
}

func (cb *configBuilder) WithTLSSettings(insecureSkipTLSVerify bool) *configBuilder {
	cb.log.Infof("WithTLSSettings called. InsecureSkipTLSVerify: %v", insecureSkipTLSVerify)
	if insecureSkipTLSVerify {
		cb.config.DisableSSL = aws.Bool(true)
	} else {
		cb.config.DisableSSL = aws.Bool(false)
	}
	return cb
}

func (cb *configBuilder) WithMaxRetries(retries int) *configBuilder {
	cb.log.Infof("WithMaxRetries called. Retries: %d", retries)
	if retries != 0 {
		cb.config.MaxRetries = aws.Int(retries)
	} else {
		cb.config.MaxRetries = aws.Int(3)
	}

	return cb
}

func (cb *configBuilder) S3ForcePathStyle(s3ForcePathStyle bool) *configBuilder {
	cb.log.Infof("S3ForcePathStyle called. s3ForcePathStyle: %v", s3ForcePathStyle)
	cb.config.S3ForcePathStyle = aws.Bool(s3ForcePathStyle)

	return cb
}

func (cb *configBuilder) Build() (*aws.Config, error) {
	cb.log.Infof("configBuilder Build called.")

	return cb.config, nil
}

func newS3(ak, sk string, s3Config *aws.Config) (*s3.S3, *s3manager.Uploader, error) {
	if s3Config == nil {
		return nil, nil, errors.New("s3Config is nil")
	}

	logrus.Infof("new S3Client called. Endpoint: %s, Region: %s", *s3Config.Endpoint, *s3Config.Region)

	staticCredentials := credentials.NewStaticCredentials(ak, sk, "")
	_, err := staticCredentials.Get()
	if err != nil {
		return nil, nil, err
	}

	s3Config.Credentials = staticCredentials

	var newSession *session.Session
	newSession, err = session.NewSession(s3Config)
	if err != nil {
		return nil, nil, err
	}

	uploader := s3manager.NewUploader(newSession)

	return s3.New(newSession), uploader, nil
}
