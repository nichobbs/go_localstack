package examples

import (
    "log"
    "fmt"
    "testing"
    "os"
    "github.com/mitchelldavis/go_localstack/pkg/localstack"
    
    "github.com/aws/aws-sdk-go/service/apigateway"
    "github.com/aws/aws-sdk-go/service/kinesis"
    "github.com/aws/aws-sdk-go/service/dynamodb"
    "github.com/aws/aws-sdk-go/service/dynamodbstreams"
	"github.com/aws/aws-sdk-go/service/s3"
    "github.com/aws/aws-sdk-go/service/firehose"
    "github.com/aws/aws-sdk-go/service/lambda"
    "github.com/aws/aws-sdk-go/service/sns"
    "github.com/aws/aws-sdk-go/service/sqs"
    "github.com/aws/aws-sdk-go/service/redshift"
)

var LOCALSTACK *localstack.Localstack

func TestMain(t *testing.M) {
    apigateway, _ := localstack.NewLocalstackService("apigateway")
    kinesis, _ := localstack.NewLocalstackService("kinesis")
	dynamodb, _ := localstack.NewLocalstackService("dynamodb")
	dynamodbstreams, _ := localstack.NewLocalstackService("dynamodbstreams")
	//es, _ := localstack.NewLocalstackService("es")
	s3, _ := localstack.NewLocalstackService("s3")
	firehose, _ := localstack.NewLocalstackService("firehose")
	lambda, _ := localstack.NewLocalstackService("lambda")
	sns, _ := localstack.NewLocalstackService("sns")
	sqs, _ := localstack.NewLocalstackService("sqs")
	redshift, _ := localstack.NewLocalstackService("redshift")
	email, _ := localstack.NewLocalstackService("ses")
	route53, _ := localstack.NewLocalstackService("route53")
	cloudformation, _ := localstack.NewLocalstackService("cloudformation")
	cloudwatch, _ := localstack.NewLocalstackService("cloudwatch")
	ssm, _ := localstack.NewLocalstackService("ssm")
	secretsmanager, _ := localstack.NewLocalstackService("secretsmanager")
	stepfunctions, _ := localstack.NewLocalstackService("stepfunctions")
	logs, _ := localstack.NewLocalstackService("logs")
	sts, _ := localstack.NewLocalstackService("sts")
	iam, _ := localstack.NewLocalstackService("iam")

    // Gather them all up...
    LOCALSTACK_SERVICES := &localstack.LocalstackServiceCollection {
        *apigateway,
        *kinesis,
        *dynamodb,
        *dynamodbstreams,
        //*es,
        *s3,
        *firehose,
        *lambda,
        *sns,
        *sqs,
        *redshift,
        *email,
        *route53,
        *cloudformation,
        *cloudwatch,
        *ssm,
        *secretsmanager,
        *stepfunctions,
        *logs,
        *sts,
        *iam,
    }

    // Initialize the services
    var err error
    LOCALSTACK, err = localstack.NewLocalstack(LOCALSTACK_SERVICES)
    if err != nil {
        log.Fatal(fmt.Sprintf("Unable to create the localstack instance: %s", err))
    }
    if LOCALSTACK == nil {
        log.Fatal("LOCALSTACK was nil.")
    }

    // If you need to initialize s3 or sqs, do it here.

    // RUN TESTS HERE
    result := t.Run()

    // We can't use defer this because os.Exit terminates the application in place
    // and the defered function won't be called.  We need to call os.Exit because
    // we need to correctly report the test results.
    LOCALSTACK.Destroy()

    os.Exit(result)
}

