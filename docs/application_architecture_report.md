# Artifacts Export Service - Architecture Report

## Application Overview

The Artifacts Export Service is a Lambda-based application designed to automatically copy files from a source S3 bucket to multiple destination S3 buckets based on file types and organizational configurations. The service operates synchronously, triggering whenever a new file is placed in the source bucket.

## Architecture Components

### 1. AWS Lambda Function
- Core component that handles the file copying logic
- Triggered by S3 events (when files are uploaded to the source bucket)
- Filters files based on configured file extensions (jpg, pdf, indd, xml, etc.)
- Processes files synchronously to ensure immediate availability in destination buckets

### 2. Configuration Management
- Uses a JSON configuration file stored in a separate S3 bucket
- Configuration defines customer/organization mappings to destination buckets
- Includes retry settings for error handling
- Allows for easy addition or modification of customer destinations

#### Sample Configuration JSON
```json
{
  "customers": [
    {
      "org": "org1",
      "destinationBucket": "org1-export"
    },
    {
      "org": "org2",
      "destinationBucket": "org2-export"
    },
    {
      "org": "org3",
      "destinationBucket": "org3-export"
    },
    {
      "org": "org4",
      "destinationBucket": "org4-export"
    }
  ],
  "retry": 1
}
```

This configuration maps each organization to its respective destination bucket and sets the retry count to 1.

### 3. Data Flow
- Source S3 bucket → Lambda function → Multiple destination S3 buckets
- Files are filtered by suffix (e.g., jpg, pdf, indd, xml)
- Each organization's files are routed to their specific destination bucket
- Maintains the original file structure and naming

### 4. Environment Configuration
- **Production Environment**:
  - Uses environment variables for configuration
  - Required parameters:
    - SUFFIX_TO_COPY: Comma-separated list of file extensions to process
    - CONFIGURATION_BUCKET: S3 bucket containing the configuration file
    - CONFIGURATION_KEY: Path to the configuration file within the bucket

- **Local Development Environment**:
  - Uses additional local environment variables
  - Required parameters:
    - AWS credentials (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, etc.)
    - ENV: Set to "local"
    - LOCAL_CONFIGURATION: Path to local configuration file
    - LOCAL_DEST: Directory for copied files
    - SUFFIX_TO_COPY: File extensions to process
    - DESTINATION_KEY: Target filename

### 5. Error Handling
- Includes configurable retry mechanism
- Retry count defined in the configuration file

## Multi-tenant Design

The application is designed for multi-tenant use, where different organizations' files need to be copied to their respective destination buckets. This architecture maintains separation of data while using a common ingestion point, providing:

- Simplified file management
- Centralized configuration
- Efficient resource utilization
- Scalability for adding new organizations

## Deployment Considerations

When deploying this Lambda service, ensure:
- Proper IAM permissions for S3 bucket access
- Appropriate S3 event triggers are configured
- Sufficient Lambda timeout settings for large file processing
- Monitoring and alerting for any failures
