package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

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
	Id           int    `json:"id"`
	Status       string `json:"status"`
	Type         string `json:"type"`
	ResourceID   int    `json:"resource_id"`
	ResourceType string `json:"resource_type"`
	Region       string `json:"region"`
}

func (a Action) String() string {
	return fmt.Sprintf("Started %s for %s %d (Region %s): %s\n",
		a.Type, a.ResourceType, a.ResourceID, a.Region, a.Status)
}

type DropletCreation struct {
	Name      string `json:"name"`
	Image     int    `json:"image"`
	Size      string `json:"size"`
	Region    string `json:"region"`
	SSHKeyIds []int  `json:"ssh_keys,omitempty"`
}

type Region struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type Droplet struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Memory int    `json:"memory"`
	VCPUs  int    `json:"vcpus"`
	Disk   int    `json:"disk"`
	Region Region `json:"region"`
	Status string `json:"status"`
}

type SSHKey struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type ImageTransfer struct {
	Type   string `json:"type"`
	Region string `json:"region"`
}

type Image struct {
	Id        int      `json:"id"`
	Name      string   `json:"name"`
	Regions   []string `json:"regions"`
	Public    bool     `json:"public"`
	CreatedAt string   `json:"created_at"`
}

const (
	DEFAULT_REGION = "ams1"
	API_URL        = "https://api.digitalocean.com/v2/"
)

func newApiRequest(method, url string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return request, err
	}
	request.Header.Add("Authorization", "Bearer "+options.Token)
	return request, nil
}

func doApiRequest(method, url string, reqData, respData interface{}) error {
	b := new(bytes.Buffer)
	request, err := newApiRequest(method, url, b)
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

	fmt.Println(response.Status)
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

func createDroplet(c *cli.Context) {
	setAppOptions(c)
	if len(c.Args()) != 2 {
		fmt.Println("Error with arguments:", c.Args(), "\n")
		tmpl := cli.CommandHelpTemplate
		cli.CommandHelpTemplate = strings.Replace(cli.CommandHelpTemplate, "[arguments...]", "<Droplet name> <Image>", -1)
		cli.ShowCommandHelp(c, "create")
		cli.CommandHelpTemplate = tmpl
		return
	}
	dc := DropletCreation{Name: c.Args().Get(0), Region: c.String("region"), Size: "512mb", Image: -1}
	images, err := getImages()
	if err != nil {
		fmt.Print(err)
		return
	}
	for _, i := range images {
		if i.Name == c.Args().Get(1) {
			dc.Image = i.Id
		}
	}
	if dc.Image == -1 {
		fmt.Printf("Cannot create droplet: No image '%s' available\n\n", c.Args().Get(1))
		return
	}
	keys, err := getSSHKeys()
	if err != nil {
		fmt.Print(err)
		return
	}
	// TODO handle no-keys and keys
	for _, k := range keys {
		dc.SSHKeyIds = append(dc.SSHKeyIds, k.Id)
	}
	var resp struct {
		D Droplet `json:"droplet"`
	}
	err = doApiRequest("POST", API_URL+"droplets", dc, &resp)
	if err != nil {
		fmt.Print(err)
		return
	}
	fmt.Printf("Created droplet %s in region %s with ID %d\n\n", resp.D.Name, resp.D.Region.Name, resp.D.Id)
}

func deleteDroplet(id int) {
	err := doApiRequest("DELETE", API_URL+fmt.Sprintf("droplets/%d", id), nil, nil)
	if err != nil {
		fmt.Println("Error while deleting droplet:", err)
		return
	}
	fmt.Println("Deleted Droplet with ID", id)
}

func deleteDropletByName(c *cli.Context) {
	setAppOptions(c)
	dropletName := c.Args().Get(0)
	dropletId := -1
	droplets, err := getDroplets()
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, d := range droplets {
		if d.Name == dropletName {
			dropletId = d.Id
			break
		}
	}
	if dropletId == -1 {
		fmt.Printf("Cannot delete drople: No droplet '%s' available\n\n", dropletName)
		return
	}
	deleteDroplet(dropletId)
}

func transferImage(c *cli.Context) {
	setAppOptions(c)
	it := ImageTransfer{"transfer", c.Args().Get(1)}
	imageId := -1
	images, err := getImages()
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, i := range images {
		if i.Name == c.Args().Get(0) {
			imageId = i.Id
			break
		}
	}
	if imageId == -1 {
		fmt.Printf("Cannot transfer image: No image '%s' available\n\n", c.Args().Get(0))
		return
	}
	var resp struct {
		A Action `json:"action"`
	}
	url := API_URL + fmt.Sprintf("images/%d/actions", imageId)
	err = doApiRequest("POST", url, it, &resp)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp.A.String())
}

func getDroplets() ([]Droplet, error) {
	var list struct {
		Droplets []Droplet `json:"droplets"`
	}
	url := API_URL + "droplets"
	err := doApiRequest("GET", url, nil, &list)
	if err != nil {
		return nil, err
	}
	return list.Droplets, nil
}

func getImages() ([]Image, error) {
	var list struct {
		Images []Image `json:"images"`
	}
	url := API_URL + "images"
	err := doApiRequest("GET", url, nil, &list)
	if err != nil {
		return nil, err
	}
	return list.Images, nil
}

