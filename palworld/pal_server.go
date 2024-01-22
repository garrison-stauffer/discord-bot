package palworld

import (
	"context"
	"fmt"
	"garrison-stauffer.com/discord-bot/environment"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/gorcon/rcon"
	"sync"
	"time"
)

type Server struct {
	instance *InstanceManager
	password string
	mux      *sync.Mutex
}

func NewServer(cfg aws.Config) *Server {
	fmt.Println("mmm " + environment.PalServerPassword())
	return &Server{
		NewInstanceManager(cfg),
		environment.PalServerPassword(),
		&sync.Mutex{},
	}
}

func (s *Server) IsServerReady(ctx context.Context) (bool, error) {
	instance, err := s.instance.GetInstanceMetadata(ctx)
	if err != nil {
		return false, err
	} else if !instance.IsRunning {
		return false, nil
	}

	palServerOn, err := s.checkRconConnectivity(instance.PublicIp)
	if err != nil {
		return false, fmt.Errorf("error checking rcon: %w", err)
	}

	return palServerOn, nil
}

func (s *Server) GetServerInfoIfReady(ctx context.Context) *InstanceMetadata {
	isReady, err := s.IsServerReady(ctx)

	if err != nil || !isReady {
		return nil
	}

	meta, err := s.instance.GetInstanceMetadata(ctx)
	if err != nil {
		fmt.Printf("error looking up server metadata for final server info: %v\n", err)
		return nil
	}
	return &meta
}

func (s *Server) StartServer(ctx context.Context) (*InstanceMetadata, error) {
	// lock actions to start and shutdown the server - critical code that needs synchronization
	s.mux.Lock()
	defer s.mux.Unlock()

	// get instance into a running state
	err := s.getInstanceToRunningState(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting instance into a running state: %w", err)
	}

	err = s.getPalServerToRunning(ctx)
	if err != nil {
		return nil, fmt.Errorf("error while starting the server code on a running ec2 instance: %w", err)
	}

	meta, _ := s.instance.GetInstanceMetadata(ctx)
	return &meta, nil
}

func (s *Server) StopServer(ctx context.Context) error {
	// lock actions to start and shutdown the server - critical code that needs synchronization
	s.mux.Lock()
	defer s.mux.Unlock()

	meta, err := s.instance.GetInstanceMetadata(ctx)
	if err != nil {
		return fmt.Errorf("error looking up instance status: %w", err)
	}

	if !meta.IsRunning {
		return nil
	}

	err = s.saveAndShutdownServerOverRcon(meta.PublicIp)
	if err != nil {
		return fmt.Errorf("error while saving and shutting down over rcon: %w")
	}

	err = s.instance.StopInstance(ctx)
	if err != nil {
		return fmt.Errorf("error while shutting down ec2 instance: %w", err)
	}

	return nil
}

func (s *Server) getInstanceToRunningState(ctx context.Context) error {
	meta, _ := s.instance.GetInstanceMetadata(ctx)

	if meta.Status == "running" {
		return nil
	} else if meta.Status == "pending" {
		// nothing to do here, it will reach running in the following block
	} else if meta.Status == "stopped" {
		_ = s.instance.StartInstance(ctx)
	} else if meta.Status == "stopping" {
		for i := 0; i < 180 && meta.Status != "stopped"; i++ {
			time.Sleep(2 * time.Second)
			meta, _ = s.instance.GetInstanceMetadata(ctx)
		}

		_ = s.instance.StartInstance(ctx)
	}

	for i := 0; i < 180 && meta.Status != "running"; i++ {
		time.Sleep(2 * time.Second)
		meta, _ = s.instance.GetInstanceMetadata(ctx)
	}

	if meta.Status != "running" {
		return fmt.Errorf("instance didn't reach a running state after 6-12 minutes")
	}

	return nil
}

func (s *Server) getPalServerToRunning(ctx context.Context) error {
	meta, _ := s.instance.GetInstanceMetadata(ctx)
	if meta.PublicIp == nil {
		return fmt.Errorf("unexpected state - public IP is not available")
	}

	isPalServerUp, _ := s.checkRconConnectivity(meta.PublicIp)
	if isPalServerUp {
		return nil
	}

	err := s.instance.StartPalServer(ctx)
	if err != nil {
		return fmt.Errorf("error while issuing SSM command to start the server: %w", err)
	}

	for i := 0; i < 180; i++ {
		time.Sleep(2 * time.Second)
		isPalServerUp, err = s.checkRconConnectivity(meta.PublicIp)
		if isPalServerUp {
			break
		}
	}

	if !isPalServerUp {
		return fmt.Errorf("pal server is not running after 6 minutes of pinging rcon: %w", err)
	}

	return nil
}

func (s *Server) checkRconConnectivity(ip *string) (bool, error) {
	conn, err := rcon.Dial(*ip+":25575", s.password)
	if err != nil {
		// assume dial fails because server is not running
		return false, nil
	}
	defer conn.Close()

	_, err = conn.Execute("Info")
	return err == nil, err
}

func (s *Server) saveAndShutdownServerOverRcon(ip *string) error {
	conn, err := rcon.Dial(*ip+":25575", s.password)
	if err != nil {
		// assume dial fails because server is not running - no op here
		return nil
	}

	_, err = conn.Execute("Save")
	_, err = conn.Execute("Shutdown 45")
	conn.Close()
	time.Sleep(60 * time.Second)

	return nil
}
