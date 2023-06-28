// Modified only to change the DefaultSocketLocation.

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/urfave/cli"
)

var (
	DefaultSocketLocation = "\x00logserver" // \x00 is the null byte, which we need to use for abstract namespace sockets instead of @ for some reason
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
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}
		return errors.New(strings.TrimRight(string(body), "\n"))
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
