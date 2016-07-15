package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/docker/docker/pkg/archive"
	ct "github.com/flynn/flynn/controller/types"
	"github.com/flynn/flynn/host/image"
	"github.com/flynn/flynn/pinkerton"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) != 2 {
		log.Fatalf("usage: %s NAME", os.Args[0])
	}
	if err := build(os.Args[1]); err != nil {
		log.Fatalln("ERROR:", err)
	}
}

func build(name string) error {
	cmd := exec.Command("docker", "build", "-t", name, ".")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error building docker image: %s", err)
	}

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
	layers := make([]*ct.ImageLayer, 0, len(history))
	for i := len(history) - 1; i >= 0; i-- {
		layer := history[i]
		ids = append(ids, layer.ID)
		if len(layer.Tags) > 0 {
			l, err := b.CreateLayer(ids)
			if err != nil {
				return err
			}
			ids = make([]string, 0, len(history))
			layers = append(layers, l)
		}
	}

	image := &ct.ImageManifest{
		Type: ct.ImageManifestTypeV1,
		Rootfs: []*ct.ImageRootfs{{
			Platform: ct.DefaultImagePlatform,
			Layers:   layers,
		}},
	}

	return json.NewEncoder(os.Stdout).Encode(image)
}

func (b *Builder) CreateLayer(ids []string) (*ct.ImageLayer, error) {
	imageID := ids[len(ids)-1]

	lock, err := os.OpenFile(fmt.Sprintf("/var/lib/flynn/image/tmp/layer-%s.json.lock", imageID), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer lock.Close()
	defer os.Remove(lock.Name())

	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return nil, err
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)

	path := fmt.Sprintf("/var/lib/flynn/image/tmp/layer-%s.json", imageID)
	f, err := os.Open(path)
	if err == nil {
		defer f.Close()
		var layer ct.ImageLayer
		return &layer, json.NewDecoder(f).Decode(&layer)
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	dir, err := ioutil.TempDir("", "docker-layer-")
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

	layer, err := b.repo.CreateLayer(dir)
	if err != nil {
		return nil, err
	}
	f, err = os.Create(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(&layer); err != nil {
		os.Remove(path)
		return nil, err
	}
	return layer, nil
}
