package localstack

import (
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/golang/mock/gomock"
	"github.com/nichobbs/go_localstack/pkg/mock_localstack"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
)

const (
	LocalstackName = "testLocalstackName"
	defaultURL     = "http://1.0.0.0:9566"
)

var portBindings = map[docker.Port][]docker.PortBinding{
	"4566/tcp": {{HostIP: "1.0.0.0", HostPort: "9566"}},
	"4567/tcp": {{HostIP: "1.0.0.0", HostPort: "9567"}},
	"4568/tcp": {{HostIP: "1.0.0.0", HostPort: "9568"}},
	"4569/tcp": {{HostIP: "1.0.0.0", HostPort: "9569"}},
	"4570/tcp": {{HostIP: "1.0.0.0", HostPort: "9570"}},
	"4571/tcp": {{HostIP: "1.0.0.0", HostPort: "9571"}},
	"4572/tcp": {{HostIP: "1.0.0.0", HostPort: "9572"}},
	"4573/tcp": {{HostIP: "1.0.0.0", HostPort: "9573"}},
	"4574/tcp": {{HostIP: "1.0.0.0", HostPort: "9574"}},
	"4575/tcp": {{HostIP: "1.0.0.0", HostPort: "9575"}},
	"4576/tcp": {{HostIP: "1.0.0.0", HostPort: "9576"}},
	"4577/tcp": {{HostIP: "1.0.0.0", HostPort: "9577"}},
	"4579/tcp": {{HostIP: "1.0.0.0", HostPort: "9579"}},
	"4580/tcp": {{HostIP: "1.0.0.0", HostPort: "9580"}},
	"4581/tcp": {{HostIP: "1.0.0.0", HostPort: "9581"}},
	"4582/tcp": {{HostIP: "1.0.0.0", HostPort: "9582"}},
	"4583/tcp": {{HostIP: "1.0.0.0", HostPort: "9583"}},
	"4584/tcp": {{HostIP: "1.0.0.0", HostPort: "9584"}},
	"4585/tcp": {{HostIP: "1.0.0.0", HostPort: "9585"}},
	"4586/tcp": {{HostIP: "1.0.0.0", HostPort: "9586"}},
	"4592/tcp": {{HostIP: "1.0.0.0", HostPort: "9592"}},
	"4593/tcp": {{HostIP: "1.0.0.0", HostPort: "9593"}},
}

var endpointIds = []string{
	endpoints.IamServiceID,
	endpoints.StsServiceID,
	endpoints.LogsServiceID,
	endpoints.StatesServiceID,
	endpoints.SecretsmanagerServiceID,
	endpoints.SsmServiceID,
	endpoints.MonitoringServiceID,
	endpoints.CloudformationServiceID,
	endpoints.Route53ServiceID,
	endpoints.EmailServiceID,
	endpoints.RedshiftServiceID,
	endpoints.SqsServiceID,
	endpoints.SnsServiceID,
	endpoints.LambdaServiceID,
	endpoints.FirehoseServiceID,
	endpoints.ApigatewayServiceID,
	endpoints.ApigatewayServiceID,
	endpoints.KinesisServiceID,
	endpoints.DynamodbServiceID,
	endpoints.StreamsDynamodbServiceID,
	endpoints.EsServiceID,
	endpoints.S3ServiceID,
}

func getLocalstackFound(services *LocalstackServiceCollection,
	ctrl *gomock.Controller) (*mock_localstack.MockDockerWrapper, *docker.Container) {
	m := mock_localstack.NewMockDockerWrapper(ctrl)
	container := &docker.Container{
		Config: &docker.Config{
			Env: []string{
				fmt.Sprintf("SERVICES=%s", services.GetServiceMap()),
			},
		},
	}

	m.
		EXPECT().
		ListContainers(gomock.Any()).
		Times(1).
		Return([]docker.APIContainers{
			{
				Image: fmt.Sprintf("%s:%s", LocalstackRepository, LocalstackTag),
				Names: []string{fmt.Sprintf("/%s", LocalstackName)},
			},
		}, nil)

	m.
		EXPECT().
		InspectContainer(gomock.Any()).
		Times(1).
		Return(container, nil)

	return m, container
}

