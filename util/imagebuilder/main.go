package main

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/docker/docker/pkg/archive"
	ct "github.com/flynn/flynn/controller/types"
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

	root := "/var/lib/flynn/image"
	layerDir := filepath.Join(root, "layers")
	tmpDir := filepath.Join(root, "tmp")
	for _, dir := range []string{root, layerDir, tmpDir} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}

	builder := &Builder{
		context:  context,
		layerDir: layerDir,
		tmpDir:   tmpDir,
	}

	return builder.Build(name)
}

type Builder struct {
	context  *pinkerton.Context
	layerDir string
	tmpDir   string
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

// CreateLayer creates a squashfs layer from a docker layer ID chain by
// creating a temporary directory, applying the relevant diffs then calling
// mksquashfs.
//
// Each squashfs layer is serialized as JSON and cached in a temporary file to
// avoid regenerating existing layers, with access wrapped with a lock file in
// case multiple images are being built at the same time.
func (b *Builder) CreateLayer(ids []string) (*ct.ImageLayer, error) {
	imageID := ids[len(ids)-1]
	layerCache := filepath.Join(b.tmpDir, fmt.Sprintf("layer-%s.json", imageID))

	// acquire the lock file using flock(2) to synchronize access to the
	// layer cache
	lockPath := layerCache + ".lock"
	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer os.Remove(lock.Name())
	defer lock.Close()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return nil, err
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)

	// if the layer cache exists, deserialize and return
	f, err := os.Open(layerCache)
	if err == nil {
		defer f.Close()
		var layer ct.ImageLayer
		return &layer, json.NewDecoder(f).Decode(&layer)
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	// apply the docker layer diffs to a temporary directory
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

	// create the squashfs layer
	layer, err := b.mksquashfs(dir)
	if err != nil {
		return nil, err
	}

	// write the serialized layer to the cache file
	f, err = os.Create(layerCache)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(&layer); err != nil {
		os.Remove(layerCache)
		return nil, err
	}
	return layer, nil
}

func (b *Builder) mksquashfs(dir string) (*ct.ImageLayer, error) {
	tmp, err := ioutil.TempFile(b.tmpDir, "squashfs-")
	if err != nil {
		return nil, err
	}
	defer tmp.Close()

	if out, err := exec.Command("mksquashfs", dir, tmp.Name(), "-noappend").CombinedOutput(); err != nil {
		os.Remove(tmp.Name())
		return nil, fmt.Errorf("mksquashfs error: %s: %s", err, out)
	}

	h := sha512.New()
	length, err := io.Copy(h, tmp)
	if err != nil {
		os.Remove(tmp.Name())
		return nil, err
	}

	sha512 := hex.EncodeToString(h.Sum(nil))
	dst := filepath.Join(b.layerDir, sha512)
	if err := os.Rename(tmp.Name(), dst); err != nil {
		return nil, err
	}

	return &ct.ImageLayer{
		Type:       ct.ImageLayerTypeSquashfs,
		Length:     length,
		Mountpoint: "/",
		Hashes:     map[string]string{"sha512": sha512},
	}, nil
}
