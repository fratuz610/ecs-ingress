package service

import (
	"fmt"
	"strings"

	"bitbucket.org/nnnco/rev-proxy/shared"
	"bitbucket.org/nnnco/rev-proxy/util"
)

// EcsServiceDescr blah blah
type EcsServiceDescr struct {
	ClusterName  string
	ServiceArn   string
	ServiceName  string
	LocationList []EcsServiceIPPort
}

// EcsServiceIPPort blah blah
type EcsServiceIPPort struct {
	TaskArn          string
	PublicIPAddress  string
	PrivateIPAddress string
	Port             int64
}

// EcsService simplified client to access ECS resources on AWS
type EcsService struct {
	ecsClient *util.EcsClient
	ec2Client *util.Ec2Client
}

// NewEcsService Creates a new ecs client
func NewEcsService(cfg *shared.Config) *EcsService {

	ret := &EcsService{
		ecsClient: util.NewEcsClient(cfg),
		ec2Client: util.NewEc2Client(cfg),
	}

	return ret
}

// GetServicesAndPorts returns a list of services with relative tasks and ports
func (e *EcsService) GetServicesAndPorts(clusterName string) (map[string]EcsServiceDescr, error) {

	retMap := make(map[string]EcsServiceDescr, 0)

	// maps a ec2 instance id to an ec2 instance networking info
	ec2InstanceCache := make(map[string]util.Ec2Instance)

	serviceArnList, err := e.ecsClient.ListServices(clusterName)

	if err != nil {
		return nil, err
	}

	for _, serviceArn := range serviceArnList {

		serviceNameList := strings.Split(serviceArn, "/")

		ecsServiceDescr := EcsServiceDescr{
			ClusterName:  clusterName,
			ServiceName:  serviceNameList[len(serviceNameList)-1],
			ServiceArn:   serviceArn,
			LocationList: make([]EcsServiceIPPort, 0),
		}

		// for each service we query the tasks
		tmpTaskArnList, err := e.ecsClient.ListServiceTasks(clusterName, serviceArn)

		if err != nil {
			return nil, err
		}

		// for each task we get a description
		for _, taskArn := range tmpTaskArnList {

			// we then describe each task
			taskDescr, err := e.ecsClient.DescribeTask(clusterName, taskArn)

			if err != nil {
				return nil, err
			}

			if taskDescr == nil {
				// we ignore missing errors as transient tasks may have died already
				continue
			}

			// we determine the EC2 instance
			containerInstance, err := e.ecsClient.DescribeContainerInstance(clusterName, taskDescr.InstanceArn)

			if err != nil {
				return nil, err
			}

			if containerInstance == nil {
				return nil, fmt.Errorf("No container instance with arn %v - unable to proceed", taskDescr.InstanceArn)
			}

			ec2Instance, found := ec2InstanceCache[containerInstance.Ec2InstanceID]

			// caching
			if !found {

				localEc2Instance, err := e.ec2Client.DescribeInstance(containerInstance.Ec2InstanceID)

				if err != nil {
					return nil, err
				}

				if localEc2Instance == nil {
					return nil, fmt.Errorf("No ec2 instance with ID %v - unable to proceed", containerInstance.Ec2InstanceID)
				}

				ec2InstanceCache[containerInstance.Ec2InstanceID] = *localEc2Instance
				ec2Instance = *localEc2Instance
			}

			ecsServiceDescr.LocationList = append(ecsServiceDescr.LocationList, EcsServiceIPPort{
				TaskArn:          taskArn,
				PrivateIPAddress: ec2Instance.PrivateIPAddress,
				PublicIPAddress:  ec2Instance.PublicIPAddress,
				Port:             taskDescr.HostPort,
			})
		}

		retMap[ecsServiceDescr.ServiceName] = ecsServiceDescr
	}

	return retMap, nil
}