func getLocalstackEmpty(_ *LocalstackServiceCollection, ctrl *gomock.Controller) *mock_localstack.MockDockerWrapper {
	m := mock_localstack.NewMockDockerWrapper(ctrl)

	m.
		EXPECT().
		ListContainers(gomock.Any()).
		Times(1).
		Return(nil, nil)

	return m
}

func Test_getLocalstack_ErrorWithListContainers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mock_localstack.NewMockDockerWrapper(ctrl)

	m.
		EXPECT().
		ListContainers(gomock.Any()).
		Times(1).
		Return(nil, errors.New("dummy Error"))

	m.
		EXPECT().
		InspectContainer(gomock.Any()).
		Times(0).
		Return(nil, nil)

	sqs, _ := NewLocalstackService("sqs")
	services := &LocalstackServiceCollection{
		*sqs,
	}
	actual, err := getLocalstack(services, m, LocalstackName, LocalstackRepository, LocalstackTag)

	if actual != nil {
		log.Fatal("We're expecting the localstack result to be nil.")
	}

	if err == nil {
		log.Fatal("We're expecting the error returned to be populated.")
	}
}

func Test_getLocalstack_UnknownImage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mock_localstack.NewMockDockerWrapper(ctrl)

	m.
		EXPECT().
		ListContainers(gomock.Any()).
		Times(1).
		Return([]docker.APIContainers{
			{Image: "DummyImage:1.0.0"},
		}, nil)

	m.
		EXPECT().
		InspectContainer(gomock.Any()).
		Times(0).
		Return(nil, nil)

	sqs, _ := NewLocalstackService("sqs")
	services := &LocalstackServiceCollection{
		*sqs,
	}
	actual, err := getLocalstack(services, m, LocalstackName, LocalstackRepository, LocalstackTag)

	if actual != nil || err != nil {
		log.Fatal("We're expecting both the localstack and error return results to be nil.")
	}
}

func Test_getLocalstack_ErrorWithInspectContainer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mock_localstack.NewMockDockerWrapper(ctrl)

	m.
		EXPECT().
		ListContainers(gomock.Any()).
		Times(1).
		Return([]docker.APIContainers{
			{
				Image: fmt.Sprintf("%s:%s", LocalstackRepository, LocalstackTag),
				Names: []string{fmt.Sprintf("/%s", LocalstackName)},
			},
		}, nil)

	m.
		EXPECT().
		InspectContainer(gomock.Any()).
		Times(1).
		Return(nil, errors.New("dummy Error"))

	sqs, _ := NewLocalstackService("sqs")
	services := &LocalstackServiceCollection{
		*sqs,
	}
	actual, err := getLocalstack(services, m, LocalstackName, LocalstackRepository, LocalstackTag)

	if actual != nil {
		log.Fatal("We're expecting the localstack result to be nil.")
	}

	if err == nil {
		log.Fatal("We're expecting the error returned to be populated.")
	}
}

func Test_getLocalstack_ContainerExistsButHasDifferentName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mock_localstack.NewMockDockerWrapper(ctrl)

	// Setup call to ListContainers
	m.
		EXPECT().
		ListContainers(gomock.Any()).
		Times(1).
		Return([]docker.APIContainers{
			{
				Image: fmt.Sprintf("%s:%s", LocalstackRepository, LocalstackTag),
				Names: []string{"DummyName"},
			},
		}, nil)

	m.
		EXPECT().
		InspectContainer(gomock.Any()).
		Times(0)

	sqs, _ := NewLocalstackService("sqs")
	services := &LocalstackServiceCollection{
		*sqs,
	}
	actual, err := getLocalstack(services, m, LocalstackName, LocalstackRepository, LocalstackTag)

	if actual != nil {
		log.Fatal("We're expecting the localstack result to be nil.")
	}

	if err != nil {
		log.Fatal("We're not expecting an error here.")
	}
}

func Test_getLocalstack_UnableToFindContainer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mock_localstack.NewMockDockerWrapper(ctrl)

	m.
		EXPECT().
		ListContainers(gomock.Any()).
		Times(1).
		Return(nil, nil)

	m.
		EXPECT().
		InspectContainer(gomock.Any()).
		Times(0).
		Return(nil, nil)

	sqs, _ := NewLocalstackService("sqs")
	services := &LocalstackServiceCollection{
		*sqs,
	}
	actual, err := getLocalstack(services, m, LocalstackName, LocalstackRepository, LocalstackTag)

	if actual != nil || err != nil {
		log.Fatal("We're expecting both the localstack and error return results to be nil.")
	}
}

