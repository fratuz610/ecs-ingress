package util

import (
	"fmt"

	"bitbucket.org/nnnco/rev-proxy/shared"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// Ec2Instance a ecs container instance cut down view
type Ec2Instance struct {
	InstanceID       string
	PrivateIPAddress string
	PrivateDNSName   string
	PublicIPAddress  string
}

// Ec2Client simplified client to access ECS resources on AWS
type Ec2Client struct {
	cfg     *shared.Config
	session *session.Session
	ec2Svc  *ec2.EC2
}

// NewEc2Client Creates a new ec2 client
func NewEc2Client(cfg *shared.Config) *Ec2Client {

	mySession := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{Region: aws.String(cfg.AWS.Region)},
	}))

	// Create a EC2/ECS clients from just a session

	ret := &Ec2Client{
		cfg:     cfg,
		session: mySession,
		ec2Svc:  ec2.New(mySession),
	}

	return ret
}

// DescribeInstances describes up ...
func (e *Ec2Client) DescribeInstances(instanceIDList []string) ([]Ec2Instance, error) {

	if len(instanceIDList) > 100 {
		return nil, fmt.Errorf("Unable to query more than 100 instances at a time")
	}

	retList := make([]Ec2Instance, 0)

	input := &ec2.DescribeInstancesInput{
		InstanceIds: aws.StringSlice(instanceIDList),
	}

	reply, err := e.ec2Svc.DescribeInstances(input)

	if err != nil {
		return nil, err
	}

	// we sift through the data
	for _, reservation := range reply.Reservations {

		for _, instance := range reservation.Instances {

			retList = append(retList, Ec2Instance{
				InstanceID:       aws.StringValue(instance.InstanceId),
				PrivateIPAddress: aws.StringValue(instance.PrivateIpAddress),
				PrivateDNSName:   aws.StringValue(instance.PrivateDnsName),
				PublicIPAddress:  aws.StringValue(instance.PublicIpAddress),
			})
		}

	}

	return retList, nil

}

// DescribeInstance describes a single instance
func (e *Ec2Client) DescribeInstance(instanceID string) (*Ec2Instance, error) {

	retList, err := e.DescribeInstances([]string{instanceID})

	if err != nil {
		return nil, err
	}

	if len(retList) < 1 {
		return nil, nil
	}

	return &retList[0], nil
}
