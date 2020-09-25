package client

import "io"

// MultiClient represents the ikatago client
type MultiClient struct {
	Clients []Client
}

// NewMultiClient creates the multi client
func NewMultiClient(clients []Client) (*MultiClient, error) {
	return &MultiClient{
		Clients: clients,
	}, nil
}

// RunKatago runs the katago
func (client *MultiClient) RunKatago(options RunKatagoOptions, subCommands []string, inputReader io.Reader, outputWriter io.Writer, stderrWriter io.Writer) error {
	for _, subclient := range client.Clients {
		err := subclient.RunKatago(options, subCommands, inputReader, outputWriter, stderrWriter)
		if err != nil {
			return err
		}
	}
	return nil
}

// QueryServer queries the server
func (client *MultiClient) QueryServer(outputWriter io.Writer) error {
	return client.Clients[0].QueryServer(outputWriter)
}
