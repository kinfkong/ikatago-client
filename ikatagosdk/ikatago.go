package ikatagosdk

import (
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/kinfkong/ikatago-client/client"
	"github.com/kinfkong/ikatago-client/utils"
)

type genericOptions struct {
	NoCompress         *bool   `long:"no-compress" description:"compress the data during transmission"`
	RefreshInterval    *int    `long:"refresh-interval" description:"sets the refresh interval in cent seconds"`
	TransmitMoveNum    *int    `long:"transmit-move-num" description:"limits number of moves when transmission during analyze"`
	KataLocalConfig    *string `long:"kata-local-config" description:"The katago config file. like, gtp_example.cfg"`
	KataOverrideConfig *string `long:"kata-override-config" description:"The katago override-config, like: analysisPVLen=30,numSearchThreads=30"`
	KataName           *string `long:"kata-name" description:"The katago binary name"`
	KataWeight         *string `long:"kata-weight" description:"The katago weight name"`
	KataConfig         *string `long:"kata-config" description:"The katago config name"`
	EngineType         *string `long:"engine-type" description:"sets the enginetype"`
	ForceNode          *string `long:"force-node" description:"in cluster, force to a specific node."`
	Token              *string `long:"token" description:"sets the token"`
	GpuType            *string `long:"gpu-type" description:"sets the gpu type"`
}

// Client the client wrapper
type Client struct {
	remoteClient *client.Client
	extraArgs    *string
}

// DataCallbackFunc Represents the data callback function
type DataCallbackFunc func(content []byte)

// DataCallback defines the data callback
type DataCallback interface {
	Callback(content []byte)
	StderrCallback(content []byte)
	OnReady()
}

type dataNotifier struct {
	callback DataCallbackFunc
}

// KatagoRunner represents katago runner
type KatagoRunner struct {
	client             *Client
	noCompress         bool
	refreshInterval    int
	transmitMoveNum    int
	useRawData         bool
	kataLocalConfig    *string
	kataOverrideConfig *string
	kataName           *string
	kataWeight         *string
	kataConfig         *string
	subCommands        []string
	reader             io.Reader
	writer             io.Writer
	stderrWriter       io.Writer
	commandWriter      io.Writer
	sessionResult      *client.SessionResult
	started            bool
}

func (notifier *dataNotifier) Write(p []byte) (n int, err error) {
	if notifier.callback != nil {
		notifier.callback(p)
	}
	return len(p), nil
}

