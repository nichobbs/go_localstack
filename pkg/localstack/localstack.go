/*
This package was written to help writing tests with Localstack.
(https://github.com/localstack/localstack)  It uses libraries that help create
and manage a Localstack docker container for your go tests.

Requirements

    Go v1.11.0 or higher
    Docker (Tested on version 19.03.0-rc Community Edition)
*/
package localstack

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
)

// LocalstackRepository is the Localstack Docker repository
const LocalstackRepository string = "localstack/localstack"

// LocalstackTag is the last tested version of the Localstack Docker repository
const LocalstackTag string = "0.11.5"

// Localstack is a structure used to control the lifecycle of the Localstack
// Docker container.
type Localstack struct {
	// Resource is a pointer to the dockertest.Resource
	// object that is the localstack docker container.
	// (https://godoc.org/github.com/ory/dockertest#Resource)
	Resource *dockertest.Resource
	// Services is a pointer to a collection of service definitions
	// that are being requested from this particular instance of Localstack.
	Services *LocalstackServiceCollection
}

// Destroy simply shuts down and cleans up the Localstack container out of docker.
func (ls *Localstack) Destroy() error {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return fmt.Errorf("could not connect to docker: %s", err)
	}

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(ls.Resource); err != nil {
		return fmt.Errorf("could not purge resource: %s", err)
	}

	return nil
}

// EndpointResolver is necessary to route traffic to AWS services in your code to the Localstack
// endpoints.
func (ls Localstack) EndpointFor(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
	availableServices := map[string]string{
		"apigateway":       "apigateway",
		"kinesis":          "kinesis",
		"dynamodb":         "dynamodb",
		"streams.dynamodb": "dynamodbstreams",
		"es":               "es",
		"s3":               "s3",
		"firehose":         "firehose",
		"lambda":           "lambda",
		"sns":              "sns",
		"sqs":              "sqs",
		"redshift":         "redshift",
		"email":            "ses",
		"route53":          "route53",
		"cloudformation":   "cloudformation",
		"monitoring":       "cloudwatch",
		"ssm":              "ssm",
		"secretsmanager":   "secretsmanager",
		"states":           "stepfunctions",
		"logs":             "logs",
		"sts":              "sts",
		"iam":              "iam"}
	for k := range availableServices {
		if k == service && ls.Services.Contains(availableServices[service]) {
			return endpoints.ResolvedEndpoint{URL: fmt.Sprintf("http://%s", ls.Resource.GetHostPort("4566/tcp"))}, nil
		}
	}
	return endpoints.DefaultResolver().EndpointFor(service, region, optFns...)
}

// CreateAWSSession should be used to make sure that your AWS SDK traffic is routing to Localstack correctly.
func (ls *Localstack) CreateAWSSession() *session.Session {
	return session.Must(session.NewSession(&aws.Config{
		Region:           aws.String("us-east-1"),
		EndpointResolver: *ls,
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials("a", "b", "c"),
	}))
}

// NewLocalstack creates a new Localstack docker container based on the latest version.
func NewLocalstack(services *LocalstackServiceCollection) (*Localstack, error) {
	return NewSpecificLocalstack(services, "", LocalstackRepository, "latest")
}

func NewPersistentLocalstack(services *LocalstackServiceCollection, data string) (*Localstack, error) {
	return NewPersistentSpecificLocalstack(services, "", LocalstackRepository, "latest", data)
}

func NewNamedPersistentLocalstack(services *LocalstackServiceCollection, name, data string) (*Localstack, error) {
	return NewPersistentSpecificLocalstack(services, name, LocalstackRepository, "latest", data)
}

// NewSpecificLocalstack creates a new Localstack docker container based on
// the given name, repository, and tag given.  NOTE:  The Docker image used should be a
// Localstack image.  The behaviour is unknown otherwise.  This method is provided
// to allow special situations like using a tag other than latest or when referencing
// an internal Localstack image.
func NewSpecificLocalstack(services *LocalstackServiceCollection, name, repository, tag string) (*Localstack, error) {
	return NewPersistentSpecificLocalstack(services, name, repository, tag, "")
}

