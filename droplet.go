package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
)

type DropletCreation struct {
	Name      string `json:"name"`
	Image     int    `json:"image"`
	Size      string `json:"size"`
	Region    string `json:"region"`
	SSHKeyIds []int  `json:"ssh_keys,omitempty"`
}

type DropletActionRequest struct {
	Type   string `json:"type"`
	Image  int    `json:"image,omitempty"`
	Size   string `json:"size,omitempty"`
	Name   string `json:"name,omitempty"`
	Kernel string `json:"kernel,omitempty"`
}

type Droplet struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Memory   int    `json:"memory"`
	VCPUs    int    `json:"vcpus"`
	Disk     int    `json:"disk"`
	Region   Region `json:"region"`
	Status   string `json:"status"`
	Networks struct {
		IPv4 []struct {
			Address string `json:"ip_address"`
		} `json:"v4"`
		IPv6 []struct {
			Address string `json:"ip_address"`
		} `json:"v6"`
	} `json:"networks"`
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
		fmt.Println(err)
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
		fmt.Println(err)
		return
	}
	// TODO handle no-keys and keys
	for _, k := range keys {
		dc.SSHKeyIds = append(dc.SSHKeyIds, k.Id)
	}
	var resp struct {
		D Droplet `json:"droplet"`
	}
	err = doApiPost(API_URL+"droplets", dc, &resp)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Created droplet %s in region %s with ID %d\n\n", resp.D.Name, resp.D.Region.Name, resp.D.Id)
}

func doDropletAction(id int, a DropletActionRequest) {
	var resp struct {
		A Action `json:"action"`
	}
	url := API_URL + fmt.Sprintf("droplets/%d/actions", id)
	err := doApiPost(url, a, &resp)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp.A.String())
}

func simpleDropletActionFunc(actionType string) func(*cli.Context) {
	f := func(c *cli.Context) {
		setAppOptions(c)
		dropletId, err := resolveDropletId(c.Args().Get(0))
		if err != nil {
			fmt.Println(err)
			return
		}
		a := DropletActionRequest{Type: actionType}
		doDropletAction(dropletId, a)
	}
	return f
}

func createDropletSnapshot(c *cli.Context) {
	setAppOptions(c)
	if len(c.Args()) != 2 {
		fmt.Println("Need exact two arguments: create <droplet name or id> <image name>")
		return
	}
	dropletId, err := resolveDropletId(c.Args().Get(0))
	if err != nil {
		fmt.Println(err)
		return
	}
	a := DropletActionRequest{Type: "snapshot", Name: c.Args().Get(1)}
	doDropletAction(dropletId, a)
}

func deleteDroplet(id int) {
	err := doApiDelete(API_URL + fmt.Sprintf("droplets/%d", id))
	if err != nil {
		fmt.Println("Error while deleting droplet:", err)
		return
	}
	fmt.Println("Deleted Droplet with ID", id)
}

func resolveDropletId(dropletString string) (int, error) {
	// first try to resolve the ID number directly
	parsedId, err := strconv.ParseInt(dropletString, 10, 64)
	if err == nil {
		return int(parsedId), nil
	}
	// didn't work, so assume it's a droplet name
	dropletId := -1
	droplets, err := getDroplets()
	if err != nil {
		return -1, err
	}
	for _, d := range droplets {
		if d.Name == dropletString {
			dropletId = d.Id
			break
		}
	}
	if dropletId == -1 {
		return -1, fmt.Errorf("No droplet '%s' available\n\n", dropletString)
	}
	return dropletId, nil
}

func deleteDropletByName(c *cli.Context) {
	setAppOptions(c)
	dropletId, err := resolveDropletId(c.Args().Get(0))
	if err != nil {
		fmt.Println(err)
		return
	}
	deleteDroplet(dropletId)
}

func getDroplets() ([]Droplet, error) {
	var list struct {
		Droplets []Droplet `json:"droplets"`
	}
	url := API_URL + "droplets"
	err := doApiGet(url, &list)
	if err != nil {
		return nil, err
	}
	return list.Droplets, nil
}

func setupDropletCommands() cli.Command {
	return cli.Command{
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
				Usage:     "Create a Droplet from a Image: create <droplet name> <image name or id>",
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
			{
				Name:      "poweron",
				ShortName: "p",
				Usage:     "Power on a Droplet",
				Action:    simpleDropletActionFunc("power_on"),
			},
			{
				Name:   "poweroff",
				Usage:  "Power off a Droplet",
				Action: simpleDropletActionFunc("power_off"),
			},
			{
				Name:      "shutdown",
				ShortName: "s",
				Usage:     "Shutdown a Droplet",
				Action:    simpleDropletActionFunc("shutdown"),
			},
			{
				Name:      "reboot",
				ShortName: "r",
				Usage:     "Reboot a Droplet",
				Action:    simpleDropletActionFunc("reboot"),
			},
			{
				Name:   "powercycle",
				Usage:  "Power off and on a Droplet",
				Action: simpleDropletActionFunc("power_cycle"),
			},
			{
				Name:   "passwordreset",
				Usage:  "Reset the root password for a Droplet",
				Action: simpleDropletActionFunc("password_reset"),
			},
			{
				Name:   "ipv6",
				Usage:  "Enable IPv6 for a Droplet",
				Action: simpleDropletActionFunc("enable_ipv6"),
			},
			{
				Name:   "disablebackups",
				Usage:  "Disable backups for a Droplet",
				Action: simpleDropletActionFunc("disable_backups"),
			},
			{
				Name:   "privatenetworking",
				Usage:  "Enable private for a Droplet",
				Action: simpleDropletActionFunc("enable_private_networking"),
			},
			{
				Name:      "snapshot",
				ShortName: "n",
				Usage:     "Create a snapshot image for the Droplet",
				Action:    createDropletSnapshot,
			},
		},
	}
}
