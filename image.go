package main

import (
	"fmt"
	"strconv"

	"github.com/codegangsta/cli"
)

type ImageTransfer struct {
	Type   string `json:"type"`
	Region string `json:"region"`
}

type Image struct {
	ID        int      `json:"id"`
	Name      string   `json:"name"`
	Regions   []string `json:"regions"`
	Public    bool     `json:"public"`
	CreatedAt string   `json:"created_at"`
}

func getImages() ([]Image, error) {
	var list struct {
		Images []Image `json:"images"`
	}
	err := doAPIGet(APIURL+"images", &list)
	if err != nil {
		return nil, err
	}
	return list.Images, nil
}

func transferImage(c *cli.Context) {
	it := ImageTransfer{"transfer", c.Args().Get(1)}
	imageID, err := resolveImageID(c.Args().Get(0))
	if err != nil {
		fmt.Println(err)
		return
	}
	var resp struct {
		A Action `json:"action"`
	}
	url := APIURL + fmt.Sprintf("images/%d/actions", imageID)
	err = doAPIPost(url, it, &resp)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp.A.String())
}

func resolveImageID(imageString string) (int, error) {
	// first try to resolve the ID number directly
	parsedID, err := strconv.ParseInt(imageString, 10, 64)
	if err == nil {
		return int(parsedID), nil
	}
	// didn't work, so assume it's a droplet name
	imageID := -1
	images, err := getImages()
	if err != nil {
		return -1, err
	}
	for _, i := range images {
		if i.Name == imageString {
			imageID = i.ID
			break
		}
	}
	if imageID == -1 {
		return -1, fmt.Errorf("No image %s available", imageString)
	}
	return imageID, nil
}

func deleteImage(c *cli.Context) {
	imageID, err := resolveImageID(c.Args().Get(0))
	if err != nil {
		fmt.Println(err)
		return
	}
	err = doAPIDelete(APIURL + fmt.Sprintf("images/%d", imageID))
	if err != nil {
		fmt.Println("Error while deleting image:", err)
		return
	}
	fmt.Println("Delete Image with ID", imageID)
}

func setupImageCommands() cli.Command {
	return cli.Command{
		Name:      "image",
		ShortName: "i",
		Usage:     "Perform image actions such as transfer.",
		Subcommands: []cli.Command{
			{
				Name:      "list",
				ShortName: "l",
				Usage:     "An alias for list images ",
				Action:    wrapAction(listImages),
			},
			{
				Name:      "transfer",
				ShortName: "t",
				Usage:     "Transfer an Image to another region ",
				Action:    wrapAction(transferImage),
			},
			{
				Name:      "delete",
				ShortName: "d",
				Usage:     "Destroy and delete an image",
				Action:    wrapAction(deleteImage),
			},
		},
	}
}