func NewPersistentSpecificLocalstack(services *LocalstackServiceCollection, name, repository, tag, data string) (*Localstack, error) {
	return newPersistentLocalstack(services, &_DockerWrapper{}, name, repository, tag, data)
}

func getLocalstack(_ *LocalstackServiceCollection, dockerWrapper DockerWrapper, name,
	repository, tag string) (*dockertest.Resource, error) {
	if name != "" {
		containers, err := dockerWrapper.ListContainers(docker.ListContainersOptions{All: true})
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve docker containers: %s", err)
		}
		//nolint:gocritic
		for _, c := range containers {
			if c.Image == fmt.Sprintf("%s:%s", repository, tag) {
				for _, internalName := range c.Names {
					if internalName == fmt.Sprintf("/%s", name) {
						container, err := dockerWrapper.InspectContainer(c.ID)
						if err != nil {
							return nil, fmt.Errorf("unable to inspect container %s: %s", c.ID, err)
						}
						return &dockertest.Resource{Container: container}, nil
					}
				}
			}
		}
	}

	return nil, nil
}

//nolint:unparam
func newLocalstack(services *LocalstackServiceCollection, wrapper DockerWrapper, name, repository, tag string) (*Localstack, error) {
	return newPersistentLocalstack(services, wrapper, name, repository, tag, "")
}

func newPersistentLocalstack(services *LocalstackServiceCollection, wrapper DockerWrapper,
	name, repository, tag, data string) (*Localstack, error) {
	localstack, err := getLocalstack(services, wrapper, name, repository, tag)
	if err != nil {
		return nil, err
	}

	if localstack == nil {
		// Fifth, If we didn't find a running container before, we spin one up now.
		options := &dockertest.RunOptions{
			Repository: repository,
			Tag:        tag,
			Name:       name, // If name == "", docker ignores it.
			Env: []string{
				fmt.Sprintf("SERVICES=%s", services.GetServiceMap()),
			},

			// PortBindings: map[docker.Port][]docker.PortBinding{
			//	"4566": {{
			//		HostPort: "4566",
			//	}},
			// },
			// ExposedPorts: []string{"4566"},

		}
		if len(data) > 0 {
			options.Env = append(options.Env, fmt.Sprintf("DATA_DIR=%s", data))
			options.Mounts = []string{"/tmp/localstack/data:/tmp/localstack/data"}
		}
		localstack, err = wrapper.RunWithOptions(options)
		if err != nil {
			return nil, fmt.Errorf("could not start resource: %s", err)
		}
	}

	// Sixth, we wait for the services to be ready before we allow the tests
	// to be run.
	for _, service := range *services {
		if err := wrapper.Retry(func() error {
			// We have to use a method that checks the output
			// of the docker container here because simply checking for
			// connetivity on the ports doesn't work.
			client, err := docker.NewClientFromEnv()
			if err != nil {
				return fmt.Errorf("unable to create a docker client: %s", err)
			}

			buffer := new(bytes.Buffer)

			logsOptions := docker.LogsOptions{
				Container:    localstack.Container.ID,
				OutputStream: buffer,
				RawTerminal:  true,
				Stdout:       true,
				Stderr:       true,
			}
			err = client.Logs(logsOptions)
			if err != nil {
				return fmt.Errorf("unable to retrieve logs for container %s: %s", localstack.Container.ID, err)
			}

			scanner := bufio.NewScanner(buffer)
			for scanner.Scan() {
				token := strings.TrimSpace(scanner.Text())
				expected := "Ready."
				if strings.Contains(strings.TrimSpace(token), expected) {
					return nil
				}
			}
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("reading input: %s", err)
			}
			return errors.New("not Ready")
		}); err != nil {
			return nil, fmt.Errorf("unable to connect to %s: %s", service.Name, err)
		}
	}

	return &Localstack{
		Resource: localstack,
		Services: services,
	}, nil
}
