package upload

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gorant/database"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type BucketFile struct {
	Key          string
	Size         int64
	LastModified time.Time
}

func (f *BucketFile) SizeString() string {
	s := f.Size / 1000
	return strconv.FormatInt(s, 10)
}

func (f *BucketFile) LastModifiedString() string {
	return f.LastModified.Format("2006-01-02 15:04:05")
}

type BucketConfig struct {
	Store              string
	BucketName         string
	AccountID          string
	BaseEndpoint       string
	PublicAccessDomain string
	AccessKeyID        string
	AccessKeySecret    string
}

type LookupFile struct {
	File        multipart.File
	FileID      uuid.UUID
	FileKey     string
	FileStore   string
	FileBucket  string
	FileBaseURL string
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

func WithStore(storeName string) func(*BucketConfig) {
	return func(c *BucketConfig) {
		c.Store = storeName
	}
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

func UploadToLocal(file multipart.File, fileName string) error {
	localDir := "./static/uploads"
	dst, err := os.Create(filepath.Join(localDir, fileName))
	if err != nil {
		return err
	}
	defer dst.Close()

	bytes, err := io.Copy(dst, file)
	if err != nil {
		return err
	}

	fmt.Printf("Filename: %s\r\nNumber of bytes written: %s", filepath.Join(localDir, fileName), strconv.FormatInt(bytes, 10))

	return nil
}

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

func GetOrphanFilesDB() ([]BucketFile, error) {
	var files []BucketFile
	var f BucketFile
	rows, err := database.DB.Query(`SELECT files.file_key FROM files WHERE files.file_id NOT IN
										(SELECT DISTINCT comments.file_id
										FROM comments
										WHERE comments.file_id IS NOT NULL);`)
	if err != nil {
		return files, err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&f.Key); err != nil {
			return files, err
		}
		files = append(files, f)
	}
	return files, nil
}

func (bc *BucketConfig) DeleteBucketFiles(files []BucketFile) error {
	client, err := bc.ConnectBucket()
	if err != nil {
		return err
	}
	var filesList []types.ObjectIdentifier
	var f types.ObjectIdentifier
	for _, v := range files {
		f.Key = aws.String(v.Key)
		filesList = append(filesList, f)
	}
	fmt.Println(filesList)
	fmt.Println(client)
	_, err = client.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
		Bucket: &bc.BucketName,
		Delete: &types.Delete{Objects: filesList},
	})
	if err != nil {
		return err
	}

	return nil
}

func DeleteOrphanFilesDB(files []BucketFile) error {
	var keys []string
	for _, v := range files {
		key := v.Key
		keys = append(keys, key)
	}
	// Bindvars only work with (?), not ($1)
	query, args, err := sqlx.In("DELETE FROM files WHERE file_key IN (?);", keys)
	if err != nil {
		return err
	}
	query = database.DB.Rebind(query)
	fmt.Println(query)
	_, err = database.DB.Exec(query, args...)
	if err != nil {
		return err
	}
	return nil
}
