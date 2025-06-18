package service

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type AWSService interface {
	UploadFile(filePath string, fileName string, year string, intake string, teamID string) (string, string, error)
	GeneratePresignedURL(key string) (string, error)
	DeleteFile(key string) error
	DownloadBucketAsZip(w io.Writer) error
}

type awsService struct {
	bucketName string
	client     *s3.Client
}

// func NewAWSService(bucketName string) (AWSService, error) {
// 	cfg, err := config.LoadDefaultConfig(context.TODO())
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to load AWS config: %v", err)
// 	}

// 	fmt.Printf("%+v\n", cfg)
// 	client := s3.NewFromConfig(cfg)

// 	return &awsService{
// 		bucketName: bucketName,
// 		client:     client,
// 	}, nil
// }

func NewAWSService(bucketName string, region string, accessKeyId string, secretKey string) (AWSService, error) {
	// 1. Validate required AWS config
	if region == "" || accessKeyId == "" || secretKey == "" {
		return nil, fmt.Errorf("missing AWS configuration - check Region, AccessKeyID and SecretAccessKey")
	}

	// 2. Create AWS config with explicit credentials
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyId,
			secretKey,
			"",
		)),
		config.WithRetryer(func() aws.Retryer {
			return retry.NewStandard(func(o *retry.StandardOptions) {
				o.MaxAttempts = 3
			})
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	// 3. Verify credentials work
	creds, err := awsCfg.Credentials.Retrieve(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("invalid AWS credentials: %v", err)
	}

	log.Printf("AWS initialized with %s... in %s (expires: %v)",
		creds.AccessKeyID[:4],
		region,
		creds.Expires)

	return &awsService{
		bucketName: bucketName,
		client:     s3.NewFromConfig(awsCfg),
	}, nil
}

func (s *awsService) UploadFile(filePath string, fileName string, year string, intake string, teamID string) (string, string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to open file %s: %v", filePath, err)
	}
	defer file.Close()

	key := fmt.Sprintf("projects/%s/%s/%s/%s",
		year,
		intake,
		teamID,
		fileName)

	_, err = s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String("application/octet-stream"),
		ACL:         types.ObjectCannedACLPrivate,
	})

	if err != nil {
		return "", "", fmt.Errorf("failed to upload file to S3: %v", err)
	}

	return key, fileName, nil
}

func (s *awsService) GeneratePresignedURL(key string) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	req, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Hour * 24 * 7 // 1 week expiration
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}

	return req.URL, nil
}

func (s *awsService) DeleteFile(key string) error {
	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %v (bucket: %s, key: %s)",
			err, s.bucketName, key)
	}

	// Wait until file is actually deleted (optional)
	waiter := s3.NewObjectNotExistsWaiter(s.client)
	return waiter.Wait(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, 5*time.Minute) // Timeout after 5 minutes
}

func (s *awsService) DownloadBucketAsZip(w io.Writer) error {
	// Create a zip writer that streams to the response
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	// List all objects in the bucket
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucketName),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range page.Contents {
			// Create a zip entry for each file
			entry, err := zipWriter.Create(*obj.Key)
			if err != nil {
				return fmt.Errorf("failed to create zip entry: %w", err)
			}

			// Stream file directly from S3 to zip
			result, err := s.client.GetObject(context.TODO(), &s3.GetObjectInput{
				Bucket: aws.String(s.bucketName),
				Key:    obj.Key,
			})
			if err != nil {
				return fmt.Errorf("failed to get object %s: %w", *obj.Key, err)
			}
			defer result.Body.Close()

			if _, err := io.Copy(entry, result.Body); err != nil {
				return fmt.Errorf("failed to write to zip: %w", err)
			}
		}
	}

	return nil
}
