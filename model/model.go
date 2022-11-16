package model

// SSHOptions represents the ssh options
type SSHOptions struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	User string `json:"user"`

	Password string `json:"password"`
}
type AllOpts struct {
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
	ExtraInfo  *string `long:"extra-info" description:"The extra info"`
	Command    string  `long:"cmd" description:"The command to run the katago" default:"run-katago"`
}
