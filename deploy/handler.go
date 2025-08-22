package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func Handler(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "./WiiNewsPR", "-o", "/tmp", "-c", "/tmp/cache")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run WiiNewsPR: %w\nOutput: %s", err, string(output))
	}

	hour := time.Now().Format("15")
	filePath := fmt.Sprintf("/tmp/v2/1/049/news.bin.%s", hour)

	bucketName := os.Getenv("S3_BUCKET")
	if bucketName == "" {
		return fmt.Errorf("S3_BUCKET environment variable is required")
	}

	keyPrefix := os.Getenv("S3_PREFIX")
	if keyPrefix == "" {
		return fmt.Errorf("S3_PREFIX environment variable is required")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	uploader := manager.NewUploader(s3.NewFromConfig(cfg))

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", filePath, err)
	}
	defer file.Close()

	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(fmt.Sprintf("%snews.bin.%s", keyPrefix, hour)),
		Body:        file,
		ContentType: aws.String("application/octet-stream"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload: %w", err)
	}

	fmt.Printf("Uploaded news.bin.%s to S3 bucket %s\n", hour, bucketName)
	return nil
}

func main() {
	lambda.Start(Handler)
}
