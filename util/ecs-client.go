package util

import (
	"fmt"

	"bitbucket.org/nnnco/rev-proxy/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
)

// EcsTask an ecs task cut down view
type EcsTask struct {
	TaskArn       string
	HostPort      int64
	ContainerPort int64
	InstanceArn   string
}

// EcsContainerInstance a ecs container instance cut down view
type EcsContainerInstance struct {
	EcsContainerInstanceArn string
	Ec2InstanceID           string
}

// EcsClient simplified client to access ECS resources on AWS
type EcsClient struct {
	cfg     *shared.Config
	session *session.Session
	ecsSvc  *ecs.ECS
	ec2Svc  *ec2.EC2
}

// NewEcsClient Creates a new ecs client
func NewEcsClient(cfg *shared.Config) *EcsClient {

	mySession := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{Region: aws.String(cfg.AWS.Region)},
	}))

	// Create a EC2/ECS clients from just a session

	ret := &EcsClient{
		cfg:     cfg,
		session: mySession,
		ecsSvc:  ecs.New(mySession),
		ec2Svc:  ec2.New(mySession),
	}

	return ret
}

// ListServices returns a list of services in the specified cluster
func (e *EcsClient) ListServices(clusterName string) ([]string, error) {

	var nextToken *string = nil

	retList := make([]string, 0)

	for {

		input := &ecs.ListServicesInput{
			Cluster:   &clusterName,
			NextToken: nextToken,
		}

		reply, err := e.ecsSvc.ListServices(input)

		if err != nil {
			return nil, err
		}

		retList = append(retList, aws.StringValueSlice(reply.ServiceArns)...)

		if reply.NextToken == nil {
			break
		}

		nextToken = reply.NextToken
	}

	return retList, nil
}

// ListServiceTasks returns a list of all tasks in a cluster
func (e *EcsClient) ListServiceTasks(clusterName string, serviceName string) ([]string, error) {

	var nextToken *string = nil

	retList := make([]string, 0)

	for {

		input := &ecs.ListTasksInput{
			Cluster:       &clusterName,
			ServiceName:   &serviceName,
			DesiredStatus: aws.String("RUNNING"),
			NextToken:     nextToken,
		}

		reply, err := e.ecsSvc.ListTasks(input)

		if err != nil {
			return nil, err
		}

		retList = append(retList, aws.StringValueSlice(reply.TaskArns)...)

		if reply.NextToken == nil {
			break
		}

		nextToken = reply.NextToken
	}

	return retList, nil
}

// DescribeTasks describes up to 100 tasks
func (e *EcsClient) DescribeTasks(clusterName string, taskArnList []string) ([]EcsTask, error) {

	if len(taskArnList) > 100 {
		return nil, fmt.Errorf("Unable to query more than 100 tasks at a time")
	}

	retList := make([]EcsTask, 0)

	input := &ecs.DescribeTasksInput{
		Cluster: &clusterName,
		Tasks:   aws.StringSlice(taskArnList),
	}

	reply, err := e.ecsSvc.DescribeTasks(input)

	if err != nil {
		return nil, err
	}

	// we sift through the data
	for _, task := range reply.Tasks {

		taskArn := aws.StringValue(task.TaskArn)

		// if no containers in the task?
		if len(task.Containers) < 1 {
			continue
		}

		firstContainer := *task.Containers[0]

		// if no network bindings in the task?
		if len(firstContainer.NetworkBindings) < 1 {
			continue
		}

		firstNetworkBinding := *firstContainer.NetworkBindings[0]

		retList = append(retList, EcsTask{
			TaskArn:       taskArn,
			HostPort:      aws.Int64Value(firstNetworkBinding.HostPort),
			ContainerPort: aws.Int64Value(firstNetworkBinding.ContainerPort),
			InstanceArn:   aws.StringValue(task.ContainerInstanceArn),
		})
	}

	return retList, nil
}

// DescribeTask describes a single task
func (e *EcsClient) DescribeTask(clusterName string, taskArn string) (*EcsTask, error) {

	retList, err := e.DescribeTasks(clusterName, []string{taskArn})

	if err != nil {
		return nil, err
	}

	if len(retList) < 1 {
		return nil, nil
	}

	return &retList[0], nil
}

// DescribeContainerInstances describes up to 100 container instances
func (e *EcsClient) DescribeContainerInstances(clusterName string, containerInstanceArnList []string) ([]EcsContainerInstance, error) {

	if len(containerInstanceArnList) > 100 {
		return nil, fmt.Errorf("Unable to query more than 100 container instances at a time")
	}

	retList := make([]EcsContainerInstance, 0)

	input := &ecs.DescribeContainerInstancesInput{
		Cluster:            &clusterName,
		ContainerInstances: aws.StringSlice(containerInstanceArnList),
	}

	reply, err := e.ecsSvc.DescribeContainerInstances(input)

	if err != nil {
		return nil, err
	}

	// we sift through the data
	for _, containerInstance := range reply.ContainerInstances {

		retList = append(retList, EcsContainerInstance{
			EcsContainerInstanceArn: aws.StringValue(containerInstance.ContainerInstanceArn),
			Ec2InstanceID:           aws.StringValue(containerInstance.Ec2InstanceId),
		})
	}

	return retList, nil

}

// DescribeContainerInstance describes a single container instance
func (e *EcsClient) DescribeContainerInstance(clusterName string, containerInstanceArn string) (*EcsContainerInstance, error) {

	retList, err := e.DescribeContainerInstances(clusterName, []string{containerInstanceArn})

	if err != nil {
		return nil, err
	}

	if len(retList) < 1 {
		return nil, nil
	}

	return &retList[0], nil
}
