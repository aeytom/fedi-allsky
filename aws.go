package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func uploadS3(filename string) (string, error) {
	// The session the S3 Uploader will use
	sess := session.Must(session.NewSession())

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)

	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(ArgS3Bucket),
		Key:    aws.String(filepath.Base(filename)),
		Body:   f,
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("file uploaded to, %s\n", aws.StringValue(&result.Location)), nil

}