func getSSHKeys() ([]SSHKey, error) {
	var list struct {
		SSHKeys []SSHKey `json:"ssh_keys"`
	}
	url := API_URL + "account/keys"
	err := doApiRequest("GET", url, nil, &list)
	if err != nil {
		return nil, err
	}
	return list.SSHKeys, nil
}

func listDroplets(c *cli.Context) {
	setAppOptions(c)
	droplets, err := getDroplets()
	if err != nil {
		fmt.Print(err)
		return
	}
	if len(droplets) == 0 {
		fmt.Println("\nNo Droplets available\n")
		return
	}
	fmt.Println("\nAvailable Droplets\n")
	fmt.Printf("%-10s  %-20s  %-20s  %s\n", "ID", "NAME", "REGION", "STATUS")
	for _, d := range droplets {
		fmt.Printf("%-10d  %-20s  %-20s  %s\n", d.Id, d.Name, d.Region.Name, d.Status)
	}
	fmt.Println()
}

func ImageByName(name string) (int, error) {
	images, err := getImages()
	if err != nil {
		return -1, err
	}
	for _, i := range images {
		if i.Name == name {
			return i.Id, nil
		}
	}
	return -1, fmt.Errorf("No image with name %s available", name)
}

func listImages(c *cli.Context) {
	setAppOptions(c)
	images, err := getImages()
	if err != nil {
		fmt.Print(err)
		return
	}
	if !c.Bool("all") {
		var l []Image
		for idx := range images {
			if !images[idx].Public {
				l = append(l, images[idx])
			}
		}
		images = l
	}
	if len(images) == 0 {
		fmt.Println("\nNo Images available\n")
		return
	}
	fmt.Println("\nAvailable Images\n")
	fmt.Printf("%-10s  %-25s  %-20s  %s\n\n", "ID", "NAME", "CREATION", "REGIONS")
	for _, i := range images {
		t, _ := time.Parse(time.RFC3339, i.CreatedAt)
		s := t.Format(time.RFC822)
		fmt.Printf("%-10d  %-25s  %-20s  %s\n", i.Id, i.Name, s, i.Regions)
	}
	fmt.Println()
}

func SSHKeyByName(name string) (int, error) {
	keys, err := getSSHKeys()
	if err != nil {
		return -1, err
	}
	for _, k := range keys {
		if k.Name == name {
			return k.Id, nil
		}
	}
	return -1, fmt.Errorf("No key with name %s available", name)
}

func listSSHKeys(c *cli.Context) {
	setAppOptions(c)
	keys, err := getSSHKeys()
	if err != nil {
		fmt.Print(err)
		return
	}
	if len(keys) == 0 {
		fmt.Println("\nNo Keys available\n")
		return
	}
	fmt.Println("\nAvailable SSH Keys\n\nID\tNAME\n")
	for _, k := range keys {
		fmt.Printf("%d\t%s\n", k.Id, k.Name)
	}
	fmt.Println()
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
		{
			Name:      "list",
			ShortName: "l",
			Usage:     "List droplets, images, ssh keys",
			Subcommands: []cli.Command{
				{
					Name:      "droplets",
					ShortName: "d",
					Usage:     "List all droplets",
					Action:    listDroplets,
				},
				{
					Name:        "images",
					ShortName:   "i",
					Usage:       "List images",
					Description: "List available system images. By default only private images are shown. Use -a to show all images.",
					Action:      listImages,
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "all, a",
							Usage: "Print all images (public and private)",
						},
					},
				},
				{
					Name:      "keys",
					ShortName: "k",
					Usage:     "List all keys",
					Action:    listSSHKeys,
				},
			},
		},
		{
			Name:      "droplet",
			ShortName: "d",
			Usage:     "Create, modify or destroy applets.",
			Subcommands: []cli.Command{
				{
					Name:      "list",
					ShortName: "l",
					Usage:     "An alias for list droplets ",
					Action:    listDroplets,
				},
				{
					Name:      "create",
					ShortName: "c",
					Usage:     "Create a Droplet from a Image",
					Description: "A droplet is created from a specific Image. If a region is specified, it will be " +
						"created in this region. Otherwise the default region will be used. By default, all " +
						"available SSH Keys will be embedded in the image. Use either --no-keys or " +
						"--keys=keyname1[,keyname2[,...]] to change that.",
					Action: createDroplet,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "region, r",
							Usage: fmt.Sprintf("Which region to create the Droplet in. [%s]", DEFAULT_REGION),
							Value: DEFAULT_REGION,
						},
						cli.BoolFlag{
							Name:  "no-keys",
							Usage: "Do not set up any SSH keys in the Droplet",
						},
						cli.StringSliceFlag{
							Name:  "keys",
							Value: &cli.StringSlice{},
							Usage: "Comma-separated lists of SSH Key names to set up (cf. 'list keys')",
						},
					},
				},
				{
					Name:      "delete",
					ShortName: "d",
					Usage:     "Destroy and delete a Droplet",
					Action:    deleteDropletByName,
				},
			},
		},
		{
			Name:      "image",
			ShortName: "i",
			Usage:     "Perform image actions such as transfer.",
			Subcommands: []cli.Command{
				{
					Name:      "list",
					ShortName: "l",
					Usage:     "An alias for list images ",
					Action:    listImages,
				},
				{
					Name:      "transfer",
					ShortName: "t",
					Usage:     "Transfer an Image to another region ",
					Action:    transferImage,
				},
			},
		},
	}
	app.Run(os.Args)
}