func Test_APIGateway(t *testing.T) {
	svc := apigateway.New(LOCALSTACK.CreateAWSSession())
	result, err := svc.GetRestApis(&apigateway.GetRestApisInput{})
	if err != nil {
		t.Error(err)
	}

    if len(result.Items) != 0 {
        t.Error("The number of Rest Apis returned should be zero.")
    }
}
func Test_Kinesis(t *testing.T) {
	svc := kinesis.New(LOCALSTACK.CreateAWSSession())
    result, err := svc.ListStreams(&kinesis.ListStreamsInput { })
    if err != nil {
        t.Error(err)
    }

    if len(result.StreamNames) != 0 {
        t.Error("The number of returned streams should be zero.")
    }
}
func Test_Dynamodb(t *testing.T) {
	svc := dynamodb.New(LOCALSTACK.CreateAWSSession())
    result, err := svc.ListTables(&dynamodb.ListTablesInput{ })
    if err != nil {
        t.Error(err)
    }

    if len(result.TableNames) != 0 {
        t.Error("The number of returned table names should be zero.")
    }
}
func Test_DynamoDBStreams(t *testing.T) {
	svc := dynamodbstreams.New(LOCALSTACK.CreateAWSSession())
    result, err := svc.ListStreams(&dynamodbstreams.ListStreamsInput{ })
    if err != nil {
        t.Error(err)
    }

    if len(result.Streams) != 0 {
        t.Error("The number of returned streams should be zero.")
    }
}
// TODO:
func Test_ES(t *testing.T) { }
func Test_S3(t *testing.T) {
	svc := s3.New(LOCALSTACK.CreateAWSSession())
	result, err := svc.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		t.Error(err)
	}

    if len(result.Buckets) != 0 {
        t.Error("The number of buckets returned should be zero.")
    }
}
func Test_Firehose(t *testing.T) {
	svc := firehose.New(LOCALSTACK.CreateAWSSession())
	result, err := svc.ListDeliveryStreams(&firehose.ListDeliveryStreamsInput{})
	if err != nil {
		t.Error(err)
	}

    if len(result.DeliveryStreamNames) != 0 {
        t.Error("The number of delivery streams returned should be zero.")
    }
}
func Test_Lambda(t *testing.T) {
	svc := lambda.New(LOCALSTACK.CreateAWSSession())
	result, err := svc.ListFunctions(&lambda.ListFunctionsInput{})
	if err != nil {
		t.Error(err)
	}

    if len(result.Functions) != 0 {
        t.Error("The number of funtions returned should be zero.")
    }
}
func Test_Sns(t *testing.T) {
	svc := sns.New(LOCALSTACK.CreateAWSSession())
	result, err := svc.ListTopics(&sns.ListTopicsInput{})
	if err != nil {
		t.Error(err)
	}

    if len(result.Topics) != 0 {
        t.Error("The number of topics should be zero.")
    }
}
func Test_Sqs(t *testing.T) {
	svc := sqs.New(LOCALSTACK.CreateAWSSession())
	result, err := svc.ListQueues(&sqs.ListQueuesInput{})
	if err != nil {
		t.Error(err)
	}

    if len(result.QueueUrls) != 0 {
        t.Error("The number of queues should be zero.")
    }
} 
func Test_Redshift(t *testing.T) {
	svc := redshift.New(LOCALSTACK.CreateAWSSession())
	result, err := svc.DescribeClusters(&redshift.DescribeClustersInput{})
	if err != nil {
		t.Error(err)
	}

    if len(result.Clusters) != 0 {
        t.Error("The number of clusters should be zero.")
    }
}
// TODO:
func Test_Email(t *testing.T) {}
// TODO:
func Test_Route53(t *testing.T) { }
// TODO:
func Test_Cloudformation(t *testing.T) { }
// TODO:
func Test_Monitoring(t *testing.T) { }
// TODO:
func Test_Ssm(t *testing.T) { }
// TODO:
func Test_Secretsmanager(t *testing.T) { }
// TODO:
func Test_States(t *testing.T) { }
// TODO:
func Test_Logs(t *testing.T) { }
// TODO:
func Test_Sts(t *testing.T) { }
// TODO:
func Test_Iam(t *testing.T) { }
