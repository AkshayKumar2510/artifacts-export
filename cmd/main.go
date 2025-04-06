package main

import (
	"fmt"
	"os"

	artifactHandler "artifacts-export"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	region := os.Getenv("AWS_REGION")

	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(region)},
		SharedConfigState: session.SharedConfigDisable,
	})
	if err != nil {
		fmt.Println("Error creating session:", err)
		return
	}

	s3Instance := s3.New(sess)

	exportHandler := artifactHandler.ArtifactHandler{
		S3Instance: s3Instance,
	}

	lambda.Start(exportHandler.Handler)
}
