package katassh

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bramvdbogaerde/go-scp"
	"github.com/kinfkong/ikatago-client/model"
	"github.com/kinfkong/ikatago-client/utils"
	"golang.org/x/crypto/ssh"
)

// RunStandardSCP runs the scp command
func RunStandardSCP(sshoptions model.SSHOptions, cmd string, localFile string) error {
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

	// Create a new SCP client
	client := scp.NewClient(addr, config)

	// Connect to the remote server
	err = client.Connect()
	if err != nil {
		log.Printf("cannot connect to ikatago-server.")
		return err
	}

	f, err := os.Open(localFile)
	if err != nil {
		log.Printf("ERROR cannot open file: %s\n", localFile)
		return err
	}
	defer f.Close()
	err = client.CopyFile(f, "test.txt", "0655")

	if err != nil {
		log.Println("Error while copying file ", err)
		return err
	}
	return nil
}
