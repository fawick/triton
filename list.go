package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/codegangsta/cli"
)

type TableWriter struct {
	*tabwriter.Writer
	format string
}

func (t *TableWriter) Header(a ...interface{}) {
	var h, d []interface{}
	for _, e := range a {
		s := fmt.Sprintf("%v", e)
		t.format += "%v\t"
		h = append(h, strings.ToUpper(s))
		d = append(d, "")
	}
	t.format += "\n"
	fmt.Fprintf(t, t.format, h...)
	fmt.Fprintf(t, t.format, d...)
}

func (t TableWriter) Line(a ...interface{}) {
	fmt.Fprintf(t, t.format, a...)
}

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
				Action:    wrapAction(listDroplets),
			},
			{
				Name:        "images",
				ShortName:   "i",
				Usage:       "List images",
				Description: "List available system images. By default only private images are shown. Use -a to show all images.",
				Action:      wrapAction(listImages),
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
				Action:    wrapAction(listSSHKeys),
			},
		},
	}
}

func listSSHKeys(c *cli.Context) {
	keys, err := getSSHKeys()
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(keys) == 0 {
		fmt.Print("\nNo Keys available\n\n")
		return
	}
	fmt.Print("\nAvailable SSH Keys\n\n")
	tab.Header("ID", "Name")
	for _, k := range keys {
		tab.Line(k.ID, k.Name)
	}
	tab.Flush()
	fmt.Println()
}

func listImages(c *cli.Context) {
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
		fmt.Print("\nNo Images available\n\n")
		return
	}
	fmt.Print("\nAvailable Images\n\n")
	tab.Header("ID", "Name", "Creation", "Regions")
	for _, i := range images {
		t, _ := time.Parse(time.RFC3339, i.CreatedAt)
		s := t.Format(time.RFC822)
		tab.Line(i.ID, i.Name, s, i.Regions)
	}
	tab.Flush()
	fmt.Println()
}

var tab = TableWriter{tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0), ""}

func listDroplets(c *cli.Context) {
	droplets, err := getDroplets()
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(droplets) == 0 {
		fmt.Print("\nNo Droplets available\n\n")
		return
	}
	fmt.Print("\nAvailable Droplets\n\n")
	tab.Header("ID", "Name", "Region", "Status", "IP Address")
	for _, d := range droplets {
		tab.Line(d.ID, d.Name, d.Region.Name, d.Status,
			d.Networks.IPv4[0].Address)
	}
	tab.Flush()
	fmt.Println()
}