func Test_getLocalstack_ContainerExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sqs, _ := NewLocalstackService("sqs")
	services := &LocalstackServiceCollection{
		*sqs,
	}
	m, c := getLocalstackFound(services, ctrl)

	actual, err := getLocalstack(services, m, LocalstackName, LocalstackRepository, LocalstackTag)

	if err != nil {
		log.Fatal("We're expecting the error returned to be nil.")
	}

	if actual.Container != c {
		log.Fatal("The actual result doesn't match what was expected.")
	}
}

func Test_getLocalstack_ContainerExists_WithinMultiple(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sqs, _ := NewLocalstackService("sqs")
	services := &LocalstackServiceCollection{
		*sqs,
	}

	m := mock_localstack.NewMockDockerWrapper(ctrl)
	container := &docker.Container{
		Config: &docker.Config{
			Env: []string{
				fmt.Sprintf("SERVICES=%s", services.GetServiceMap()),
			},
		},
	}

	m.
		EXPECT().
		ListContainers(gomock.Any()).
		Times(1).
		Return([]docker.APIContainers{
			{
				Image: fmt.Sprintf("%s:%s", LocalstackRepository, LocalstackTag),
				Names: []string{fmt.Sprintf("/%s", LocalstackName)},
			},
			{
				Image: fmt.Sprintf("%s:%s", LocalstackRepository, LocalstackTag),
				Names: []string{"/DummyContainer"},
			},
		}, nil)

	m.
		EXPECT().
		InspectContainer(gomock.Any()).
		Times(1).
		Return(container, nil)

	actual, err := getLocalstack(services, m, "DummyContainer", LocalstackRepository, LocalstackTag)

	if err != nil {
		log.Fatal("We're expecting the error returned to be nil.")
	}

	if actual.Container != container {
		log.Fatal("The actual result doesn't match what was expected.")
	}
}

func Test_NewLocalstack_GetLocalstackReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	sqs, _ := NewLocalstackService("sqs")
	services := &LocalstackServiceCollection{
		*sqs,
	}
	m := mock_localstack.NewMockDockerWrapper(ctrl)

	m.
		EXPECT().
		ListContainers(gomock.Any()).
		Times(1).
		Return(nil, errors.New("dummy Error"))

	result, err := newLocalstack(services, m, LocalstackName, LocalstackRepository, LocalstackTag)

	if result != nil {
		log.Fatal("We were expecting the returned container to be nil.")
	}

	if err == nil {
		log.Fatal("We were expecting the returned error to be populated.")
	}
}

func Test_NewLocalstack_GetLocalstackReturnsResult(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	sqs, _ := NewLocalstackService("sqs")
	s3, _ := NewLocalstackService("s3")
	services := &LocalstackServiceCollection{
		*sqs,
		*s3,
	}
	m, c := getLocalstackFound(services, ctrl)

	m.
		EXPECT().
		RunWithOptions(gomock.Any()).
		Times(0)

	m.
		EXPECT().
		Retry(gomock.Any()).
		Times(2).
		Return(nil)

	result, err := newLocalstack(services, m, LocalstackName, LocalstackRepository, LocalstackTag)

	if err != nil {
		log.Fatal("We were expecting the returned error to be nil.")
	}

	if result.Resource.Container != c {
		log.Fatal("The actual result doesn't match what was expected.")
	}
}

func Test_NewLocalstack_GetLocalstackReturnsResult_RetryFailsOnFirstService(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	sqs, _ := NewLocalstackService("sqs")
	s3, _ := NewLocalstackService("s3")
	services := &LocalstackServiceCollection{
		*sqs,
		*s3,
	}
	m, _ := getLocalstackFound(services, ctrl)

	m.
		EXPECT().
		RunWithOptions(gomock.Any()).
		Times(0)

	gomock.InOrder(
		m.
			EXPECT().
			Retry(gomock.Any()).
			Return(errors.New("dummyError")),
		m.
			EXPECT().
			Retry(gomock.Any()).
			Times(0).
			Return(nil),
	)

	result, err := newLocalstack(services, m, LocalstackName, LocalstackRepository, LocalstackTag)

	if result != nil {
		log.Fatal("We were expecting the returned container to be nil.")
	}

	if err == nil {
		log.Fatal("We were expecting the returned error to be populated.")
	}
}

