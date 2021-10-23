package model

import "golang.org/x/crypto/ssh"

// SSHOptions represents the ssh options
type SSHOptions struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	User string `json:"user"`

	Password string `json:"password"`
}

type SSHState struct {
	Session *ssh.Session
	Stopped bool
}
