package main

import (
	"context"
	"garrison-stauffer.com/discord-bot/palworld"
	"github.com/aws/aws-sdk-go-v2/config"
)

func main() {
	cfg, _ := config.LoadDefaultConfig(context.Background())
	instance := palworld.NewInstanceManager(cfg)
	instance.StopInstance(context.Background())
	//err := instance.StartPalServer(context.Background())
	//fmt.Printf("error from starting: %v\n", err)
	//
	//isRunning, err := instance.IsInstanceRunning(context.Background())
	//fmt.Printf("is running: %v err: %v\n", isRunning, err)
	//
	//meta, err := instance.GetInstanceMetadata(context.Background())
	//fmt.Printf("metadata: %+v err: %v\n", meta, err)
	//
	//err = instance.StopInstance(context.Background())
	//fmt.Printf("stopping err: %v\n", err)
	//
	//time.Sleep(500 * time.Millisecond)
	//meta, err = instance.GetInstanceMetadata(context.Background())
	//fmt.Printf("metadata: %+v err: %v\n", meta, err)
	//
	//time.Sleep(500 * time.Millisecond)
	//meta, err = instance.GetInstanceMetadata(context.Background())
	//fmt.Printf("metadata: %+v err: %v\n", meta, err)
	//
	//time.Sleep(500 * time.Millisecond)
	//meta, err = instance.GetInstanceMetadata(context.Background())
	//fmt.Printf("metadata: %+v err: %v\n", meta, err)
	//
	//time.Sleep(500 * time.Millisecond)
	//meta, err = instance.GetInstanceMetadata(context.Background())
	//fmt.Printf("metadata: %+v err: %v\n", meta, err)
	//
	//time.Sleep(500 * time.Millisecond)
	//meta, err = instance.GetInstanceMetadata(context.Background())
	//fmt.Printf("metadata: %+v err: %v\n", meta, err)
	//
	//isRunning, err = instance.IsInstanceRunning(context.Background())
	//fmt.Printf("is running: %v err: %v\n", isRunning, err)
	//
	//err = instance.StartInstance(context.Background())
	//fmt.Printf("starting instance err: %v\n", err)
	//
	//time.Sleep(500 * time.Millisecond)
	//meta, err = instance.GetInstanceMetadata(context.Background())
	//fmt.Printf("metadata: %+v err: %v\n", meta, err)
	//
	//time.Sleep(500 * time.Millisecond)
	//meta, err = instance.GetInstanceMetadata(context.Background())
	//fmt.Printf("metadata: %+v err: %v\n", meta, err)
	//
	//time.Sleep(500 * time.Millisecond)
	//meta, err = instance.GetInstanceMetadata(context.Background())
	//fmt.Printf("metadata: %+v err: %v\n", meta, err)
	//
	//time.Sleep(500 * time.Millisecond)
	//meta, err = instance.GetInstanceMetadata(context.Background())
	//fmt.Printf("metadata: %+v err: %v\n", meta, err)
}
