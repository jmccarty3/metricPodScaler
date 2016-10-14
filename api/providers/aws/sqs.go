package aws

import (
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/golang/glog"
	"github.com/jmccarty3/metricPodScaler/api/providers"
)

const defaultRegion = "us-east-1"
const envRegionName = "AWS_DEFAULT_REGION"

//SQS represents an aws SQS queue
type SQS struct {
	QueueURL string `yaml:"queueURL"`

	client *sqs.SQS
}

func getAWSCredentials() *credentials.Credentials {
	return credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.EnvProvider{},
			&ec2rolecreds.EC2RoleProvider{
				Client: ec2metadata.New(session.New(&aws.Config{})),
			},
			&credentials.SharedCredentialsProvider{},
		})
}

func getMetadataClient() *ec2metadata.EC2Metadata {
	return ec2metadata.New(session.New(&aws.Config{}))
}

func getRegion() string {
	//Check the environment first
	if region, found := os.LookupEnv(envRegionName); found {
		if region != "" {
			glog.Infof("Using region %s from environment", region)
			return region
		}
	}
	//Attempt to check the metadata service
	client := getMetadataClient()
	if client.Available() {
		region, err := client.Region()
		if err == nil {
			glog.Info("Metadata service returned %s region", region)
			return region
		}
	}

	glog.Warningf("Unable to find region from Metadata service or the environment. Using default %s", defaultRegion)
	//Give up. Use default
	return defaultRegion
}

func createSqsClient(url string) (*sqs.SQS, error) {
	sess, err := session.NewSession(&aws.Config{
		Credentials: getAWSCredentials(),
		Region:      aws.String(getRegion()),
	})
	if err != nil {
		return nil, err
	}
	return sqs.New(sess), nil
}

//Connect connects to the sqs in question
func (s *SQS) Connect() (err error) {
	glog.V(2).Infof("Trying to connect to SQS: %s", s.QueueURL)
	s.client, err = createSqsClient(s.QueueURL)

	glog.V(2).Infof("Connected to SQS: %s", s.QueueURL)
	return
}

//CurrentCount attempts to get the appox visible/non-visible messages from the queue
func (s *SQS) CurrentCount() (int64, error) {
	params := &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(s.QueueURL),
		AttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameApproximateNumberOfMessages),
			aws.String(sqs.QueueAttributeNameApproximateNumberOfMessagesNotVisible), //Need to accoutn for in flgiht/ processing messages
		},
	}

	resp, err := s.client.GetQueueAttributes(params)
	if err != nil {
		return 0, err
	}
	messageCount, err := strconv.ParseInt(*resp.Attributes[sqs.QueueAttributeNameApproximateNumberOfMessages], 10, 64)
	if err != nil {
		return 0, err
	}
	inflightCount, err := strconv.ParseInt(*resp.Attributes[sqs.QueueAttributeNameApproximateNumberOfMessagesNotVisible], 10, 64)
	if err != nil {
		return 0, err
	}
	return messageCount + inflightCount, nil
}

func init() {
	providers.AddProvider("sqs", func() providers.Provider {
		return &SQS{}
	})
}
