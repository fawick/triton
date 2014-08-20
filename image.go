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
	Id        int      `json:"id"`
	Name      string   `json:"name"`
	Regions   []string `json:"regions"`
	Public    bool     `json:"public"`
	CreatedAt string   `json:"created_at"`
}

func getImages() ([]Image, error) {
	var list struct {
		Images []Image `json:"images"`
	}
	err := doApiGet(API_URL+"images", &list)
	if err != nil {
		return nil, err
	}
	return list.Images, nil
}

func transferImage(c *cli.Context) {
	setAppOptions(c)
	it := ImageTransfer{"transfer", c.Args().Get(1)}
	imageId, err := resolveImageId(c.Args().Get(0))
	if err != nil {
		fmt.Println(err)
		return
	}
	var resp struct {
		A Action `json:"action"`
	}
	url := API_URL + fmt.Sprintf("images/%d/actions", imageId)
	err = doApiPost(url, it, &resp)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp.A.String())
}

func resolveImageId(imageString string) (int, error) {
	// first try to resolve the ID number directly
	parsedId, err := strconv.ParseInt(imageString, 10, 64)
	if err == nil {
		return int(parsedId), nil
	}
	// didn't work, so assume it's a droplet name
	imageId := -1
	images, err := getImages()
	if err != nil {
		return -1, err
	}
	for _, i := range images {
		if i.Name == imageString {
			imageId = i.Id
			break
		}
	}
	if imageId == -1 {
		return -1, fmt.Errorf("No image %s available", imageString)
	}
	return imageId, nil
}

func deleteImage(c *cli.Context) {
	setAppOptions(c)
	imageId, err := resolveImageId(c.Args().Get(0))
	if err != nil {
		fmt.Println(err)
		return
	}
	err = doApiDelete(API_URL + fmt.Sprintf("images/%d", imageId))
	if err != nil {
		fmt.Println("Error while deleting image:", err)
		return
	}
	fmt.Println("Delete Image with ID", imageId)
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
				Action:    listImages,
			},
			{
				Name:      "transfer",
				ShortName: "t",
				Usage:     "Transfer an Image to another region ",
				Action:    transferImage,
			},
			{
				Name:      "delete",
				ShortName: "d",
				Usage:     "Destroy and delete an image",
				Action:    deleteImage,
			},
		},
	}
}
