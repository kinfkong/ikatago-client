package katassh

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kinfkong/ikatago-client/model"
	"github.com/kinfkong/ikatago-client/utils"
	"golang.org/x/crypto/ssh"
)

type KataSSHSession struct {
	Stopped bool
	Session *ssh.Session
}

// RunSSH runs the ssh command
func (kataSSHSession *KataSSHSession) RunSSH(sshoptions model.SSHOptions, cmd string, stdinReader io.Reader, stderrWriter io.Writer, outputWriter io.Writer) error {
	config := &ssh.ClientConfig{
		Timeout:         30 * time.Second,
		User:            sshoptions.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	config.Auth = []ssh.AuthMethod{ssh.Password(sshoptions.Password)}
	addr := fmt.Sprintf("%s:%d", sshoptions.Host, sshoptions.Port)
	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	if kataSSHSession.Stopped {
		return nil
	}
	kataSSHSession.Session = session
	defer session.Close()
	session.Stderr = stderrWriter
	session.Stdin = stdinReader
	session.Stdout = outputWriter

	log.Printf("DEBUG running equal commad: ssh -p %d %s@%s %s\n", sshoptions.Port, sshoptions.User, sshoptions.Host, cmd)

	err = session.Run(cmd)
	if err != nil {
		return err
	}
	return nil
}

// RunSCP runs the scp command
func (kataSSHSession *KataSSHSession) RunSCP(sshoptions model.SSHOptions, localFile string, engineType *string) error {
	// check file existence
	if !utils.FileExists(localFile) {
		log.Printf("ERROR config file not found: %s\n", localFile)
		return errors.New("file_not_found")
	}

	basefileName := filepath.Base(localFile)
	if !strings.HasSuffix(basefileName, ".cfg") {
		log.Printf("ERROR config file name must ends with .cfg")
		return errors.New("invalid_file_extension")
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
		return err
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	if kataSSHSession.Stopped {
		return nil
	}
	kataSSHSession.Session = session
	defer session.Close()
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	// session.Stdin = os.Stdin
	log.Printf("DEBUG running scp command: %s %s\n", "scp-config", localFile)
	writer, err := session.StdinPipe()
	if err != nil {
		return err
	}
	f, err := os.Open(localFile)
	if err != nil {
		log.Printf("ERROR cannot open file: %s\n", localFile)
		return err
	}
	_, err = io.Copy(writer, f)
	if err != nil {
		log.Printf("ERROR failed to send file: %s\n", localFile)
		return err
	}
	go func() {
		time.Sleep(3 * time.Second)
		writer.Close()
	}()
	if engineType != nil && len(*engineType) > 0 {
		err = session.Run(fmt.Sprintf("scp-config %s --engine-type %s", basefileName, *engineType))
	} else {
		err = session.Run(fmt.Sprintf("scp-config %s", basefileName))
	}

	if err != nil {
		return err
	}
	return nil
}

// RunKatago runs the ssh as katago
func (kataSSHSession *KataSSHSession) RunKatago(sshoptions model.SSHOptions, cmd string, inputReader io.Reader, outputWriter io.Writer, stderrWriter io.Writer, useRawData bool, onReady func()) error {
	config := &ssh.ClientConfig{
		Timeout:         30 * time.Second,
		User:            sshoptions.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	config.Auth = []ssh.AuthMethod{ssh.Password(sshoptions.Password)}
	addr := fmt.Sprintf("%s:%d", sshoptions.Host, sshoptions.Port)
	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	session, err := sshClient.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	if kataSSHSession.Stopped {
		return nil
	}
	kataSSHSession.Session = session

	session.Stderr = stderrWriter
	session.Stdin = inputReader
	reader, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	go func() {
		buf := make([]byte, 4096)
		var theReader io.Reader = nil
		var gtpReader *GTPReader = nil
		if !useRawData {
			gtpReader = NewGTPReader(reader)
		} else {
			theReader = reader
		}

		for {
			var n int = 0
			var err error = nil
			if !useRawData {
				n, err = gtpReader.Read(buf)
			} else {
				n, err = theReader.Read(buf)
			}

			outputWriter.Write(buf[:n])
			if err != nil {
				if err == io.EOF {
					break
				} else {
					log.Printf("ERROR failed to read from buffer, %+v\n", err)
					return
				}
			}
		}
	}()

	log.Printf("DEBUG running equal commad: ssh -p %d %s@%s %s\n", sshoptions.Port, sshoptions.User, sshoptions.Host, cmd)
	if onReady != nil {
		onReady()
	}
	err = session.Run(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (kataSSHSession *KataSSHSession) Stop() {
	kataSSHSession.Stopped = true
	if kataSSHSession.Session != nil {
		kataSSHSession.Session.Close()
		kataSSHSession.Session = nil
	}
}
