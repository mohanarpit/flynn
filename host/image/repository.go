package image

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	ct "github.com/flynn/flynn/controller/types"
)

func NewRepository(root string) (*Repository, error) {
	manifestDir := filepath.Join(root, "manifests")
	layerDir := filepath.Join(root, "layers")
	mntDir := filepath.Join(root, "mnt")
	tmpDir := filepath.Join(root, "tmp")
	for _, dir := range []string{root, manifestDir, mntDir, layerDir, tmpDir} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, err
		}
	}
	return &Repository{
		Root:        root,
		manifestDir: manifestDir,
		layerDir:    layerDir,
		mntDir:      mntDir,
		tmpDir:      tmpDir,
	}, nil
}

type Repository struct {
	Root string

	manifestDir string
	layerDir    string
	mntDir      string
	tmpDir      string
}

func (r *Repository) manifestPath(id string) string {
	return filepath.Join(r.manifestDir, fmt.Sprintf("%s.json", id))
}

func (r *Repository) CreateLayer(dir string) (*ct.ImageLayer, error) {
	tmp, err := ioutil.TempFile(r.tmpDir, "squashfs-")
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
	dst := filepath.Join(r.layerDir, sha512)
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
