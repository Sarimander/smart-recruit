package oss

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"logic-grpc-service/internal/config"
	"logic-grpc-service/internal/pkg/filevalidate"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type Client struct {
	bucket              *oss.Bucket
	uploadExpireSeconds int64
	downloadExpireSeconds int64
}

func New(cfg config.OSSConfig) (*Client, error) {
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("create oss client: %w", err)
	}
	bucket, err := client.Bucket(cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("get bucket: %w", err)
	}
	return &Client{
		bucket:                bucket,
		uploadExpireSeconds:   cfg.UploadExpireSeconds,
		downloadExpireSeconds: cfg.DownloadExpireSeconds,
	}, nil
}

func (c *Client) GenerateUploadURL(userID int64, filename string) (uploadURL, ossKey string, expire int64, err error) {
	if err = filevalidate.ValidateFilename(filename); err != nil {
		return "", "", 0, err
	}
	ext := strings.ToLower(filepath.Ext(filename))
	ossKey = fmt.Sprintf("resumes/%d/%d%s", userID, time.Now().UnixNano(), ext)
	expire = c.uploadExpireSeconds
	uploadURL, err = c.bucket.SignURL(ossKey, oss.HTTPPut, expire)
	if err != nil {
		return "", "", 0, fmt.Errorf("sign upload url: %w", err)
	}
	return uploadURL, ossKey, expire, nil
}

func (c *Client) GenerateDownloadURL(ossKey string) (downloadURL string, expire int64, err error) {
	if err = filevalidate.ValidateOSSKey(ossKey); err != nil {
		return "", 0, err
	}
	expire = c.downloadExpireSeconds
	downloadURL, err = c.bucket.SignURL(ossKey, oss.HTTPGet, expire)
	if err != nil {
		return "", 0, fmt.Errorf("sign download url: %w", err)
	}
	return downloadURL, expire, nil
}
