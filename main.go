package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/kinfkong/ikatago-client/config"
	"github.com/kinfkong/ikatago-client/katassh"
	"github.com/kinfkong/ikatago-client/platform"
	"github.com/kinfkong/ikatago-client/utils"
)

const (
	AppVersion = "1.3.0"
)

var opts struct {
	World           *string `short:"w" long:"world" description:"The world url."`
	Platform        string  `short:"p" long:"platform" description:"The platform, like aistudio, colab" required:"true"`
	Username        string  `short:"u" long:"username" description:"Your username to connect" required:"true"`
	Password        string  `long:"password" description:"Your password to connect" required:"true"`
	NoCompress      bool    `long:"no-compress" description:"compress the data during transmission"`
	RefreshInterval int     `long:"refresh-interval" description:"sets the refresh interval in cent seconds" default:"30"`
	TransmitMoveNum int     `long:"transmit-move-num" description:"limits number of moves when transmission during analyze" default:"20"`
	ConfigFile      *string `short:"c" long:"cli-config" description:"The config file of the client (not katago config file)"`
	KataLocalConfig *string `long:"kata-local-config" description:"The katago config file. like, gtp_example.cfg"`
	KataName        *string `long:"kata-name" description:"The katago binary name"`
	KataWeight      *string `long:"kata-weight" description:"The katago weight name"`
	KataConfig      *string `long:"kata-config" description:"The katago config name"`
	Command         string  `long:"cmd" description:"The command to run the katago" default:"run-katago"`
}

func parseArgs() {
	subcommands, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal("Cannot parse args: ", err)
	}
	config.Init(opts.ConfigFile)
	// overrides the config with args
	if opts.World != nil {
		config.GetConfig().Set("world.url", *opts.World)
	}
	if opts.KataName != nil {
		config.GetConfig().Set("cmd.kataName", *opts.KataName)
	}
	if opts.KataWeight != nil {
		config.GetConfig().Set("cmd.kataWeight", *opts.KataWeight)
	}
	if opts.KataConfig != nil {
		config.GetConfig().Set("cmd.kataConfig", *opts.KataConfig)
	}
	if opts.KataLocalConfig != nil {
		config.GetConfig().Set("cmd.kataLocalConfig", *opts.KataLocalConfig)
	}
	config.GetConfig().Set("user.name", opts.Username)
	config.GetConfig().Set("user.password", opts.Password)
	config.GetConfig().Set("platform.name", opts.Platform)
	config.GetConfig().Set("cmd.cmd", opts.Command)
	config.GetConfig().Set("cmd.subcommands", subcommands)
	config.GetConfig().Set("cmd.noCompress", opts.NoCompress)
	config.GetConfig().Set("cmd.refreshInterval", opts.RefreshInterval)
	config.GetConfig().Set("cmd.transmitMoveNum", opts.TransmitMoveNum)

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

func buildRunKatagoCommand() string {
	cmd := config.GetConfig().GetString("cmd.cmd")
	kataName := config.GetConfig().GetString("cmd.kataName")
	kataWeight := config.GetConfig().GetString("cmd.kataWeight")
	kataConfig := config.GetConfig().GetString("cmd.kataConfig")
	kataLocalConfig := config.GetConfig().GetString("cmd.kataLocalConfig")
	if len(kataName) > 0 {
		cmd = cmd + fmt.Sprintf(" --name %s", kataName)
	}
	if len(kataWeight) > 0 {
		cmd = cmd + fmt.Sprintf(" --weight %s", kataWeight)
	}
	if len(kataConfig) > 0 {
		cmd = cmd + fmt.Sprintf(" --config %s", kataConfig)
	}
	if len(kataLocalConfig) > 0 {
		cmd = cmd + fmt.Sprintf(" --custom-config %s", filepath.Base(kataLocalConfig))
	}

	if !config.GetConfig().GetBool("cmd.noCompress") {
		cmd = cmd + fmt.Sprintf(" --compress")
	}
	cmd = cmd + fmt.Sprintf(" --refresh-interval %d", config.GetConfig().GetInt("cmd.refreshInterval"))
	cmd = cmd + fmt.Sprintf(" --transmit-move-num %d", config.GetConfig().GetInt("cmd.transmitMoveNum"))

	subcommands := config.GetConfig().GetStringSlice("cmd.subcommands")
	if subcommands != nil && len(subcommands) > 0 {
		cmd = cmd + " -- " + strings.Join(subcommands, " ")
	}
	return cmd
}

func main() {
	fmt.Printf("ikatago version: %s\n", AppVersion)
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
	if len(config.GetConfig().GetString("cmd.kataLocalConfig")) > 0 {
		// copy the config file to remote
		err = katassh.RunSCP(*sshOptions, config.GetConfig().GetString("cmd.kataLocalConfig"))
		if err != nil {
			log.Fatal("Cannot copy config file to ikatago-server. ", err)
		}
	}

	err = katassh.RunKatago(*sshOptions, buildRunKatagoCommand())
	if err != nil {
		log.Fatal(err)
	}
}