func Test_NewLocalstack_GetLocalstackReturnsResult_RetryFailsOnSecondtService(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	sqs, _ := NewLocalstackService("sqs")
	s3, _ := NewLocalstackService("s3")
	services := &LocalstackServiceCollection{
		*sqs,
		*s3,
	}
	m, _ := getLocalstackFound(services, ctrl)

	m.
		EXPECT().
		RunWithOptions(gomock.Any()).
		Times(0)

	gomock.InOrder(
		m.
			EXPECT().
			Retry(gomock.Any()).
			Return(nil),
		m.
			EXPECT().
			Retry(gomock.Any()).
			Return(errors.New("dummyError")),
	)

	result, err := newLocalstack(services, m, LocalstackName, LocalstackRepository, LocalstackTag)

	if result != nil {
		log.Fatal("We were expecting the returned container to be nil.")
	}

	if err == nil {
		log.Fatal("We were expecting the returned error to be populated.")
	}
}

func Test_NewLocalstack_StartFreshContainer_RunWithOptionsReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	sqs, _ := NewLocalstackService("sqs")
	s3, _ := NewLocalstackService("s3")
	services := &LocalstackServiceCollection{
		*sqs,
		*s3,
	}
	m := getLocalstackEmpty(services, ctrl)

	m.
		EXPECT().
		RunWithOptions(gomock.Any()).
		Times(1).
		Return(nil, errors.New("dummy Error"))

	m.
		EXPECT().
		Retry(gomock.Any()).
		Times(0).
		Return(nil)

	result, err := newLocalstack(services, m, LocalstackName, LocalstackRepository, LocalstackTag)

	if result != nil {
		log.Fatal("We were expecting the returned container to be nil.")
	}

	if err == nil {
		log.Fatal("We were expecting the returned error to be populated.")
	}
}

func Test_NewLocalstack_StartFreshContainer_RunWithOptions(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	sqs, _ := NewLocalstackService("sqs")
	s3, _ := NewLocalstackService("s3")
	services := &LocalstackServiceCollection{
		*sqs,
		*s3,
	}
	m := getLocalstackEmpty(services, ctrl)
	resource := &dockertest.Resource{}

	m.
		EXPECT().
		RunWithOptions(gomock.Any()).
		Times(1).
		Return(resource, nil)

	m.
		EXPECT().
		Retry(gomock.Any()).
		Times(2).
		Return(nil)

	result, err := newLocalstack(services, m, LocalstackName, LocalstackRepository, LocalstackTag)

	if err != nil {
		log.Fatal("We were expecting the returned error to be nil.")
	}

	if result == nil {
		log.Fatal("We were expecting the returned container to be populated.")
	}

	if result.Services != services {
		log.Fatal("The returned collection of services doesn't match what we sent in.")
	}

	if result.Resource != resource {
		log.Fatal("The returned resource is not what is expected.")
	}
}

