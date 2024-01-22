package palworld

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

const palworldEc2Instance = "i-00c202c075b0c9e48" // prod server
//const palworldEc2Instance = "i-09e495d8d2c5bffe5" // test server

type InstanceManager struct {
	instances *ec2.Client
	ssm       *ssm.Client
}

func NewInstanceManager(cfg aws.Config) *InstanceManager {
	return &InstanceManager{
		instances: ec2.NewFromConfig(cfg),
		ssm:       ssm.NewFromConfig(cfg),
	}
}

type InstanceMetadata struct {
	Id        string
	Status    string
	IsRunning bool
	// IPs may be nil if the instance is not in a running state
	PublicIp  *string
	PrivateIp *string
}

// GetInstanceMetadata fetches some information about the PalServer EC2 instance, namely whether it is running,
// and if so what IP ports it is available on.
func (i *InstanceManager) GetInstanceMetadata(ctx context.Context) (InstanceMetadata, error) {
	instances, err := i.instances.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{palworldEc2Instance},
	})

	if err != nil {
		return InstanceMetadata{}, fmt.Errorf("error describing instances: %w", err)
	}

	instance := instances.Reservations[0].Instances[0]
	return InstanceMetadata{
		Id:        *instance.InstanceId,
		Status:    string(instance.State.Name),
		IsRunning: instance.State.Name == types.InstanceStateNameRunning || instance.State.Name == types.InstanceStateNamePending,
		PublicIp:  instance.PublicIpAddress,
		PrivateIp: instance.PrivateIpAddress,
	}, nil
}

// IsInstanceRunning will check to see if the instance is running or in the process of starting according to EC2
func (i *InstanceManager) IsInstanceRunning(ctx context.Context) (bool, error) {
	instances, err := i.instances.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{palworldEc2Instance},
	})

	if err != nil {
		return false, fmt.Errorf("error describing instances: %w", err)
	}

	instanceState := instances.Reservations[0].Instances[0].State.Name
	return instanceState == types.InstanceStateNameRunning || instanceState == types.InstanceStateNamePending, nil
}

// StartInstance starts the EC2 instance, if it is not already in a running or provisioning state.
func (i *InstanceManager) StartInstance(ctx context.Context) error {
	if isRunning, _ := i.IsInstanceRunning(ctx); isRunning {
		return nil
	}

	_, err := i.instances.StartInstances(ctx, &ec2.StartInstancesInput{
		InstanceIds: []string{palworldEc2Instance},
	})

	return err
}

// StopInstance stops the EC2 instance. It assumes that the PalServer has been shut down appropriately using RCON already, but
// will check to see if the instance is running before issuing the command.
func (i *InstanceManager) StopInstance(ctx context.Context) error {
	if isRunning, err := i.IsInstanceRunning(ctx); !isRunning && err == nil {
		return nil
	}

	_, err := i.instances.StopInstances(ctx, &ec2.StopInstancesInput{
		InstanceIds: []string{palworldEc2Instance},
	})

	return err
}

// StartPalServer issues an SSM command to begin the server in a background task. It assumes the server is in a running state.
func (i *InstanceManager) StartPalServer(ctx context.Context) error {
	output, err := i.ssm.SendCommand(ctx, &ssm.SendCommandInput{
		InstanceIds:  []string{palworldEc2Instance},
		DocumentName: aws.String("StartPalServer"),
	})
	if err != nil {
		return fmt.Errorf("error starting the server: %w", err)
	}

	fmt.Printf("https://us-east-1.console.aws.amazon.com/systems-manager/run-command/%s?region=us-east-1\n", *output.Command.CommandId)
	return nil
}
