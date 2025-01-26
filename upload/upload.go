package upload

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"image"
	_ "image/gif"  // For image.Decode() of image uploads.
	_ "image/jpeg" // Ditto.
	_ "image/png"  // Ditto.
	"io"
	"log"
	"math"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorant/database"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/chai2010/webp"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/image/draw"
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
	File         multipart.File
	ID           uuid.UUID
	UserID       string
	Key          string
	ThumbnailKey string
	Store        string
	Bucket       string
	BaseURL      string
	UploadedAt   time.Time
}

type NullFile struct {
	File         multipart.File
	ID           uuid.NullUUID
	UserID       sql.NullString
	Key          sql.NullString
	ThumbnailKey sql.NullString
	Store        sql.NullString
	Bucket       sql.NullString
	BaseURL      sql.NullString
	UploadedAt   time.Time
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
		return client, fmt.Errorf("error connecting to bucket: %w", err)
	}

	client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(bc.BaseEndpoint)
	})
	return client, nil
}

func (bc *BucketConfig) UploadToBucket(file multipart.File, fileName string, fileType string) (string, string, uuid.UUID, error) {
	client, err := bc.ConnectBucket()
	if err != nil {
		log.Fatal(err)
	}
	uniqueKey := uuid.New()
	var buf bytes.Buffer
	var bufThumb bytes.Buffer
	var thumbnailFileName string
	maxWidthThumbnail := 200
	switch fileType {
	case "image/jpeg", "image/png":
		buf, err = ImagetoWebp(file)
		if err != nil {
			fmt.Println("Failed at image to webp step")
			log.Fatal(err)
		}
		// reset this since NewReader doesn't reset position to beginning of file.
		_, err := file.Seek(0, 0)
		if err != nil {
			log.Fatal(err)
		}
		bufThumb, err = GenerateThumbnail(file, maxWidthThumbnail)
		if err != nil {
			fmt.Println("Failed at image to thumbnail step")
			log.Fatal(err)
		}
		if strings.Contains(fileName, ".") {
			// Has an extension in filename.
			nm := strings.Split(fileName, ".")
			fileName = nm[0] + ".webp"
			thumbnailFileName = nm[0] + "-tn.webp"
		} else {
			// No extension in filename, just add a .webp suffix.
			fileName = fileName + ".webp"
			thumbnailFileName = fileName + "-tn.webp"
		}
		fileKey := uniqueKey.String() + "-" + fileName
		_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: &bc.BucketName,
			Key:    aws.String(fileKey),
			Body:   bytes.NewReader(buf.Bytes()),
		})
		if err != nil {
			return fileName, thumbnailFileName, uniqueKey, fmt.Errorf("error with uploading webp (from jpg/png) in bucket: %w", err)
		}
		fileKey = uniqueKey.String() + "-" + thumbnailFileName
		_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: &bc.BucketName,
			Key:    aws.String(fileKey),
			Body:   bytes.NewReader(bufThumb.Bytes()),
		})
		if err != nil {
			return fileName, thumbnailFileName, uniqueKey, fmt.Errorf("error with uploading webp thumbnail (from jpg/png) in bucket: %w", err)
		}
	default:
		fileKey := uniqueKey.String() + "-" + fileName
		_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: &bc.BucketName,
			Key:    aws.String(fileKey),
			Body:   file,
		})
		if err != nil {
			return fileName, thumbnailFileName, uniqueKey, fmt.Errorf("error with uploading non-jpg/png in bucket: %w", err)
		}

		if strings.Contains(fileName, ".") {
			// Has an extension in filename.
			nm := strings.Split(fileName, ".")
			thumbnailFileName = nm[0] + "-tn.webp"
		} else {
			// No extension in filename, just add a .webp suffix.
			thumbnailFileName = fileName + "-tn.webp"
		}
		fileKey = uniqueKey.String() + "-" + thumbnailFileName
		_, err := file.Seek(0, 0)
		if err != nil {
			log.Fatal(err)
		}
		bufThumb, err = GenerateThumbnail(file, maxWidthThumbnail)
		if err != nil {
			fmt.Println("Failed at image to thumbnail step")
			log.Fatal(err)
		}
		_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: &bc.BucketName,
			Key:    aws.String(fileKey),
			Body:   bytes.NewReader(bufThumb.Bytes()),
		})
		if err != nil {
			return fileName, thumbnailFileName, uniqueKey, fmt.Errorf("error with uploading webp thumbnail (non-jpg/png) in bucket: %w", err)
		}
	}
	return fileName, thumbnailFileName, uniqueKey, nil
}

