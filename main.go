package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/kinfkong/ikatago-client/client"
	"github.com/kinfkong/ikatago-client/ikatagosdk"
	"github.com/kinfkong/ikatago-client/model"
	"github.com/kinfkong/ikatago-client/utils"
)

const (
	AppVersion = "1.6.1"
)

var opts = model.AllOpts{}

type MockDataCallback struct {
}

func (callback *MockDataCallback) Callback(content []byte) {
	log.Printf(string(content))

}
func (callback *MockDataCallback) StderrCallback(content []byte) {
	log.Printf(string(content))
}
func (callback *MockDataCallback) OnReady() {
}

func TestSDK() {
	client, _ := ikatagosdk.NewClient("", "all", "zz-xxxx", "xxxx")
	client.SetExtraArgs("--gpu-type 6x --kata-weight 60b")
	// query server
	result, _ := client.QueryServer()
	log.Printf("DEBUG query result: %v", result)
	// run katago
	runner, _ := client.CreateKatagoRunner()
	var callback ikatagosdk.DataCallback = &MockDataCallback{}
	go func() {
		time.Sleep(time.Second * 15)
		runner.SendGTPCommand("version\n")
	}()
	runner.Run(callback)

}

func main() {
	l := log.New(os.Stderr, "", 0)
	fmt.Fprintln(os.Stderr, "ikatago version: ", AppVersion)
	// parse args
	subCommands, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal("Cannot parse args: ", err)
	}
	defaultWorld := utils.WorldURL
	if opts.World == nil {
		opts.World = &defaultWorld
	}
	l.Printf("DEBUG the world is: %s\n", *opts.World)
	l.Printf("DEBUG Platform: [%s] User: [%s]\n", opts.Platform, opts.Username)
	remoteClient, err := client.NewClient(client.Options{
		World:      *opts.World,
		Platform:   opts.Platform,
		Username:   opts.Username,
		Password:   opts.Password,
		EngineType: opts.EngineType,
		ForceNode:  opts.ForceNode,
		GpuType:    opts.GpuType,
		Token:      opts.Token,
	})
	if err != nil {
		log.Fatal("Failed to create client.", err)
	}
	if opts.Command == "run-katago" {
		sessionResult, err := remoteClient.RunKatago(client.RunKatagoOptions{
			NoCompress:         opts.NoCompress,
			RefreshInterval:    opts.RefreshInterval,
			TransmitMoveNum:    opts.TransmitMoveNum,
			KataLocalConfig:    opts.KataLocalConfig,
			KataOverrideConfig: opts.KataOverrideConfig,
			KataConfig:         opts.KataConfig,
			KataWeight:         opts.KataWeight,
			KataName:           opts.KataName,
			UseRawData:         false,
		}, subCommands, os.Stdin, os.Stdout, os.Stderr, nil)
		if err != nil {
			log.Printf("ERROR run katago failed: %v", err)
			log.Fatal("Failed to run katago.", err)
		}
		sessionResult.Wait()
	} else if opts.Command == "query-server" {
		// run katago command
		err := remoteClient.QueryServer(os.Stdout)
		if err != nil {
			log.Fatal("Failed to query server.", err)
		}
	} else if opts.Command == "view-config" {
		err := remoteClient.ViewConfig(client.RunKatagoOptions{
			NoCompress:         opts.NoCompress,
			RefreshInterval:    opts.RefreshInterval,
			TransmitMoveNum:    opts.TransmitMoveNum,
			KataLocalConfig:    opts.KataLocalConfig,
			KataOverrideConfig: opts.KataOverrideConfig,
			KataConfig:         opts.KataConfig,
			KataWeight:         opts.KataWeight,
			KataName:           opts.KataName,
			UseRawData:         false,
		}, subCommands, os.Stdout)
		if err != nil {
			log.Printf("ERROR view katago config failed: %v", err)
			log.Fatal("Failed to view katago config.", err)
		}
	} else {
		log.Fatal(fmt.Sprintf("Unknown command: [%s]", opts.Command))
	}
}
