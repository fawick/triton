package main

import (
	"fmt"
	"time"

	"github.com/codegangsta/cli"
)

func setupListCommands() cli.Command {
	return cli.Command{
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
	}
}

func listSSHKeys(c *cli.Context) {
	setAppOptions(c)
	keys, err := getSSHKeys()
	if err != nil {
		fmt.Println(err)
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

func listImages(c *cli.Context) {
	setAppOptions(c)
	images, err := getImages()
	if err != nil {
		fmt.Println(err)
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
