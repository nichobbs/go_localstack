package localstack

import (
	"fmt"
	"sort"
	"strings"
)

// LocalstackService defines a particular AWS service requested for a Localstack
// instances.  Note: You shold not create an instance of LocastackService directly.
// See: NewLocalstackService
//nolint:golint
type LocalstackService struct {
	// The name of the AWS Service. (I.E. "s3" or "apigateway")
	Name string
	// Protocol is the network protocol used for communication.
	Protocol string
	// Port is the port used when communicating with the service in the
	// Localstack instance.
	Port int
}

// Equals returns wether two pointers to a LocalstackService are equal.
func (service *LocalstackService) Equals(rhs *LocalstackService) bool {
	if service == nil && rhs == nil {
		return true
	} else if service == nil || rhs == nil {
		return false
	} else {
		return service.Name == rhs.Name &&
			service.Protocol == rhs.Protocol &&
			service.Port == rhs.Port
	}
}

// GetPortProtocol returns the protocol string (eg. 1234/tcp) used by Docker.
func (service *LocalstackService) GetPortProtocol() string {
	return fmt.Sprintf("%d/%s", service.Port, service.Protocol)
}

// GetNameProtocol returns the protocol string (eg. s3:1234) used by Docker.
func (service *LocalstackService) GetNamePort() string {
	return fmt.Sprintf("%s:%d", service.Name, service.Port)
}

// NewLocalstackService returns a new pointer to an instance of LocalstackService
// given the name of the service provided.  Note: The name must match an aws service
// from this list (https://docs.aws.amazon.com/cli/latest/reference/#available-services)
// and be a supported service by Localstack.
func NewLocalstackService(name string) (*LocalstackService, error) {
	services := []string{
		"apigateway",
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
		"secrestmanager",
		"stepfunctions",
		"logs",
		"sts",
		"iam",
	}
	for _, n := range services {
		if n == name {
			return &LocalstackService{
				Name:     name,
				Protocol: "tcp",
				Port:     4566,
			}, nil
		}
	}
	return nil, fmt.Errorf("unknown Localstack Service: %s", name)
}

// LocalstackServiceCollection represents a collection of LocalstackService objects.
//nolint:golint
type LocalstackServiceCollection []LocalstackService

// GetServiceMap returns a comma delimited string of all the AWS service
// names in the collection.
func (collection *LocalstackServiceCollection) GetServiceMap() string {
	var maps []string
	for _, element := range *collection {
		maps = append(maps, element.GetNamePort())
	}

	return strings.Join(maps, ",")
}

// Len returns the number of items in the collection.
func (collection LocalstackServiceCollection) Len() int {
	return len(collection)
}

// Swap will swap two items in the collection.
func (collection LocalstackServiceCollection) Swap(i, j int) {
	collection[i], collection[j] = collection[j], collection[i]
}

// Less compares two items in the collection.  This returns true if the instance
// at i is less than the instance at j.  Otherwise it will return false.
func (collection LocalstackServiceCollection) Less(i, j int) bool {
	return collection[i].Name < collection[j].Name
}

// Sort simply sorts the collection based on the names of the defined services.
// The collection returned is a pointer to the calling collection.
func (collection *LocalstackServiceCollection) Sort() *LocalstackServiceCollection {
	sort.Sort(collection)
	return collection
}

func (collection *LocalstackServiceCollection) Contains(name string) bool {
	for _, value := range *collection {
		if value.Name == name {
			return true
		}
	}

	return false
}
