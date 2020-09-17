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
	"errors"
	"fmt" 
	"strings"
	"bytes"
	"bufio"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"github.com/aws/aws-sdk-go/aws/endpoints"
    "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

// Localstack_Repository is the Localstack Docker repository
const Localstack_Repository string = "localstack/localstack"
// Localstack_Tag is the last tested version of the Localstack Docker repository
const Localstack_Tag string = "0.11.5"

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
		return errors.New(fmt.Sprintf("Could not connect to docker: %s", err))
	}

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(ls.Resource); err != nil {
		return errors.New(fmt.Sprintf("Could not purge resource: %s", err))
	}

	return nil
}

// EndpointResolver is necessary to route traffic to AWS services in your code to the Localstack
// endpoints.
func (l Localstack) EndpointFor(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
	services  := []string{"apigateway",
		"kinesis",
		"dynamodb",
		"dynamodbstreams",
		"es",
		"s3",
		"firehose",
		"lambda",
		"sns",
		"sqs",
		"redshift",
		"ses",
		"route53",
		"cloudformation",
		"cloudwatch",
		"ssm",
		"secretsmanager",
		"stepfunctions",
		"logs",
		"sts",
		"iam"}
	for _ ,v := range services {
		if v == service {
			return endpoints.ResolvedEndpoint { URL: fmt.Sprintf("http://%s", l.Resource.GetHostPort("4566/tcp")) }, nil
		}
	}
	return endpoints.DefaultResolver().EndpointFor(service, region, optFns...)
}

// CreateAWSSession should be used to make sure that your AWS SDK traffic is routing to Localstack correctly.
func (l *Localstack) CreateAWSSession() *session.Session {
	return session.Must(session.NewSession(&aws.Config{
        Region: aws.String("us-east-1"),
		EndpointResolver: *l,
		DisableSSL: aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
        Credentials: credentials.NewStaticCredentials("a", "b", "c"),
	}))
}

// NewLocalstack creates a new Localstack docker container based on the latest version.
func NewLocalstack(services *LocalstackServiceCollection) (*Localstack, error) {
	return NewSpecificLocalstack(services, "", Localstack_Repository, "latest")
}

func NewPersistentLocalstack(services *LocalstackServiceCollection,  data string) (*Localstack, error) {
	return NewSpecificLocalstack(services, "", Localstack_Repository, "latest")
}

// NewSpecificLocalstack creates a new Localstack docker container based on
// the given name, repository, and tag given.  NOTE:  The Docker image used should be a 
// Localstack image.  The behavior is unknown otherwise.  This method is provided
// to allow special situations like using a tag other than latest or when referencing 
// an internal Localstack image.
func NewSpecificLocalstack(services *LocalstackServiceCollection, name, repository, tag string) (*Localstack, error) {
	return NewPersistentSpecificLocalstack(services, name, repository, tag, "")
}

func NewPersistentSpecificLocalstack(services *LocalstackServiceCollection, name, repository, tag, data string) (*Localstack, error) {
	return newPersistentLocalstack(services, &_DockerWrapper{ }, name, repository, tag, data)
}

func getLocalstack(services *LocalstackServiceCollection, dockerWrapper DockerWrapper, name, repository, tag string) (*dockertest.Resource, error) {

    if name != "" {
        containers, err := dockerWrapper.ListContainers(docker.ListContainersOptions { All: true })
        if err != nil {
            return nil, errors.New(fmt.Sprintf("Unable to retrieve docker containers: %s", err))
        }
        for _, c := range containers {
            if c.Image == fmt.Sprintf("%s:%s", repository, tag) {
                for _,internalName := range c.Names {
                    if internalName == fmt.Sprintf("/%s", name) {
                        container, err := dockerWrapper.InspectContainer(c.ID)
                        if err !=  nil {
                            return nil, errors.New(fmt.Sprintf("Unable to inspect container %s: %s", c.ID, err))
                        }
                        return &dockertest.Resource{ Container: container }, nil
                    }
                }
            }
        }
    }

	return nil, nil
}

func newPersistentLocalstack(services *LocalstackServiceCollection, wrapper DockerWrapper, name, repository, tag string, data string) (*Localstack, error) {

	localstack, err := getLocalstack(services, wrapper, name, repository, tag)
	if err != nil {
		return nil, err	
	}

	if localstack == nil {

		// Fifth, If we didn't find a running container before, we spin one up now.
		options := &dockertest.RunOptions{
			Repository: repository,
			Tag: tag,
			Name: name, //If name == "", docker ignores it.
			Env: []string{
				fmt.Sprintf("SERVICES=%s", services.GetServiceMap()),

			},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"4566": {{
					HostPort: "4566",
				}},
			},
			ExposedPorts: []string{"4566"},

		}
		if len(data) > 0 {
			options.Env = append(options.Env, fmt.Sprintf("DATA_DIR=%s", data))
		}
		localstack, err = wrapper.RunWithOptions(options)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Could not start resource: %s", err))
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
				return errors.New(fmt.Sprintf("Unable to create a docker client: %s", err))
			}

			buffer := new(bytes.Buffer)

			logsOptions := docker.LogsOptions {
				Container: localstack.Container.ID,
				OutputStream: buffer,
				RawTerminal: true,
				Stdout: true,
				Stderr: true,
			}
			err = client.Logs(logsOptions)
			if err != nil {
				return errors.New(fmt.Sprintf("Unable to retrieve logs for container %s: %s", localstack.Container.ID, err))
			}

			scanner := bufio.NewScanner(buffer)
			for scanner.Scan() {
				token := strings.TrimSpace(scanner.Text())
				expected := "Ready."
				if strings.Contains(strings.TrimSpace(token),expected) {
					return nil
				}
			}
			if err := scanner.Err(); err != nil {
				return errors.New(fmt.Sprintf("Reading input: %s", err))
			}
			return errors.New("Not Ready")
		}); err != nil {
			return nil, errors.New(fmt.Sprintf("Unable to connect to %s: %s", service.Name, err))
		}
	}

	return &Localstack{
		Resource: localstack,
		Services: services,
	}, nil
}