func ImagetoWebp(file multipart.File) (bytes.Buffer, error) {
	var buf bytes.Buffer
	var img image.Image
	var err error
	var format string
	img, format, err = image.Decode(file)
	if err != nil {
		return buf, fmt.Errorf("error: failed at image to webp: %w", err)
	}
	fmt.Println("Image format: ", format)

	if err = webp.Encode(&buf, img, &webp.Options{Lossless: false, Quality: 70}); err != nil {
		log.Println(err)
	}
	p := &buf
	return *p, nil
}

func GenerateThumbnail(file multipart.File, width int) (bytes.Buffer, error) {
	var err error
	var buf bytes.Buffer
	src, format, err := image.Decode(file)
	if err != nil {
		return buf, fmt.Errorf("error decoding file for thumbnail: %w", err)
	}
	fmt.Println(format)
	ratio := (float64)(src.Bounds().Max.Y) / (float64)(src.Bounds().Max.X)
	height := int(math.Round(float64(width) * ratio))
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	if err = webp.Encode(&buf, dst, &webp.Options{Lossless: false, Quality: 60}); err != nil {
		log.Println(err)
	}
	p := &buf
	return *p, nil
}

func ToLocalWebp(file multipart.File) (string, error) {
	var buf bytes.Buffer
	localDir := "./static/uploads/"
	newFileName := uuid.NewString() + ".webp"
	img, format, err := image.Decode(file)
	if err != nil {
		return newFileName, fmt.Errorf("error decoding file: %w", err)
	}
	fmt.Println(format)
	// output, err := os.Create(localDir + uuid.NewString())
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// defer output.Close()
	if err = webp.Encode(&buf, img, &webp.Options{Lossless: false, Quality: 75}); err != nil {
		log.Println(err)
	}
	if err = os.WriteFile(localDir+newFileName, buf.Bytes(), 0o600); err != nil {
		log.Println(err)
	}
	return newFileName, nil
}

func ToLocal(file multipart.File, fileName string) error {
	localDir := "./static/uploads"
	dst, err := os.Create(filepath.Join(localDir, fileName))
	if err != nil {
		return fmt.Errorf("error creating destination file for upload: %w", err)
	}
	defer dst.Close()
	bytes, err := io.Copy(dst, file)
	if err != nil {
		return fmt.Errorf("error copying multipart.File to destination file: %w", err)
	}
	fmt.Printf("Filename: %s\r\nNumber of bytes written: %s", filepath.Join(localDir, fileName), strconv.FormatInt(bytes, 10))
	return nil
}

func (bc *BucketConfig) ListBucket() ([]BucketFile, error) {
	var files []BucketFile
	client, err := bc.ConnectBucket()
	if err != nil {
		return files, fmt.Errorf("error connecting to bucket: %w", err)
	}
	op, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket:  &bc.BucketName,
		MaxKeys: aws.Int32(100),
	})
	if err != nil {
		return files, fmt.Errorf("error listing objects in bucket: %w", err)
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
		return files, fmt.Errorf("error querying file_key from files table: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&f.Key); err != nil {
			return files, fmt.Errorf("error scanning file_key from db: %w", err)
		}
		files = append(files, f)
	}
	return files, nil
}

func (bc *BucketConfig) DeleteBucketFiles(files []BucketFile) error {
	client, err := bc.ConnectBucket()
	if err != nil {
		return fmt.Errorf("error connecting to bucket: %w", err)
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
		return fmt.Errorf("error deleting from bucket: %w", err)
	}
	return nil
}

func DeleteDBFileRecord(key string) error {
	_, err := database.DB.Exec(`DELETE FROM files WHERE file_key=$1`, key)
	if err != nil {
		return fmt.Errorf("error deleting db file record: %w", err)
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
		return fmt.Errorf("error using sqlx.In for deleting orphan files: %w", err)
	}
	query = database.DB.Rebind(query)
	fmt.Println(query)
	_, err = database.DB.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error deleting from files table: %w", err)
	}
	return nil
}
