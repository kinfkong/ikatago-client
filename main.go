package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/kinfkong/ikatago-client/client"
	"github.com/kinfkong/ikatago-client/ikatagosdk"
	"github.com/kinfkong/ikatago-client/utils"
)

const (
	AppVersion = "1.5.0"
)

var opts struct {
	World              *string `short:"w" long:"world" description:"The world url."`
	Platform           string  `short:"p" long:"platform" description:"The platform, like aistudio, colab" required:"true"`
	Username           string  `short:"u" long:"username" description:"Your username to connect" required:"true"`
	Password           string  `long:"password" description:"Your password to connect" required:"true"`
	NoCompress         bool    `long:"no-compress" description:"compress the data during transmission"`
	RefreshInterval    int     `long:"refresh-interval" description:"sets the refresh interval in cent seconds" default:"30"`
	EngineType         *string `long:"engine-type" description:"sets the enginetype"`
	Token              *string `long:"token" description:"sets the token"`
	GpuType            *string `long:"gpu-type" description:"sets the gpu type"`
	TransmitMoveNum    int     `long:"transmit-move-num" description:"limits number of moves when transmission during analyze" default:"20"`
	KataLocalConfig    *string `long:"kata-local-config" description:"The katago config file. like, gtp_example.cfg"`
	KataOverrideConfig *string `long:"kata-override-config" description:"The katago override-config, like: analysisPVLen=30,numSearchThreads=30"`

	KataName   *string `long:"kata-name" description:"The katago binary name"`
	ForceNode  *string `long:"force-node" description:"in cluster, force to a specific node."`
	KataWeight *string `long:"kata-weight" description:"The katago weight name"`
	KataConfig *string `long:"kata-config" description:"The katago config name"`
	Command    string  `long:"cmd" description:"The command to run the katago" default:"run-katago"`
}

func TestSDK() {
	client, _ := ikatagosdk.NewClient("", "all", "zz-xxxx", "834173")
	result, _ := client.QueryServer()
	log.Printf("DEBUG query result: %v", result)
}

/*func testSomething() {
	a := 1637380857
	log.Printf("%x", a)
}*/
func main() {
	// testSomething()
	// return
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
	} else {
		log.Fatal(fmt.Sprintf("Unknown command: [%s]", opts.Command))
	}
}
