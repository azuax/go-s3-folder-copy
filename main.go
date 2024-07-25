package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func main() {
	// Define command-line flags
	localFolderPath := flag.String("path", "", "Local folder path to upload")
	bucketName := flag.String("bucket", "", "S3 bucket name")
	numWorkers := flag.Int("t", 10, "Number of concurrent workers")
	baseFolderS3 := flag.String("base-folder-s3", "", "Base path in S3 bucket")

	// Parse the flags
	flag.Parse()

	// Check if required flags are provided
	if *localFolderPath == "" || *bucketName == "" {
		log.Fatalf("Usage: %s --path <local folder path> --bucket <S3 bucket name> [--base-folder-s3 <Base folder to store in>] [--t <number of workers>]", os.Args[0])
	}

	envRegion := os.Getenv("AWS_REGION")
	if envRegion == "" {
		envRegion = "us-east-1"
	}

	// use folder name as base folder if not provided
	if *baseFolderS3 == "" {
		*baseFolderS3 = filepath.Base(*localFolderPath)
	}

	// Load the AWS SDK configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(envRegion))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create an S3 service client
	s3Client := s3.NewFromConfig(cfg)

	// Channel to send file paths to workers
	fileCh := make(chan string)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < *numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range fileCh {
				fmt.Printf("Uploading file: %s\n", filePath)
				err := uploadFileToS3(s3Client, *bucketName, *localFolderPath, filePath, *baseFolderS3)
				if err != nil {
					log.Printf("Failed to upload file %s: %v", filePath, err)
				}
			}
		}()
	}

	// Walk through the local folder and send file paths to the channel
	err = filepath.Walk(*localFolderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		//Exclude folders:
		excludedFolders := []string{".DS_Store", ".env", ".venv", "virtualenv", "__pycache__"}

		// Exclude files:
		excludedFiles := []string{".DS_Store"}

		if info.IsDir() && slices.Contains(excludedFolders, info.Name()) {
			return filepath.SkipDir
		} else if slices.Contains(excludedFiles, info.Name()) {
			return nil
		}

		if !info.IsDir() {
			fileCh <- path
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Failed to walk through the local folder: %v", err)
	}

	// Close the channel and wait for all workers to finish
	close(fileCh)
	wg.Wait()

	fmt.Println("Upload completed successfully.")
}

func uploadFileToS3(client *s3.Client, bucketName, baseFolder, filePath string, baseFolderS3 string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	relativePath, err := filepath.Rel(baseFolder, filePath)
	if err != nil {
		return err
	}

	relativePath = filepath.Join(baseFolderS3, relativePath)

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(relativePath),
		Body:   file,
		ACL:    types.ObjectCannedACLPrivate,
	})

	return err
}
