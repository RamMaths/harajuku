package awsS3

import (
	"bytes"
	"context"
	"os"
	"testing"

	"harajuku/backend/internal/adapter/storage/awsS3"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func TestSaveFile(t *testing.T) {
	// Initialize AWS session
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))
	
	// Create S3 adapter
	s3Adapter := awsS3.NewAwsS3(sess, "aws-harajuku-bucket-001")

	// 1. Read test file
	testFilePath := "/Users/ram/Downloads/minutas.md" // Create this file in your project
	fileData, err := os.ReadFile(testFilePath)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// 2. Upload file
	objectKey := "minutas.md"
	uploadedPath, err := s3Adapter.Save(context.Background(), fileData, objectKey)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 3. Verify upload was successful
	if uploadedPath != objectKey {
		t.Errorf("Expected uploaded path %q, got %q", objectKey, uploadedPath)
	}

	// 4. (Optional) Verify file contents
	// ... Add verification logic if needed
}

func TestSaveAndVerifyFile(t *testing.T) {
	ctx := context.Background()
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))
	
	bucket := "ews-bucket-test-001"
	s3Adapter := awsS3.NewAwsS3(sess, bucket)
	s3Client := s3.New(sess)

	// Test data
	testContent := []byte("This is a test file content")
	objectKey := "test-upload.txt"

	// Cleanup before test (in case previous test failed)
	_, _ = s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
	})

	t.Run("Save file successfully", func(t *testing.T) {
		// Upload file
		uploadedPath, err := s3Adapter.Save(ctx, testContent, objectKey)
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		if uploadedPath != objectKey {
			t.Errorf("Expected uploaded path %q, got %q", objectKey, uploadedPath)
		}

		// Verify file exists in S3
		_, err = s3Client.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(objectKey),
		})
		if err != nil {
			t.Errorf("Failed to verify uploaded file: %v", err)
		}
	})

	t.Run("Verify file contents", func(t *testing.T) {
		// Download and verify contents
		downloadedData, err := s3Adapter.Get(ctx, objectKey)
		if err != nil {
			t.Fatalf("Failed to get file: %v", err)
		}

		if !bytes.Equal(testContent, downloadedData) {
			t.Errorf("File content mismatch")
		}
	})

	// Cleanup after test
	t.Cleanup(func() {
		_, err := s3Client.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(objectKey),
		})
		if err != nil {
			t.Logf("Warning: failed to clean up test file: %v", err)
		}
	})
}
