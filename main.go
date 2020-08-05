package main

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/jessevdk/go-flags"
	"github.com/kinfkong/ikatago-client/config"
	"github.com/kinfkong/ikatago-client/katassh"
	"github.com/kinfkong/ikatago-client/platform"
	"github.com/kinfkong/ikatago-client/utils"
)

var opts struct {
	World          *string `short:"w" long:"world" description:"The world url."`
	Platform       string  `short:"p" long:"platform" description:"The platform, like aistudio, colab" required:"true"`
	Username       string  `short:"u" long:"username" description:"Your username to connect" required:"true"`
	Password       string  `long:"password" description:"Your password to connect" required:"true"`
	ConfigFile     *string `short:"c" long:"cli-config" description:"The config file of the client (not katago config file)"`
	KataConfigFile *string `long:"kata-config" description:"The katago config file. like, gtp_example.cfg"`
	Command        string  `long:"cmd" description:"The command to run the katago" default:"run-katago"`
}

func parseArgs() {
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal("Cannot parse args", err)
	}
	config.Init(opts.ConfigFile)
	// overrides the config with args
	if opts.World != nil {
		config.GetConfig().Set("world.url", *opts.World)
	}
	config.GetConfig().Set("user.name", opts.Username)
	config.GetConfig().Set("user.password", opts.Password)
	config.GetConfig().Set("platform.name", opts.Platform)
	config.GetConfig().Set("cmd.cmd", opts.Command)
	log.Printf("DEBUG the world is: %s\n", config.GetConfig().GetString("world.url"))
	log.Printf("DEBUG Platform: [%s] User: [%s]\n", config.GetConfig().GetString("platform.name"), config.GetConfig().GetString("user.name"))
}

func getPlatformFromWorld() (*platform.Platform, error) {
	type World struct {
		Platforms []platform.Platform `json:"platforms"`
	}
	worldJSONString, err := utils.DoHTTPRequest("GET", config.GetConfig().GetString("world.url"), nil, nil)
	if err != nil {
		return nil, err
	}
	world := &World{}
	err = json.Unmarshal([]byte(worldJSONString), &world)
	if err != nil {
		return nil, err
	}
	for _, platform := range world.Platforms {
		if platform.Name == config.GetConfig().GetString("platform.name") {
			return &platform, nil
		}
	}
	log.Printf("ERROR platform not found in the world. platform: %s", config.GetConfig().GetString("platform.name"))
	return nil, errors.New("platform_not_found")
}

func main() {
	parseArgs()
	platform, err := getPlatformFromWorld()
	if err != nil {
		log.Fatal(err)
	}
	sshOptions, err := platform.GetSSHOptions(config.GetConfig().GetString("user.name"))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("DEBUG ssh info: %+v\n", *sshOptions)
	err = katassh.RunSSH(*sshOptions, "run-katago")
	if err != nil {
		log.Fatal(err)
	}
}
