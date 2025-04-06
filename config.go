package artifacts_export

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
)

type CustomerConfig struct {
	Customers []Customer `json:"customers"`
	Retry     int        `json:"retry"`
}
type Customer struct {
	Org               string `json:"org"`
	DestinationBucket string `json:"destinationBucket"`
}

func (ah *ArtifactHandler) ReadCustomerConfiguration(ctx context.Context, org string) error {
	var config CustomerConfig

	var err error
	localConfigPath := os.Getenv("LOCAL_CONFIGURATION")
	if localConfigPath != "" {
		config, err = localConfigRead(localConfigPath)
		if err != nil{
		    return fmt.Errorf("couldn't read local config file: %w", err)
		}
	} else {
		config, err = ah.S3ConfigRead(ctx, org)
		if err != nil{
		    return fmt.Errorf("couldn't read config file from S3: %w", err)
		}
	}

	for _, customer := range config.Customers {
		if customer.Org == org {
			ah.Retry = config.Retry
			ah.CustomerDetail = customer
			logrus.WithFields(logrus.Fields{
				"org":               customer.Org,
				"destinationBucket": customer.DestinationBucket,
				"retry":             config.Retry,
			}).Info("Configuration read")
			break
		}
	}
	return nil
}

func localConfigRead(localConfigPath string) (CustomerConfig, error){
    var config CustomerConfig
    fileBytes, err := os.ReadFile(localConfigPath)
	if err != nil {
	    return fmt.Errorf("couldn't read local config file: %w", err)
	}
    err = json.Unmarshal(fileBytes, &config)
	if err != nil {
	    return fmt.Errorf("couldn't decode local config file: %w", err)
	}

    return config, nil
}

func (ah *ArtifactHandler) S3ConfigRead(ctx context.Context, org string) (CustomerConfig, error){
    var config CustomerConfig

    configurationBucket := os.Getenv("CONFIGURATION_BUCKET")
    configKey := os.Getenv("CONFIGURATION_KEY")

    logrus.Infof("Reading configuration for org:%s", org)

    resp, err := ah.S3Instance.GetObjectWithContext(ctx, &s3.GetObjectInput{
        Bucket: aws.String(configurationBucket),
    	Key:    aws.String(configKey),
   	})
    if err != nil {
        return fmt.Errorf("couldn't read config file bucket:%s, key:%s. Getting error: %w", configurationBucket, configKey, err)
    }

    defer func() {
        err := resp.Body.Close()
    	if err != nil {
    		logrus.Errorf("Error closing response body: %v", err)
    	}
    }()

   	err = json.NewDecoder(resp.Body).Decode(&config)
    if err != nil {
    	return fmt.Errorf("couldn't decode config file : %w", err)
    }
}