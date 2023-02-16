package main

import (
	"context"
	"errors"
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
	VERSION               = "dev"
)

// Client structure is used to store the server info
type Client struct {
	SocketLocation string
	httpc          *http.Client
}

func main() {
	app := cli.NewApp()
	app.Version = VERSION
	app.Usage = "Dynamically change loglevel"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "set",
			Usage: "Set loglevel",
		},
		cli.StringFlag{
			Name:   "socket-location",
			EnvVar: "LOG_SOCKET_LOCATION",
			Value:  DefaultSocketLocation,
		},
	}
	app.Action = run
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
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
		err := client.setLogLevel(c.String("set"))
		if err != nil {
			return fmt.Errorf("Error when setting loglevel: %s", err)
		}
		return nil
	}
	err := client.getLogLevel()
	if err != nil {
		return fmt.Errorf("Error when getting loglevel: %s", err)
	}
	return nil
}

func (client *Client) setLogLevel(level string) error {
	response, err := client.httpc.Post("http://unix/v1/loglevel",
		"application/x-www-form-urlencoded",
		strings.NewReader(fmt.Sprintf("level=%v", level)))
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}
		return errors.New(strings.TrimRight(string(body), "\n"))
	}
	_, err = io.Copy(os.Stdout, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func (client *Client) getLogLevel() error {
	response, err := client.httpc.Get("http://unix/v1/loglevel")
	if err != nil {
		return err
	}
	_, err = io.Copy(os.Stdout, response.Body)
	if err != nil {
		return err
	}

	return nil
}
