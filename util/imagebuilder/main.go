package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/docker/docker/pkg/archive"
	"github.com/flynn/flynn/host/image"
	"github.com/flynn/flynn/pinkerton"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) != 2 {
		log.Fatalf("usage: %s IMAGE", os.Args[0])
	}
	if err := build(os.Args[1]); err != nil {
		log.Fatalln("ERROR:", err)
	}
}

func build(name string) error {
	context, err := pinkerton.BuildContext("aufs", "/var/lib/docker")
	if err != nil {
		return err
	}

	repo, err := image.NewRepository("/var/lib/flynn/image")
	if err != nil {
		return err
	}

	builder := &Builder{
		context: context,
		repo:    repo,
	}

	return builder.Build(name)
}

type Builder struct {
	context *pinkerton.Context
	repo    *image.Repository
}

func (b *Builder) Build(name string) error {
	history, err := b.context.History(name)
	if err != nil {
		return err
	}

	ids := make([]string, 0, len(history))
	var parent *image.Image
	for i := len(history) - 1; i >= 0; i-- {
		layer := history[i]
		ids = append(ids, layer.ID)
		if len(layer.Tags) > 0 {
			image, err := b.CreateImage(ids, parent)
			if err != nil {
				return err
			}
			ids = make([]string, 0, len(history))
			parent = image
		}
	}

	return nil
}

func (b *Builder) CreateImage(ids []string, parent *image.Image) (*image.Image, error) {
	imageID := ids[len(ids)-1]

	if image, _ := b.repo.Lookup(imageID); image != nil {
		return image, nil
	}

	dir, err := ioutil.TempDir("", "flynn-image-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	for i, id := range ids {
		parent := ""
		if i > 0 {
			parent = ids[i-1]
		}
		diff, err := b.context.Diff(id, parent)
		if err != nil {
			return nil, err
		}
		if err := archive.Untar(diff, dir, &archive.TarOptions{}); err != nil {
			return nil, err
		}
	}

	return b.repo.CreateImage(dir, imageID, parent)
}
