// Package amazons3 uploads or downloads files from Amazon Simple Storage Service (S3)
package amazons3

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
        "encoding/base64"
	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	ivAction             = "action"
	ivEncodedImageData   = "encodedImageData"
	ivAwsAccessKeyID     = "awsAccessKeyID"
	ivAwsSecretAccessKey = "awsSecretAccessKey"
	ivAwsRegion          = "awsRegion"
	ivS3BucketName       = "s3BucketName"
	ivLocalLocation      = "localLocation"
	ivS3Location         = "s3Location"
	ivS3NewLocation      = "s3NewLocation"
	ovResult             = "result"
)

// log is the default package logger
var log = logger.GetLogger("activity-amazons3")

// MyActivity is a stub for your Activity implementation
type MyActivity struct {
	metadata *activity.Metadata
}

// NewActivity creates a new activity
func NewActivity(metadata *activity.Metadata) activity.Activity {
	return &MyActivity{metadata: metadata}
}

// Metadata implements activity.Activity.Metadata
func (a *MyActivity) Metadata() *activity.Metadata {
	return a.metadata
}

// Eval implements activity.Activity.Eval
func (a *MyActivity) Eval(context activity.Context) (done bool, err error) {

	// Get the action
	action := context.GetInput(ivAction).(string)
	encodedImageData := context.GetInput(ivEncodedImageData).(string)
	awsRegion := context.GetInput(ivAwsRegion).(string)
	s3BucketName := context.GetInput(ivS3BucketName).(string)
	// localLocation is a file when uploading a file or a directory when downloading a file
	localLocation := context.GetInput(ivLocalLocation).(string)
	s3Location := context.GetInput(ivS3Location).(string)
	s3NewLocation := context.GetInput(ivS3NewLocation).(string)

	// AWS Credentials, only if needed
	var awsAccessKeyID, awsSecretAccessKey = "", ""
	if context.GetInput(ivAwsAccessKeyID) != nil {
		awsAccessKeyID = context.GetInput(ivAwsAccessKeyID).(string)
	}
	if context.GetInput(ivAwsSecretAccessKey) != nil {
		awsSecretAccessKey = context.GetInput(ivAwsSecretAccessKey).(string)
	}

	// Create a session with Credentials only if they are set
	var awsSession *session.Session
	if awsAccessKeyID != "" && awsSecretAccessKey != "" {
		// Create new credentials using the accessKey and secretKey
		awsCredentials := credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, "")

		// Create a new session with AWS credentials
		awsSession = session.Must(session.NewSession(&aws.Config{
			Credentials: awsCredentials,
			Region:      aws.String(awsRegion),
		}))
	} else {
		// Create a new session without AWS credentials
		awsSession = session.Must(session.NewSession(&aws.Config{
			Region: aws.String(awsRegion),
		}))
	}

	// See which action needs to be taken
	var s3err error
	switch action {
	case "download":
		s3err = downloadFileFromS3(awsSession, localLocation, s3Location, s3BucketName)
	case "upload":
		s3err = uploadFileToS3(awsSession, localLocation, s3Location, s3BucketName, encodedImageData)
	case "delete":
		s3err = deleteFileFromS3(awsSession, s3Location, s3BucketName)
	case "copy":
		s3err = copyFileOnS3(awsSession, s3Location, s3BucketName, s3NewLocation)
	}
	if s3err != nil {
		// Set the output value in the context
		context.SetOutput(ovResult, s3err.Error())
		return true, s3err
	}

	// Set the output value in the context
	context.SetOutput(ovResult, "OK")

	return true, nil
}

// Function to download a file from an S3 bucket
func downloadFileFromS3(awsSession *session.Session, directory string, s3Location string, s3BucketName string) error {
	// Create an instance of the S3 Manager
	s3Downloader := s3manager.NewDownloader(awsSession)

	// Create a new temporary file
	f, err := os.Create(filepath.Join(directory, s3Location))
	if err != nil {
		return err
	}

	// Prepare the download
	objectInput := &s3.GetObjectInput{
		Bucket: aws.String(s3BucketName),
		Key:    aws.String(s3Location),
	}

	// Download the file to disk
	_, err = s3Downloader.Download(f, objectInput)
	if err != nil {
		return err
	}

	return nil
}

// Function to delete a file from an S3 bucket
func deleteFileFromS3(awsSession *session.Session, s3Location string, s3BucketName string) error {
	// Create an instance of the S3 Manager
	s3Session := s3.New(awsSession)

	objectDelete := &s3.DeleteObjectInput{
		Bucket: aws.String(s3BucketName),
		Key:    aws.String(s3Location),
	}

	// Delete the file from S3
	_, err := s3Session.DeleteObject(objectDelete)
	if err != nil {
		return err
	}

	return nil
}

// Function to upload a file from an S3 bucket
func uploadFileToS3(awsSession *session.Session, localFile string, s3Location string, s3BucketName string, encodedImageData string) error {
	// Create an instance of the S3 Manager
	s3Uploader := s3manager.NewUploader(awsSession)

	// Create a file pointer to the source
	//reader, err := os.Open(localFile)
	reader, err := base64.StdEncoding.DecodeString(encodedImageData)
	if err != nil {
		return err
	}
	//defer reader.Close()

	// Prepare the upload
	uploadInput := &s3manager.UploadInput{
		Bucket: aws.String(s3BucketName),
		Key:    aws.String(s3Location),
		Body:   bytes.NewReader(reader),
	}

	// Upload the file
	_, err = s3Uploader.Upload(uploadInput)
	if err != nil {
		return err
	}

	return nil
}

// Function to copy a file in an S3 bucket
func copyFileOnS3(awsSession *session.Session, s3Location string, s3BucketName string, s3NewLocation string) error {
	// Create an instance of the S3 Session
	s3Session := s3.New(awsSession)

	// Prepare the copy object
	objectInput := &s3.CopyObjectInput{
		Bucket:     aws.String(s3BucketName),
		CopySource: aws.String(fmt.Sprintf("/%s/%s", s3BucketName, s3Location)),
		Key:        aws.String(s3NewLocation),
	}

	// Copy the object
	_, err := s3Session.CopyObject(objectInput)
	if err != nil {
		return err
	}

	return nil
}
