package ikatagosdk

import (
	"bytes"
	"io"
	"strings"

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
	kataOverrideConfig *string
	kataName           *string
	kataWeight         *string
	kataConfig         *string
	subCommands        []string
	reader             io.Reader
	writer             io.Writer
	stderrWriter       io.Writer
	commandWriter      io.Writer
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

// CreateKatagoRunner creates  the katago runner
func (client *Client) CreateKatagoRunner() (*KatagoRunner, error) {
	return &KatagoRunner{
		client:          client,
		refreshInterval: 30,
		transmitMoveNum: 25,
		noCompress:      false,
		useRawData:      false,
		subCommands:     make([]string, 0),
	}, nil
}

// Run runs the katago
func (katagoRunner *KatagoRunner) Run(callback DataCallback) error {
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

	return katagoRunner.client.remoteClient.RunKatago(options, katagoRunner.subCommands, katagoRunner.reader, katagoRunner.writer, katagoRunner.stderrWriter, callback.OnReady)
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
	katagoRunner.client.remoteClient.StopCurrentSession()
	return nil
}