func Test_EndpointFor(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	apigateway, _ := NewLocalstackService("apigateway")
	kinesis, _ := NewLocalstackService("kinesis")
	dynamodb, _ := NewLocalstackService("dynamodb")
	dynamodbstreams, _ := NewLocalstackService("dynamodbstreams")
	es, _ := NewLocalstackService("es")
	s3, _ := NewLocalstackService("s3")
	firehose, _ := NewLocalstackService("firehose")
	lambda, _ := NewLocalstackService("lambda")
	sns, _ := NewLocalstackService("sns")
	sqs, _ := NewLocalstackService("sqs")
	redshift, _ := NewLocalstackService("redshift")
	email, _ := NewLocalstackService("ses")
	route53, _ := NewLocalstackService("route53")
	cloudformation, _ := NewLocalstackService("cloudformation")
	cloudwatch, _ := NewLocalstackService("cloudwatch")
	ssm, _ := NewLocalstackService("ssm")
	secretsmanager, _ := NewLocalstackService("secretsmanager")
	stepfunctions, _ := NewLocalstackService("stepfunctions")
	logs, _ := NewLocalstackService("logs")
	sts, _ := NewLocalstackService("sts")
	iam, _ := NewLocalstackService("iam")
	services := &LocalstackServiceCollection{
		*apigateway,
		*kinesis,
		*dynamodb,
		*dynamodbstreams,
		*es,
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
	m, c := getLocalstackFound(services, ctrl)

	m.
		EXPECT().
		RunWithOptions(gomock.Any()).
		Times(0)

	m.
		EXPECT().
		Retry(gomock.Any()).
		Times(21).
		Return(nil)

	result, err := newLocalstack(services, m, LocalstackName, LocalstackRepository, LocalstackTag)

	if err != nil {
		log.Fatal("We were expecting the returned error to be nil.")
	}

	if result.Resource.Container != c {
		log.Fatal("The actual result doesn't match what was expected.")
	}

	if result.Services == nil {
		log.Fatal("The Services property of the Localstack object should not be nil.")
	}

	c.NetworkSettings = &docker.NetworkSettings{
		Ports: portBindings,
	}

	opt := func(opts *endpoints.Options) {}

	for _, e := range endpointIds {
		ep, _ := result.EndpointFor(e, "us-west-2", opt)

		if ep.URL != defaultURL {
			t.Errorf("The return URL was not correct.  Received %s", ep.URL)
		}
	}
}

func Test_EndpointFor_OnlyRegisteredServices(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	sqs, _ := NewLocalstackService("sqs")
	s3, _ := NewLocalstackService("s3")
	services := &LocalstackServiceCollection{
		*sqs,
		*s3,
	}
	m, c := getLocalstackFound(services, ctrl)

	m.
		EXPECT().
		RunWithOptions(gomock.Any()).
		Times(0)

	m.
		EXPECT().
		Retry(gomock.Any()).
		Times(2).
		Return(nil)

	result, err := newLocalstack(services, m, LocalstackName, LocalstackRepository, LocalstackTag)

	if err != nil {
		log.Fatal("We were expecting the returned error to be nil.")
	}

	if result.Resource.Container != c {
		log.Fatal("The actual result doesn't match what was expected.")
	}

	c.NetworkSettings = &docker.NetworkSettings{
		Ports: portBindings,
	}

	opt := func(opts *endpoints.Options) {}

	for _, e := range endpointIds {
		ep, _ := result.EndpointFor(e, "us-west-2", opt)
		if ep.URL != defaultURL {
			t.Errorf("The return URL was not correct.  Received %s", ep.URL)
		}
	}
}

func Test_CreateAWSSession(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	sqs, _ := NewLocalstackService("sqs")
	s3, _ := NewLocalstackService("s3")
	services := &LocalstackServiceCollection{
		*sqs,
		*s3,
	}
	m, c := getLocalstackFound(services, ctrl)

	m.
		EXPECT().
		RunWithOptions(gomock.Any()).
		Times(0)

	m.
		EXPECT().
		Retry(gomock.Any()).
		Times(2).
		Return(nil)

	result, err := newLocalstack(services, m, LocalstackName, LocalstackRepository, LocalstackTag)

	if err != nil {
		log.Fatal("We were expecting the returned error to be nil.")
	}

	if result.Resource.Container != c {
		log.Fatal("The actual result doesn't match what was expected.")
	}

	sess := result.CreateAWSSession()
	if *sess.Config.Region != *aws.String("us-east-1") {
		t.Errorf("The region returned was not what was expected:  %s", *sess.Config.Region)
	}
	if *sess.Config.DisableSSL != *aws.Bool(true) {
		t.Error("The DisableSSL value should be true")
	}
	if *sess.Config.S3ForcePathStyle != *aws.Bool(true) {
		t.Error("The S3ForcePathStyle value should be true")
	}

	// This is weird.  I don't know how to compare function pointers in Golang.
	// So, I'm just making sure it's not nil right now until I can figure this out.
	// TODO: Do a better job making sure that the resulting function pointer is accurate.
	if sess.Config.EndpointResolver == nil {
		t.Error("The resulting Resolver shouldn't be nil")
	}
}