// NewClient creates the new mobile client
func NewClient(world string, platform string, username string, password string) (*Client, error) {
	defaultWorld := utils.WorldURL
	if len(world) == 0 {
		world = defaultWorld
	}
	remoteClient, err := client.NewClient(client.Options{
		World:    world,
		Platform: platform,
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, err
	}
	return &Client{
		remoteClient: remoteClient,
	}, nil
}

// SetExtraArgs sets the extra args like "--gpu-type 3x --engine-type katago"
func (client *Client) SetExtraArgs(extraArgs string) {
	opts := genericOptions{}
	_, err := flags.ParseArgs(&opts, strings.Split(extraArgs, " "))
	if err != nil {
		return
	}
	if opts.EngineType != nil {
		client.SetEngineType(*opts.EngineType)
	}
	if opts.GpuType != nil {
		client.SetGpuType(*opts.GpuType)
	}
	if opts.ForceNode != nil {
		client.SetForceNode(*opts.ForceNode)
	}
	if opts.Token != nil {
		client.SetToken(*opts.Token)
	}
	client.extraArgs = &extraArgs
}

// SetToken sets the token
func (client *Client) SetToken(token string) {
	client.remoteClient.Options.Token = &token
}

// SetToken sets the force node
func (client *Client) SetForceNode(forceNode string) {
	client.remoteClient.Options.ForceNode = &forceNode
}

// SetGpuType sets the gpu type
func (client *Client) SetGpuType(gpuType string) {
	client.remoteClient.Options.GpuType = &gpuType
}

// SetEngineType sets the engine type
func (client *Client) SetEngineType(engineType string) {
	client.remoteClient.Options.EngineType = &engineType
}

// QueryServer queries the server info
func (client *Client) QueryServer() (string, error) {
	buf := bytes.NewBuffer(nil)
	err := client.remoteClient.QueryServer(buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// CreateKatagoRunner creates  the katago runner
func (client *Client) CreateKatagoRunner() (*KatagoRunner, error) {
	runner := &KatagoRunner{
		client:          client,
		refreshInterval: 30,
		transmitMoveNum: 25,
		noCompress:      false,
		useRawData:      false,
		subCommands:     make([]string, 0),
		started:         false,
	}
	if client.extraArgs != nil {
		opts := genericOptions{}
		subCommands, err := flags.ParseArgs(&opts, strings.Split(*client.extraArgs, " "))
		if err == nil {
			if len(subCommands) > 0 {
				runner.SetSubCommands(strings.Join(subCommands, " "))
			}
			if opts.KataWeight != nil {
				runner.SetKataWeight(*opts.KataWeight)
			}
			if opts.KataConfig != nil {
				runner.SetKataConfig(*opts.KataConfig)
			}
			if opts.KataName != nil {
				runner.SetKataName(*opts.KataName)
			}
			if opts.KataOverrideConfig != nil {
				runner.SetKataOverrideConfig(*opts.KataOverrideConfig)
			}
			if opts.NoCompress != nil {
				runner.DisableCompress(*opts.NoCompress)
			}
			if opts.RefreshInterval != nil {
				runner.SetRefreshInterval(*opts.RefreshInterval)
			}
			if opts.TransmitMoveNum != nil {
				runner.SetTransmitMoveNum(*opts.TransmitMoveNum)
			}
			if opts.KataLocalConfig != nil {
				runner.SetKataLocalConfig(*opts.KataLocalConfig)
			}
		}

	}
	return runner, nil
}

// Run runs the katago
func (katagoRunner *KatagoRunner) Run(callback DataCallback) error {
	katagoRunner.started = true
	options := client.RunKatagoOptions{
		NoCompress:         katagoRunner.noCompress,
		RefreshInterval:    katagoRunner.refreshInterval,
		TransmitMoveNum:    katagoRunner.transmitMoveNum,
		KataLocalConfig:    katagoRunner.kataLocalConfig,
		KataOverrideConfig: katagoRunner.kataOverrideConfig,
		KataName:           katagoRunner.kataName,
		KataWeight:         katagoRunner.kataWeight,
		KataConfig:         katagoRunner.kataConfig,
		UseRawData:         katagoRunner.useRawData,
	}
	katagoRunner.writer = &dataNotifier{
		callback: callback.Callback,
	}
	katagoRunner.stderrWriter = &dataNotifier{
		callback: callback.StderrCallback,
	}
	pr, pw := io.Pipe()
	katagoRunner.commandWriter = pw
	katagoRunner.reader = pr

	defer pw.Close()
	defer pr.Close()

	sessionResult, err := katagoRunner.client.remoteClient.RunKatago(options, katagoRunner.subCommands, katagoRunner.reader, katagoRunner.writer, katagoRunner.stderrWriter, callback.OnReady)
	if err != nil {
		katagoRunner.started = false
		return err
	}
	katagoRunner.sessionResult = sessionResult
	sessionResult.Wait()
	katagoRunner.started = false
	return nil
}

// SetKataWeight sets the name of the kata weight
func (katagoRunner *KatagoRunner) SetKataWeight(kataWeight string) {
	katagoRunner.kataWeight = &kataWeight
}

// SetKataConfig sets the name of the kata config name
func (katagoRunner *KatagoRunner) SetKataConfig(kataConfig string) {
	katagoRunner.kataConfig = &kataConfig
}

// SetKataLocalConfig sets the name of the kata local config file
func (katagoRunner *KatagoRunner) SetKataLocalConfig(kataLocalConfig string) {
	katagoRunner.kataLocalConfig = &kataLocalConfig
}

// SetKataOverrideConfig sets the name of the kata override-config option of kata, example: analysisPVLen=30,playoutDoublingAdvantage=3
func (katagoRunner *KatagoRunner) SetKataOverrideConfig(kataOverrideConfig string) {
	katagoRunner.kataOverrideConfig = &kataOverrideConfig
}

// SetKataName sets the name of the katago name
func (katagoRunner *KatagoRunner) SetKataName(kataName string) {
	katagoRunner.kataName = &kataName
}

// DisableCompress disables the compression
func (katagoRunner *KatagoRunner) DisableCompress(noCompress bool) {
	katagoRunner.noCompress = noCompress
}

// SetRefreshInterval sets the refresh interval
func (katagoRunner *KatagoRunner) SetRefreshInterval(refreshInterval int) {
	katagoRunner.refreshInterval = refreshInterval
}

// SetTransmitMoveNum sets the transmit move num
func (katagoRunner *KatagoRunner) SetTransmitMoveNum(transmitMoveNum int) {
	katagoRunner.transmitMoveNum = transmitMoveNum
}

// SendGTPCommand sends the gtp command
func (katagoRunner *KatagoRunner) SendGTPCommand(command string) error {
	_, err := io.WriteString(katagoRunner.commandWriter, command)
	if err != nil {
		return err
	}
	return nil
}

// SetUseRawData sets if use the raw data or not
func (katagoRunner *KatagoRunner) SetUseRawData(useRawData bool) {
	katagoRunner.useRawData = useRawData
}

// SetSubCommands sets the subcommands. for example: 'analysis -analysis-threads 12'
func (katagoRunner *KatagoRunner) SetSubCommands(subCommands string) {
	katagoRunner.subCommands = strings.Split(subCommands, " ")
}

// Stop stops the katago engine
func (katagoRunner *KatagoRunner) Stop() error {
	c := 0
	for katagoRunner.started && katagoRunner.sessionResult == nil {
		time.Sleep(100 * time.Millisecond)
		c++
		if c >= 200 {
			// 20 seconds
			break
		}
	}
	if katagoRunner.sessionResult != nil {
		katagoRunner.sessionResult.Stop()
		katagoRunner.sessionResult = nil
		katagoRunner.started = false
	}
	return nil
}
