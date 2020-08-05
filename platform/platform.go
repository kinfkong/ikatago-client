package platform

import (
	"encoding/json"
	"log"

	"github.com/kinfkong/ikatago-client/config"
	"github.com/kinfkong/ikatago-client/model"
	"github.com/kinfkong/ikatago-client/utils"
)

// Oss represents the oss bucket of this platform
type Oss struct {
	BucketEndpoint string `json:"bucketEndpoint"`
	Bucket         string `json:"Bucket"`
}

// Platform represents the platform
type Platform struct {
	Name string `json:"name"`
	Oss  Oss    `json:"oss"`
}

// GetSSHOptions gets the ssh info
func (p *Platform) GetSSHOptions(username string) (*model.SSHOptions, error) {
	sshJSONURL := "https://" + p.Oss.Bucket + "." + p.Oss.BucketEndpoint + "/users/" + username + ".ssh.json"
	response, err := utils.DoHTTPRequest("GET", sshJSONURL, nil, nil)
	if err != nil {
		log.Printf("ERROR error requestting url: %s, err: %+v\n", sshJSONURL, err)
		return nil, err
	}
	sshoptions := model.SSHOptions{}
	// parse json
	err = json.Unmarshal([]byte(response), &sshoptions)
	if err != nil {
		log.Printf("ERROR failed parsing json: %s\n", response)
		return nil, err
	}

	sshoptions.Command = config.GetConfig().GetString("cmd.cmd")
	sshoptions.Password = config.GetConfig().GetString("user.password")
	return &sshoptions, nil
}
