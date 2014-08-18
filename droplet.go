package main

import (
	"fmt"
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
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Memory int    `json:"memory"`
	VCPUs  int    `json:"vcpus"`
	Disk   int    `json:"disk"`
	Region Region `json:"region"`
	Status string `json:"status"`
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
	url := API_URL + fmt.Sprintf("droplets/%d/actions")
	err := doApiPost(url, a, &resp)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp.A.String())
}

func deleteDroplet(id int) {
	err := doApiDelete(API_URL + fmt.Sprintf("droplets/%d", id))
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

func listDroplets(c *cli.Context) {
	setAppOptions(c)
	droplets, err := getDroplets()
	if err != nil {
		fmt.Println(err)
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
