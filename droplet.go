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
	ID       int    `json:"id"`
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
	if len(c.Args()) != 2 {
		fmt.Println("Error with arguments:", c.Args())
		tmpl := cli.CommandHelpTemplate
		cli.CommandHelpTemplate = strings.Replace(cli.CommandHelpTemplate, "[arguments...]", "<Droplet name> <Image>", -1)
		cli.ShowCommandHelp(c, "create")
		cli.CommandHelpTemplate = tmpl
		return
	}
	dc := DropletCreation{Name: c.Args().Get(0), Region: c.String("region"), Size: "512mb", Image: -1}
	var err error
	dc.Image, err = resolveImageID(c.Args().Get(1))
	if err != nil {
		fmt.Println(err)
		return
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
		dc.SSHKeyIds = append(dc.SSHKeyIds, k.ID)
	}
	var resp struct {
		D Droplet `json:"droplet"`
	}
	err = doAPIPost(APIURL+"droplets", dc, &resp)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Created droplet %s in region %s with ID %d\n\n", resp.D.Name, resp.D.Region.Name, resp.D.ID)
}

func doDropletAction(id int, a DropletActionRequest) {
	var resp struct {
		A Action `json:"action"`
	}
	url := APIURL + fmt.Sprintf("droplets/%d/actions", id)
	err := doAPIPost(url, a, &resp)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp.A.String())
}

func simpleDropletActionFunc(actionType string) func(*cli.Context) {
	f := func(c *cli.Context) {
		dropletID, err := resolveDropletID(c.Args().Get(0))
		if err != nil {
			fmt.Println(err)
			return
		}
		a := DropletActionRequest{Type: actionType}
		doDropletAction(dropletID, a)
	}
	return f
}

func createDropletSnapshot(c *cli.Context) {
	if len(c.Args()) != 2 {
		fmt.Println("Need exact two arguments: create <droplet name or id> <image name>")
		return
	}
	dropletID, err := resolveDropletID(c.Args().Get(0))
	if err != nil {
		fmt.Println(err)
		return
	}
	a := DropletActionRequest{Type: "snapshot", Name: c.Args().Get(1)}
	doDropletAction(dropletID, a)
}

func resolveDropletID(dropletString string) (int, error) {
	// first try to resolve the ID number directly
	parsedID, err := strconv.ParseInt(dropletString, 10, 64)
	if err == nil {
		return int(parsedID), nil
	}
	// didn't work, so assume it's a droplet name
	dropletID := -1
	droplets, err := getDroplets()
	if err != nil {
		return -1, err
	}
	for _, d := range droplets {
		if d.Name == dropletString {
			dropletID = d.ID
			break
		}
	}
	if dropletID == -1 {
		return -1, fmt.Errorf("No droplet '%s' available\n\n", dropletString)
	}
	return dropletID, nil
}

func deleteDroplet(c *cli.Context) {
	dropletID, err := resolveDropletID(c.Args().Get(0))
	if err != nil {
		fmt.Println(err)
		return
	}
	err = doAPIDelete(APIURL + fmt.Sprintf("droplets/%d", dropletID))
	if err != nil {
		fmt.Println("Error while deleting droplet:", err)
		return
	}
	fmt.Println("Deleted Droplet with ID", dropletID)
}

func getDroplets() ([]Droplet, error) {
	var list struct {
		Droplets []Droplet `json:"droplets"`
	}
	url := APIURL + "droplets"
	err := doAPIGet(url, &list)
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
				Action:    wrapAction(listDroplets),
			},
			{
				Name:      "create",
				ShortName: "c",
				Usage:     "Create a Droplet from a Image: create <droplet name> <image name or id>",
				Description: "A droplet is created from a specific Image. If a region is specified, it will be " +
					"created in this region. Otherwise the default region will be used. By default, all " +
					"available SSH Keys will be embedded in the image. Use either --no-keys or " +
					"--keys=keyname1[,keyname2[,...]] to change that.",
				Action: wrapAction(createDroplet),
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "region, r",
						Usage: fmt.Sprintf("Which region to create the Droplet in. [%s]", DefaultRegion),
						Value: DefaultRegion,
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
				Action:    wrapAction(deleteDroplet),
			},
			{
				Name:      "poweron",
				ShortName: "p",
				Usage:     "Power on a Droplet",
				Action:    wrapAction(simpleDropletActionFunc("power_on")),
			},
			{
				Name:   "poweroff",
				Usage:  "Power off a Droplet",
				Action: wrapAction(simpleDropletActionFunc("power_off")),
			},
			{
				Name:      "shutdown",
				ShortName: "s",
				Usage:     "Shutdown a Droplet",
				Action:    wrapAction(simpleDropletActionFunc("shutdown")),
			},
			{
				Name:      "reboot",
				ShortName: "r",
				Usage:     "Reboot a Droplet",
				Action:    wrapAction(simpleDropletActionFunc("reboot")),
			},
			{
				Name:   "powercycle",
				Usage:  "Power off and on a Droplet",
				Action: wrapAction(simpleDropletActionFunc("power_cycle")),
			},
			{
				Name:   "passwordreset",
				Usage:  "Reset the root password for a Droplet",
				Action: wrapAction(simpleDropletActionFunc("password_reset")),
			},
			{
				Name:   "ipv6",
				Usage:  "Enable IPv6 for a Droplet",
				Action: wrapAction(simpleDropletActionFunc("enable_ipv6")),
			},
			{
				Name:   "disablebackups",
				Usage:  "Disable backups for a Droplet",
				Action: wrapAction(simpleDropletActionFunc("disable_backups")),
			},
			{
				Name:   "privatenetworking",
				Usage:  "Enable private for a Droplet",
				Action: wrapAction(simpleDropletActionFunc("enable_private_networking")),
			},
			{
				Name:      "snapshot",
				ShortName: "n",
				Usage:     "Create a snapshot image for the Droplet",
				Action:    wrapAction(createDropletSnapshot),
			},
		},
	}
}
