package ikatagosdk

import (
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/kinfkong/ikatago-client/client"
	"github.com/kinfkong/ikatago-client/utils"
)

// Client the client wrapper
type Client struct {
	remoteClient *client.Client
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
	engineType         *string
	gpuType            *string
	token              *string
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

// QueryServer queries the server info
func (client *Client) QueryServer() (string, error) {
	buf := bytes.NewBuffer(nil)
	err := client.remoteClient.QueryServer(buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (client *Client) SetEngineType(engineType string) {
	if len(engineType) > 0 {
		client.remoteClient.SetEngineType(&engineType)
	} else {
		client.remoteClient.SetEngineType(nil)
	}
}

// CreateKatagoRunner creates  the katago runner
func (client *Client) CreateKatagoRunner() (*KatagoRunner, error) {
	return &KatagoRunner{
		client:          client,
		refreshInterval: 30,
		transmitMoveNum: 25,
		noCompress:      false,
		useRawData:      false,
		subCommands:     make([]string, 0),
		started:         false,
		engineType:      client.remoteClient.GetEngineType(),
	}, nil
}

// Run runs the katago
func (katagoRunner *KatagoRunner) Run(callback DataCallback) error {
	katagoRunner.started = true
	options := client.RunKatagoOptions{
		NoCompress:         katagoRunner.noCompress,
		RefreshInterval:    katagoRunner.refreshInterval,
		EngineType:         katagoRunner.engineType,
		TransmitMoveNum:    katagoRunner.transmitMoveNum,
		KataLocalConfig:    katagoRunner.kataLocalConfig,
		KataOverrideConfig: katagoRunner.kataOverrideConfig,
		KataName:           katagoRunner.kataName,
		GpuType:            katagoRunner.gpuType,
		Token:              katagoRunner.token,
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

// SetGpuType sets the transmit move num
func (katagoRunner *KatagoRunner) SetGpuType(gpuType string) {
	katagoRunner.gpuType = &gpuType
}

// SetToken sets the token
func (katagoRunner *KatagoRunner) SetToken(token string) {
	katagoRunner.token = &token
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

// SetEngineType sets the engineType. for example: 'katago', 'gomoku'
func (katagoRunner *KatagoRunner) SetEngineType(engineType string) {
	katagoRunner.engineType = &engineType
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
