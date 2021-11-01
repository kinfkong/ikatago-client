package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kinfkong/ikatago-client/katassh"
	"github.com/kinfkong/ikatago-client/model"
	"github.com/kinfkong/ikatago-client/platform"
	"github.com/kinfkong/ikatago-client/utils"
)

// Options represents the client options
type Options struct {
	World      string  `json:"world"`
	Platform   string  `json:"platform"`
	Username   string  `json:"username"`
	Password   string  `json:"password"`
	EngineType *string `json:"engineType"`
}

// RunKatagoOptions represents the run katago options
type RunKatagoOptions struct {
	NoCompress         bool
	RefreshInterval    int
	TransmitMoveNum    int
	KataLocalConfig    *string
	EngineType         *string
	KataName           *string
	KataWeight         *string
	KataConfig         *string
	KataOverrideConfig *string
	UseRawData         bool
	ForceNode          *string
}

// Client represents the ikatago client
type Client struct {
	init       bool
	options    Options
	sshOptions model.SSHOptions
}

type SessionResult struct {
	session *katassh.KataSSHSession
	Err     error
	wg      sync.WaitGroup
}

func (s *SessionResult) Stop() {
	if s.session != nil {
		s.session.Stop()
		s.session = nil
	}
}
func (s *SessionResult) Wait() {
	s.wg.Wait()
}

// NewClient creates the client
func NewClient(options Options) (*Client, error) {
	return &Client{
		options: options,
		init:    false,
	}, nil
}

// RunKatago runs the katago
func (client *Client) RunKatago(options RunKatagoOptions, subCommands []string, inputReader io.Reader, outputWriter io.Writer, stderrWriter io.Writer, onReady func()) (*SessionResult, error) {
	if !client.init {
		err := client.initClient()
		if err != nil {
			return nil, err
		}
	}
	if options.KataLocalConfig != nil {
		// run scp to copy the configure
		err := (&katassh.KataSSHSession{}).RunSCP(client.sshOptions, *options.KataLocalConfig, options.EngineType)
		if err != nil {
			return nil, err
		}
	}
	s := &katassh.KataSSHSession{}
	result := SessionResult{
		session: s,
	}
	// build the ssh command
	result.wg.Add(1)
	go func() {
		err := s.RunKatago(client.sshOptions, buildRunKatagoCommand(options, subCommands), inputReader, outputWriter, stderrWriter, options.UseRawData, onReady)
		if err != nil {
			result.Err = err
		}
		result.wg.Done()
	}()

	return &result, nil
}

// QueryServer queries the server
func (client *Client) QueryServer(outputWriter io.Writer) error {
	if !client.init {
		err := client.initClient()
		if err != nil {
			return err
		}
	}
	engineType := client.options.EngineType
	// build the ssh command
	cmd := "query-server"
	if engineType != nil && len(*engineType) > 0 {
		cmd = cmd + " --engine-type " + *engineType
	}
	err := (&katassh.KataSSHSession{}).RunSSH(client.sshOptions, cmd, outputWriter)
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) SetEngineType(engineType *string) {
	client.options.EngineType = engineType
}

func (client *Client) GetEngineType() *string {
	return client.options.EngineType
}

func buildRunKatagoCommand(options RunKatagoOptions, subCommands []string) string {
	cmd := "run-katago"
	kataName := options.KataName
	kataWeight := options.KataWeight
	kataConfig := options.KataConfig
	kataLocalConfig := options.KataLocalConfig
	kataOverrideConfig := options.KataOverrideConfig
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
		cmd = cmd + " --compress"
	}
	cmd = cmd + fmt.Sprintf(" --refresh-interval %d", options.RefreshInterval)
	cmd = cmd + fmt.Sprintf(" --transmit-move-num %d", options.TransmitMoveNum)
	if options.EngineType != nil && len(*options.EngineType) > 0 {
		cmd = cmd + fmt.Sprintf(" --engine-type %s", *options.EngineType)
	}
	if options.ForceNode != nil && len(*options.ForceNode) > 0 {
		cmd = cmd + fmt.Sprintf(" --force-node %s", *options.ForceNode)
	}
	if len(subCommands) > 0 {
		cmd = cmd + " -- " + strings.Join(subCommands, " ")
		if kataOverrideConfig != nil && len(*kataOverrideConfig) > 0 {
			cmd = cmd + " -override-config " + *kataOverrideConfig
		}
	} else if kataOverrideConfig != nil && len(*kataOverrideConfig) > 0 {
		cmd = cmd + " -- gtp -override-config " + *kataOverrideConfig
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
	sshJSONURL := ""
	if p.Http != nil && p.Http.GetUrl != nil {
		sshJSONURL = *p.Http.GetUrl + "/users/" + client.options.Username + ".ssh.json"
	} else {
		sshJSONURL = "https://" + p.Oss.Bucket + "." + p.Oss.BucketEndpoint + "/users/" + client.options.Username + ".ssh.json"
	}
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
