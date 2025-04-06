package artifacts_export

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
)

func (ah *ArtifactHandler) CopyArtifactsToS3(ctx context.Context, sourceKey, sourceBucket, destinationKey string) error {

	logrus.Infof("copying artifacts for file: %s", sourceKey)
	copySourceKey := fmt.Sprintf("%s/%s", sourceBucket, sourceKey)

	destinationBucket := ah.CustomerDetail.DestinationBucket

	_, err := ah.S3Instance.CopyObjectWithContext(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(destinationBucket),
		CopySource: aws.String(copySourceKey),
		Key:        aws.String(destinationKey),
	})
	if err != nil {
		if ah.Retry > 0 {
			ah.Retry--
			logrus.Errorf("retry %d to copy artifacts to s3: %v", ah.Retry, err)
			return ah.CopyArtifactsToS3(ctx, sourceKey, sourceBucket, destinationKey)
		}
		return fmt.Errorf("%s, getting error: %w", copySourceKey, err)
	}
	logrus.Infof("Artifact successfully copied from location %s to bucket %s/%s\n", copySourceKey, destinationBucket, destinationKey)
	return nil
}

// LocalCopyArtifact : locally copy artifacts to validate working. This piece can be used specifically by developers to check out feature on local system
func (ah *ArtifactHandler) LocalCopyArtifact(ctx context.Context, sourceBucket, sourceKey, destinationKey string) error {
	localDest := os.Getenv("LOCAL_DEST")
	if localDest == "" {
		return fmt.Errorf("local destination(LOCAL_DEST) must be provided when using ENV==local")
	}
	localDestKey := fmt.Sprintf("%s/%s", localDest, destinationKey)
	getObject, err := ah.S3Instance.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(sourceBucket),
		Key:    aws.String(sourceKey),
	})
	if err != nil {
		return fmt.Errorf("unable to get object: %w", err)
	}
	body, err := io.ReadAll(getObject.Body)
	if err != nil {
		return fmt.Errorf("couldn't read object bytes: %w", err)
	}

	localDir := strings.Split(localDestKey, "/")[:len(strings.Split(localDestKey, "/"))-1]
	localDirKey := strings.Join(localDir, "/")
	err = os.MkdirAll(localDirKey, 0755)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	fmt.Println("Directory Created Successfully")

	err = os.WriteFile(localDestKey, body, 0644)
	if err != nil {
		return fmt.Errorf("unable to write file %s, getting error: %w", localDestKey, err)
	}
	return nil
}
