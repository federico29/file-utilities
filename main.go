package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

const (
	REGION      string = "us-east-1"
	BUCKET_NAME string = "fbm-files"
	FILES_PATH  string = "files/"
)

func main() {
	filePath := flag.String("file-path", "", "The absolute path of the file you want to upload")
	flag.Parse()
	if *filePath == "" {
		log.Fatal("The file path argument is required")
		os.Exit(0)
	}

	file, err := os.Open(*filePath)
	if err != nil {
		log.Fatal(err)
		os.Exit(0)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(REGION))
	if err != nil {
		log.Fatal(err)
		os.Exit(0)
	}

	s3Client := s3.NewFromConfig(cfg)

	fileKey, err := putObjectToBucket(s3Client, file, *filePath)
	if err != nil {
		log.Fatal(err)
		os.Exit(0)
	}

	url, err := getPresignedUrl(s3Client, fileKey)
	if err != nil {
		log.Fatal(err)
		os.Exit(0)
	}

	fmt.Printf("The file was uploaded, you can download with the following url: %v \n", url)
}

func putObjectToBucket(s3Client *s3.Client, myReader *os.File, filePath string) (string, error) {
	objectKey := FILES_PATH + uuid.New().String() + filepath.Ext(filePath)
	objectTags := fmt.Sprintf("RealName=%v&OriginalPath=%v", filepath.Base(filePath), filePath)
	_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:  aws.String(BUCKET_NAME),
		Key:     aws.String(objectKey),
		Body:    myReader,
		Tagging: aws.String(objectTags),
	})
	if err != nil {
		log.Fatal(err)
	}

	return objectKey, err
}

func getPresignedUrl(s3Client *s3.Client, objectKey string) (string, error) {
	presignClient := s3.NewPresignClient(s3Client)
	request, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(BUCKET_NAME),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(60 * int64(time.Second))
	})
	if err != nil {
		log.Println("Couldn't get the presigned request")
	}

	return request.URL, err
}
