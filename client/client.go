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
	ForceNode  *string `json:"forceNode"`
	GpuType    *string `json:"gpuType"`
	Token      *string `json:"token"`
	ExtraInfo  *string `json:"extraInfo"`
}

// RunKatagoOptions represents the run katago options
type RunKatagoOptions struct {
	NoCompress         bool
	RefreshInterval    int
	TransmitMoveNum    int
	KataLocalConfig    *string
	KataName           *string
	KataWeight         *string
	KataConfig         *string
	KataOverrideConfig *string
	ExtraInfo          *string
	UseRawData         bool
	ClientID           *string
}

// Client represents the ikatago client
type Client struct {
	Options    Options
	init       bool
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
		Options: options,
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
		err := (&katassh.KataSSHSession{}).RunSCP(client.sshOptions, *options.KataLocalConfig, client.BuildServerLocationOptions())
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
		err := s.RunKatago(client.sshOptions, client.BuildKatagoCommand("run-katago", options, subCommands), inputReader, outputWriter, stderrWriter, options.UseRawData, onReady)
		if err != nil {
			result.Err = err
		}
		result.wg.Done()
	}()

	return &result, nil
}

// PreloadKatago runs the katago
func (client *Client) PreloadKatago(options RunKatagoOptions, subCommands []string, inputReader io.Reader, outputWriter io.Writer, stderrWriter io.Writer, onReady func()) (*SessionResult, error) {
	if !client.init {
		err := client.initClient()
		if err != nil {
			return nil, err
		}
	}
	if options.KataLocalConfig != nil {
		// run scp to copy the configure
		err := (&katassh.KataSSHSession{}).RunSCP(client.sshOptions, *options.KataLocalConfig, client.BuildServerLocationOptions())
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
		err := s.RunKatago(client.sshOptions, client.BuildKatagoCommand("preload-katago", options, subCommands), inputReader, outputWriter, stderrWriter, options.UseRawData, onReady)
		if err != nil {
			result.Err = err
		}
		result.wg.Done()
	}()

	return &result, nil
}

// ViewConfig views the katago config
func (client *Client) ViewConfig(options RunKatagoOptions, subCommands []string, outputWriter io.Writer) error {
	if !client.init {
		err := client.initClient()
		if err != nil {
			return err
		}
	}
	if options.KataLocalConfig != nil {
		// run scp to copy the configure
		err := (&katassh.KataSSHSession{}).RunSCP(client.sshOptions, *options.KataLocalConfig, client.BuildServerLocationOptions())
		if err != nil {
			return err
		}
	}

	stdinReader, mockWriter := io.Pipe()
	mockReader, stderrWriter := io.Pipe()
	defer mockWriter.Close()
	defer stdinReader.Close()
	defer mockReader.Close()
	defer stderrWriter.Close()

	s := &katassh.KataSSHSession{}

	// build the ssh command
	err := s.RunSSH(client.sshOptions, client.BuildKatagoCommand("view-config", options, subCommands), stdinReader, stderrWriter, outputWriter)
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
	cmd := "query-server"

	// build options with server related options
	serverLocationOptions := client.BuildServerLocationOptions()
	if len(serverLocationOptions) > 0 {
		cmd = cmd + serverLocationOptions
	}

	stdinReader, mockWriter := io.Pipe()
	mockReader, stderrWriter := io.Pipe()
	defer mockWriter.Close()
	defer stdinReader.Close()
	defer mockReader.Close()
	defer stderrWriter.Close()
	err := (&katassh.KataSSHSession{}).RunSSH(client.sshOptions, cmd, stdinReader, stderrWriter, outputWriter)
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) BuildKatagoCommand(cmd string, options RunKatagoOptions, subCommands []string) string {
	kataName := options.KataName
	kataWeight := options.KataWeight
	kataConfig := options.KataConfig
	kataLocalConfig := options.KataLocalConfig
	kataOverrideConfig := options.KataOverrideConfig
	extraInfo := options.ExtraInfo
	clientID := options.ClientID
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
	if extraInfo != nil && len(*extraInfo) > 0 {
		cmd = cmd + fmt.Sprintf(" --extra-info %s", *extraInfo)
	}
	if clientID != nil && len(*clientID) > 0 {
		cmd = cmd + fmt.Sprintf(" --client-id %s", *clientID)
	}
	if !options.NoCompress {
		cmd = cmd + " --compress"
	}
	cmd = cmd + fmt.Sprintf(" --refresh-interval %d", options.RefreshInterval)
	cmd = cmd + fmt.Sprintf(" --transmit-move-num %d", options.TransmitMoveNum)

	// build options with server related options
	serverLocationOptions := client.BuildServerLocationOptions()
	if len(serverLocationOptions) > 0 {
		cmd = cmd + serverLocationOptions
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

func (client *Client) BuildServerLocationOptions() string {
	cmd := ""
	// build options with server related options
	if client.Options.EngineType != nil && len(*client.Options.EngineType) > 0 {
		cmd = cmd + fmt.Sprintf(" --engine-type %s", *client.Options.EngineType)
	}
	if client.Options.ForceNode != nil && len(*client.Options.ForceNode) > 0 {
		cmd = cmd + fmt.Sprintf(" --force-node %s", *client.Options.ForceNode)
	}
	if client.Options.GpuType != nil && len(*client.Options.GpuType) > 0 {
		cmd = cmd + fmt.Sprintf(" --gpu-type %s", *client.Options.GpuType)
	}
	if client.Options.Token != nil && len(*client.Options.Token) > 0 {
		cmd = cmd + fmt.Sprintf(" --token %s", *client.Options.Token)
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
	worldJSONString, err := utils.DoHTTPRequest("GET", client.Options.World, nil, nil)
	if err != nil {
		return nil, err
	}
	world := &World{}
	err = json.Unmarshal([]byte(worldJSONString), &world)
	if err != nil {
		return nil, err
	}
	for _, platform := range world.Platforms {
		if platform.Name == client.Options.Platform {
			return &platform, nil
		}
	}
	log.Printf("ERROR platform not found in the world. platform: %s", client.Options.Platform)
	return nil, errors.New("platform_not_found")
}

// getSSHOptions gets the ssh info
func (client *Client) getSSHOptions(p *platform.Platform) (*model.SSHOptions, error) {
	sshJSONURL := ""
	if p.Http != nil && p.Http.GetUrl != nil {
		sshJSONURL = *p.Http.GetUrl + "/users/" + client.Options.Username + ".ssh.json"
	} else {
		sshJSONURL = "https://" + p.Oss.Bucket + "." + p.Oss.BucketEndpoint + "/users/" + client.Options.Username + ".ssh.json"
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
	sshoptions.Password = client.Options.Password
	return &sshoptions, nil
}
