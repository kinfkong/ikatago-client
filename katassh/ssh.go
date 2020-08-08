package katassh

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/kinfkong/ikatago-client/model"
	"github.com/kinfkong/ikatago-client/utils"
	"golang.org/x/crypto/ssh"
)

// RunSSH runs the ssh command
func RunSSH(sshoptions model.SSHOptions, cmd string) error {
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

// RunSCP runs the scp command
func RunSCP(sshoptions model.SSHOptions, cmd string, localFile string) error {
	// check file existence
	if !utils.FileExists(localFile) {
		log.Printf("ERROR config file not found: %s\n", localFile)
		return errors.New("file_not_found")
	}
	fileSize, err := utils.GetFileSize(localFile)
	if err != nil {
		log.Printf("ERROR cannot get file size: %s\n", localFile)
		return errors.New("io_error")
	}
	if fileSize >= 1024*100 {
		log.Printf("ERROR config file: %s is too large: %v\n", localFile, fileSize)
		return errors.New("file_too_large")
	}
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
	// session.Stdin = os.Stdin
	// log.Printf("DEBUG running equal commad: ssh -p %d %s@%s %s\n", sshoptions.Port, sshoptions.User, sshoptions.Host, cmd)
	log.Printf("DEBUG running scp command: %s %s\n", cmd, localFile)
	writer, err := session.StdinPipe()
	if err != nil {
		return err
	}
	io.WriteString(writer, "Hello, I am testing")
	writer.Close()
	err = session.Run(cmd)
	if err != nil {
		return err
	}
	return nil
}
