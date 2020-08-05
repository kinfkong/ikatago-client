package model

// SSHOptions represents the ssh options
type SSHOptions struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	User string `json:"user"`

	Password string `json:"password"`
	Command  string `json:"command"`
}
