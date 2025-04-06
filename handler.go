package artifacts_export

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
)

type ArtifactHandler struct {
	S3Instance     *s3.S3
	Retry          int
	CustomerDetail Customer
}
type CloudWatchEvent struct {
	Bucket Name   `json:"bucket"`
	Object Object `json:"object"`
}
type Name struct {
	Name string `json:"name"`
}
type Object struct {
	Key string `json:"key"`
}

func (ah *ArtifactHandler) Handler(ctx context.Context, event events.CloudWatchEvent) error {
	logrus.Infof("Received EventDetail: %s", string(event.Detail))
	var data CloudWatchEvent
	err := json.Unmarshal(event.Detail, &data)
	if err != nil {
		return fmt.Errorf("unable to decode event: %w", err)
	}

	copyFile := false
	suffixToCopy := os.Getenv("SUFFIX_TO_COPY")
	filePath := data.Object.Key
	org := strings.Split(filePath, "/")[0]
	for _, suffix := range strings.Split(suffixToCopy, ",") {
		if strings.HasSuffix(filePath, strings.ToLower(suffix)) {
			copyFile = true
			break
		}
	}

	err = ah.ReadCustomerConfiguration(ctx, org)
	if err != nil {
		return fmt.Errorf("unable to read customer configuration: %w", err)
	}

	err = ah.CopyArtifacts(ctx, &data)
	if err != nil {
		logrus.Errorf("unable to copy artifacts to s3: %v", err)
		return fmt.Errorf("unable to copy artifacts to s3: %w", err)
	}

	return nil
}

func (ah *ArtifactHandler) CopyArtifacts(ctx context.Context, event *CloudWatchEvent) error {
	destinationKey := os.Getenv("DESTINATION_KEY")
	sourceBucket := event.Bucket.Name
	sourceKey := event.Object.Key
	if destinationKey == "" {
		key := strings.Split(sourceKey, "/")[1:]
		destinationKey = strings.Join(key, "/")
	}
	if os.Getenv("ENV") == "local" {
		logrus.Infof("using local environment for copying artifact")
		err := ah.LocalCopyArtifact(ctx, sourceBucket, sourceKey, destinationKey)
		if err != nil {
			return fmt.Errorf("unable to copy artifacts to local: %w", err)
		}
	} else {
		err := ah.CopyArtifactsToS3(ctx, sourceKey, sourceBucket, destinationKey)
		if err != nil {
			return fmt.Errorf("unable to copy artifacts at destination:%s from source: %w", destinationKey, err)
		}
	}
	return nil
}
