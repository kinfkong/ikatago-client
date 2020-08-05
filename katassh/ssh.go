package katassh

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kinfkong/ikatago-client/model"
	"golang.org/x/crypto/ssh"
)

func runSSH(sshoptions model.SSHOptions, cmd string) error {
	config := &ssh.ClientConfig{
		Timeout:         30 * time.Second,
		User:            sshoptions.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	config.Auth = []ssh.AuthMethod{ssh.Password(sshoptions.Password)}
	addr := fmt.Sprintf("%s:%d", sshoptions.Host, sshoptions.Port)
	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Fatal("failed to create ssh client", err)
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		log.Fatal("failed to create ssh session", err)
	}

	defer session.Close()
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	log.Printf("DEBUG running equal commad: ssh -p %d %s@%s %s\n", sshoptions.Port, sshoptions.User, sshoptions.Host, cmd)

	err = session.Run(cmd)
	if err != nil {
		return err
	}
	return nil
}
