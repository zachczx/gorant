package upload

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type BucketFile struct {
	Key          string
	Size         int64
	LastModified time.Time
}

func (f *BucketFile) SizeString() string {
	return strconv.FormatInt(f.Size, 10)
}

func (f *BucketFile) LastModifiedString() string {
	return f.LastModified.Format("2006-01-02 15:04:05")
}

type BucketConfig struct {
	BucketName         string
	AccountID          string
	BaseEndpoint       string
	PublicAccessDomain string
	AccessKeyID        string
	AccessKeySecret    string
}

type LookupFile struct {
	File       multipart.File
	FileID     uuid.UUID
	FileKey    string
	FileStore  string
	FileBucket string
}

type NullFile struct {
	File       multipart.File
	FileID     uuid.NullUUID
	FileKey    sql.NullString
	FileStore  sql.NullString
	FileBucket sql.NullString
}

func NewBucketConfig(options ...func(*BucketConfig)) *BucketConfig {
	config := &BucketConfig{}
	for _, o := range options {
		o(config)
	}
	return config
}

func WithBucketName(bucketName string) func(*BucketConfig) {
	return func(c *BucketConfig) {
		c.BucketName = bucketName
	}
}

func WithBaseEndpoint(baseEndpoint string) func(*BucketConfig) {
	return func(c *BucketConfig) {
		c.BaseEndpoint = baseEndpoint
	}
}

func WithPublicAccessDomain(publicAccessDomain string) func(*BucketConfig) {
	return func(c *BucketConfig) {
		c.PublicAccessDomain = publicAccessDomain
	}
}

func WithAccessKeyID(accessKeyID string) func(*BucketConfig) {
	return func(c *BucketConfig) {
		c.AccessKeyID = accessKeyID
	}
}

func WithAccessKeySecret(accessKeySecret string) func(*BucketConfig) {
	return func(c *BucketConfig) {
		c.AccessKeySecret = accessKeySecret
	}
}

func (bc *BucketConfig) ConnectBucket() (*s3.Client, error) {
	client := &s3.Client{}
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(bc.AccessKeyID, bc.AccessKeySecret, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return client, err
	}

	client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(bc.BaseEndpoint)
	})
	return client, nil
}

func (bc *BucketConfig) UploadToBucket(file multipart.File, fileName string) error {
	client, err := bc.ConnectBucket()
	if err != nil {
		log.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		return err
	}
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &bc.BucketName,
		Key:    aws.String(fileName),
		Body:   bytes.NewReader(buf.Bytes()),
	})
	if err != nil {
		return err
	}
	return nil
}

/* func uploadToLocal(file multipart.File, fileName string) error {
	dst, err := os.Create(filepath.Join("./static/uploads/avatars", fileName))
	if err != nil {
		return err
	}
	defer dst.Close()

	bytes, err := io.Copy(dst, file)
	if err != nil {
		return err
	}

	fmt.Printf("Filename: %s\r\nNumber of bytes written: %s", filepath.Join("./static/uploads/avatars", fileName), strconv.FormatInt(bytes, 10))

	return nil
} */

func (bc *BucketConfig) ListBucket() ([]BucketFile, error) {
	var files []BucketFile
	client, err := bc.ConnectBucket()
	if err != nil {
		return files, err
	}
	fmt.Println(bc.BucketName)
	op, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket:  &bc.BucketName,
		MaxKeys: aws.Int32(100),
	})
	if err != nil {
		return files, err
	}

	var f BucketFile
	for _, v := range op.Contents {
		f.Key = *v.Key
		f.Size = *v.Size
		f.LastModified = *v.LastModified
		files = append(files, f)
	}

	return files, nil
}
