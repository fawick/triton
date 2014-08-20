package main

import (
	"fmt"

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
	err = doApiPost(url, it, &resp)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp.A.String())
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
		},
	}
}
