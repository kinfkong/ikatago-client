package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"

	"github.com/kinfkong/ikatago-client/katassh"
	"github.com/kinfkong/ikatago-client/model"
	"github.com/kinfkong/ikatago-client/platform"
	"github.com/kinfkong/ikatago-client/utils"
)

// Options represents the client options
type Options struct {
	World    string `json:"world"`
	Platform string `json:"platform"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// RunKatagoOptions represents the run katago options
type RunKatagoOptions struct {
	NoCompress      bool
	RefreshInterval int
	TransmitMoveNum int
	KataLocalConfig *string
	KataName        *string
	KataWeight      *string
	KataConfig      *string
}

// Client represents the ikatago client
type Client struct {
	init       bool
	options    Options
	sshOptions model.SSHOptions
}

// NewClient creates the client
func NewClient(options Options) (*Client, error) {
	return &Client{
		options: options,
		init:    false,
	}, nil
}

// RunKatago runs the katago
func (client *Client) RunKatago(options RunKatagoOptions, subCommands []string, inputReader io.Reader, outputWriter io.Writer) error {
	if !client.init {
		err := client.initClient()
		if err != nil {
			return err
		}
	}
	if options.KataLocalConfig != nil {
		// run scp to copy the configure
		err := katassh.RunSCP(client.sshOptions, *options.KataLocalConfig)
		if err != nil {
			return err
		}
	}
	// build the ssh command
	err := katassh.RunKatago(client.sshOptions, buildRunKatagoCommand(options, subCommands), inputReader, outputWriter)
	if err != nil {
		return err
	}
	return nil
}

// QueryServer queries the server
func (client *Client) QueryServer(outputWriter io.Writer) error {
	if !client.init {
		err := client.initClient()
		if err != nil {
			return err
		}
	}

	// build the ssh command
	err := katassh.RunSSH(client.sshOptions, "query-server", outputWriter)
	if err != nil {
		return err
	}
	return nil
}

func buildRunKatagoCommand(options RunKatagoOptions, subCommands []string) string {
	cmd := "run-katago"
	kataName := options.KataName
	kataWeight := options.KataWeight
	kataConfig := options.KataConfig
	kataLocalConfig := options.KataLocalConfig
	if kataName != nil && len(*kataName) > 0 {
		cmd = cmd + fmt.Sprintf(" --name %s", *kataName)
	}
	if kataWeight != nil && len(*kataWeight) > 0 {
		cmd = cmd + fmt.Sprintf(" --weight %s", *kataWeight)
	}
	if kataConfig != nil && len(*kataConfig) > 0 {
		cmd = cmd + fmt.Sprintf(" --config %s", *kataConfig)
	}
	if kataLocalConfig != nil && len(*kataLocalConfig) > 0 {
		cmd = cmd + fmt.Sprintf(" --custom-config %s", filepath.Base(*kataLocalConfig))
	}

	if !options.NoCompress {
		cmd = cmd + fmt.Sprintf(" --compress")
	}
	cmd = cmd + fmt.Sprintf(" --refresh-interval %d", options.RefreshInterval)
	cmd = cmd + fmt.Sprintf(" --transmit-move-num %d", options.TransmitMoveNum)

	if subCommands != nil && len(subCommands) > 0 {
		cmd = cmd + " -- " + strings.Join(subCommands, " ")
	}
	return cmd
}

func (client *Client) initClient() error {
	platform, err := client.getPlatformFromWorld()
	if err != nil {
		return err
	}
	sshOptions, err := client.getSSHOptions(platform)
	if err != nil {
		return err
	}
	client.sshOptions = *sshOptions
	client.init = true
	return nil
}

func (client *Client) getPlatformFromWorld() (*platform.Platform, error) {
	type World struct {
		Platforms []platform.Platform `json:"platforms"`
	}
	worldJSONString, err := utils.DoHTTPRequest("GET", client.options.World, nil, nil)
	if err != nil {
		return nil, err
	}
	world := &World{}
	err = json.Unmarshal([]byte(worldJSONString), &world)
	if err != nil {
		return nil, err
	}
	for _, platform := range world.Platforms {
		if platform.Name == client.options.Platform {
			return &platform, nil
		}
	}
	log.Printf("ERROR platform not found in the world. platform: %s", client.options.Platform)
	return nil, errors.New("platform_not_found")
}

// getSSHOptions gets the ssh info
func (client *Client) getSSHOptions(p *platform.Platform) (*model.SSHOptions, error) {
	sshJSONURL := "https://" + p.Oss.Bucket + "." + p.Oss.BucketEndpoint + "/users/" + client.options.Username + ".ssh.json"
	response, err := utils.DoHTTPRequest("GET", sshJSONURL, nil, nil)
	if err != nil {
		log.Printf("ERROR error requestting url: %s, err: %+v\n", sshJSONURL, err)
		return nil, err
	}
	sshoptions := model.SSHOptions{}
	// parse json
	err = json.Unmarshal([]byte(response), &sshoptions)
	if err != nil {
		log.Printf("ERROR failed parsing json: %s\n", response)
		return nil, err
	}
	sshoptions.Password = client.options.Password
	return &sshoptions, nil
}
