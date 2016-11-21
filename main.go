package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/codegangsta/cli"
)

// TODO set up config file for defaults and TOKEN
var options struct {
	Token   string
	Debug   bool
	Verbose bool
}

func setAppOptions(c *cli.Context) {
	options.Token = c.GlobalString("token")
	options.Debug = c.GlobalBool("debug")
	options.Verbose = c.GlobalBool("verbose")
}

type Action struct {
	ID           int    `json:"id"`
	Status       string `json:"status"`
	Type         string `json:"type"`
	ResourceID   int    `json:"resource_id"`
	ResourceType string `json:"resource_type"`
	Region       string `json:"region"`
}

func (a Action) String() string {
	return fmt.Sprintf("%s for %s %d (Region %s): %s\n",
		a.Type, a.ResourceType, a.ResourceID, a.Region, a.Status)
}

type Region struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type SSHKey struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

const (
	DefaultRegion = "ams1"
	APIURL        = "https://api.digitalocean.com/v2/"
)

func newAPIRequest(method, url string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return request, err
	}
	request.Header.Add("Authorization", "Bearer "+options.Token)
	return request, nil
}

func doAPIGet(url string, respData interface{}) error {
	return doAPIRequest("GET", url, nil, respData)
}

func doAPIPost(url string, reqData, respData interface{}) error {
	return doAPIRequest("POST", url, reqData, respData)
}

func doAPIDelete(url string) error {
	return doAPIRequest("DELETE", url, nil, nil)
}

func doAPIRequest(method, url string, reqData, respData interface{}) error {
	b := new(bytes.Buffer)
	request, err := newAPIRequest(method, url, b)
	if err != nil {
		return err
	}
	if reqData != nil {
		request.Header.Add("Content-Type", "application/json")
		e := json.NewEncoder(b)
		err := e.Encode(reqData)
		if err != nil {
			return err
		}
	}
	response, err := http.DefaultClient.Do(request)

	if err != nil {
		return err
	}
	defer response.Body.Close()

	if options.Debug {
		b, _ := httputil.DumpRequestOut(request, true)
		fmt.Fprintf(os.Stderr, string(b))
		b, _ = httputil.DumpResponse(response, true)
		fmt.Fprintf(os.Stderr, string(b))
	}

	if options.Verbose {
		fmt.Println(method, url, "-", response.Status)
	}
	if response.StatusCode == 422 {
		b, _ := json.Marshal(reqData)
		return fmt.Errorf("Unprocessable Entity: %s", string(b))
	} else if response.StatusCode > 400 {
		return fmt.Errorf("%s", response.Status)
	}
	if respData != nil {
		d := json.NewDecoder(response.Body)
		err := d.Decode(respData)
		if err != nil {
			return err
		}
	}

	return nil
}

func getSSHKeys() ([]SSHKey, error) {
	var list struct {
		SSHKeys []SSHKey `json:"ssh_keys"`
	}
	err := doAPIGet(APIURL+"account/keys", &list)
	if err != nil {
		return nil, err
	}
	return list.SSHKeys, nil
}

func SSHKeyByName(name string) (int, error) {
	keys, err := getSSHKeys()
	if err != nil {
		return -1, err
	}
	for _, k := range keys {
		if k.Name == name {
			return k.ID, nil
		}
	}
	return -1, fmt.Errorf("No key with name %s available", name)
}

func wrapAction(f func(*cli.Context)) func(*cli.Context) {
	return func(c *cli.Context) {
		setAppOptions(c)
		f(c)
	}
}

func main() {
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Name = "triton"
	app.Author = "Fabian Wickborn"
	app.Version = "Without supercow powers"
	app.Usage = "The messenger for the DigitalOcean"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "Print more output",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Dump HTTP requests and responses to stderr",
		},
		cli.StringFlag{
			Name:   "token, t",
			Usage:  "The DigitalOcean API v2 Access Token",
			EnvVar: "DIGITALOCEAN_API_TOKEN",
		},
	}
	app.Commands = []cli.Command{
		setupListCommands(),
		setupDropletCommands(),
		setupImageCommands(),
	}
	app.Run(os.Args)
}
