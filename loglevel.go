package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/urfave/cli"
)

var (
	DefaultSocketLocation = "/tmp/log.sock"
)

// Client structure is used to store the server info
type Client struct {
	SocketLocation string
	httpc          *http.Client
}

func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "set",
		},
		cli.StringFlag{
			Name:   "socket-location",
			EnvVar: "LOG_SOCKET_LOCATION",
			Value:  DefaultSocketLocation,
		},
	}
	app.Action = run
	app.Run(os.Args)
}

func NewClient(c *cli.Context) *Client {
	client := &Client{
		SocketLocation: c.String("socket-location"),
	}

	client.httpc = &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", client.SocketLocation)
			},
		},
	}

	return client
}

func run(c *cli.Context) error {
	client := NewClient(c)
	if c.String("set") != "" {
		return client.setLogLevel(c.String("set"))
	}

	return client.getLogLevel()
}

func (client *Client) setLogLevel(level string) error {
	if level != "info" && level != "debug" && level != "error" {
		return fmt.Errorf("invalid log level specified")
	}
	response, err := client.httpc.Post("http://unix/v1/loglevel",
		"application/x-www-form-urlencoded",
		strings.NewReader(fmt.Sprintf("level=%v", level)))
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, response.Body)
	return nil
}

func (client *Client) getLogLevel() error {
	response, err := client.httpc.Get("http://unix/v1/loglevel")
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, response.Body)
	return nil
}
