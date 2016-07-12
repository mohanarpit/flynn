package image

import (
	"bufio"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/flynn/flynn/pkg/random"
	"github.com/tent/canonical-json-go"
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

func (r *Repository) CreateImage(dir, tag string, parent *Image) (*Image, error) {
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
	if _, err := io.Copy(h, tmp); err != nil {
		os.Remove(tmp.Name())
		return nil, err
	}

	layerID := fmt.Sprintf("%s.squashfs", hex.EncodeToString(h.Sum(nil)))
	dst := filepath.Join(r.layerDir, layerID)
	if err := os.Rename(tmp.Name(), dst); err != nil {
		return nil, err
	}

	var layers []*Layer
	if parent != nil {
		layers = parent.Layers
	}
	image := &Image{
		Version: Version1,
		Layers: append(layers, &Layer{
			ID:   layerID,
			Type: LayerTypeSquashfs,
		}),
	}
	data, err := cjson.Marshal(image)
	if err != nil {
		return nil, err
	}
	sum := sha512.Sum512(data)
	id := hex.EncodeToString(sum[:])

	manifest, err := ioutil.TempFile(r.tmpDir, "manifest-")
	if err != nil {
		return nil, err
	}
	defer manifest.Close()

	if err := json.NewEncoder(manifest).Encode(&image); err != nil {
		os.Remove(manifest.Name())
		return nil, err
	}

	manifestPath := r.manifestPath(id)
	if err := os.Rename(manifest.Name(), manifestPath); err != nil {
		return nil, err
	}
	if tag == "" {
		return image, nil
	}
	tagPath := r.manifestPath(tag)
	if err := os.Symlink(manifestPath, tagPath); err != nil && !os.IsExist(err) {
		return nil, err
	}
	return image, nil
}

func (r *Repository) Lookup(id string) (*Image, error) {
	f, err := os.Open(r.manifestPath(id))
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	defer f.Close()
	var image Image
	return &image, json.NewDecoder(f).Decode(&image)
}

func (r *Repository) Checkout(name string) (string, error) {
	image, err := LoadImage(filepath.Join(r.manifestDir, fmt.Sprintf("%s.json", name)))
	if err != nil {
		return "", err
	}

	mounts := make([]string, len(image.Layers))
	for i, layer := range image.Layers {
		path, err := r.Mount(layer)
		if err != nil {
			return "", err
		}
		// append mount paths in reverse order as overlay
		// lower dirs are stacked from right to left
		mounts[len(image.Layers)-i-1] = path
	}

	workDir := filepath.Join(r.tmpDir, "work-"+random.UUID())
	if err := os.Mkdir(workDir, 0755); err != nil {
		return "", err
	}
	upperDir := filepath.Join(r.tmpDir, "upper-"+random.UUID())
	if err := os.Mkdir(upperDir, 0755); err != nil {
		os.RemoveAll(workDir)
		return "", err
	}

	mountData := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", strings.Join(mounts, ":"), upperDir, workDir)
	if err := syscall.Mount("overlay", upperDir, "overlay", 0, mountData); err != nil {
		os.RemoveAll(workDir)
		os.RemoveAll(upperDir)
		return "", err
	}
	return upperDir, nil
}

func LoadImage(manifest string) (*Image, error) {
	f, err := os.Open(manifest)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var image Image
	return &image, json.NewDecoder(f).Decode(&image)
}

func (r *Repository) Mount(layer *Layer) (string, error) {
	src := filepath.Join(r.layerDir, layer.ID)
	if _, err := os.Stat(src); err != nil {
		return "", err
	}

	dst := filepath.Join(r.mntDir, layer.ID)
	if mounted, err := isMounted(dst); err != nil {
		return "", err
	} else if mounted {
		return dst, nil
	}
	if err := os.MkdirAll(dst, 0700); err != nil {
		return "", err
	}

	loopDev, err := loopMount(src)
	if err != nil {
		return "", err
	}

	if err := syscall.Mount(loopDev, dst, "squashfs", 0, ""); err != nil {
		return "", err
	}
	return dst, nil
}

func isMounted(path string) (bool, error) {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return false, err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		fields := strings.Split(s.Text(), " ")
		if fields[4] == path {
			return true, nil
		}
	}
	return false, s.Err()
}

const (
	LOOP_CTL_GET_FREE = 0x4C82
	LOOP_SET_FD       = 0x4C00
)

func loopMount(src string) (string, error) {
	ctrl, err := os.OpenFile("/dev/loop-control", os.O_RDWR, 0)
	if err != nil {
		return "", err
	}
	defer ctrl.Close()

	index, _, errno := syscall.Syscall(syscall.SYS_IOCTL, ctrl.Fd(), uintptr(LOOP_CTL_GET_FREE), 0)
	if errno != 0 {
		return "", fmt.Errorf("error allocating loop device: %s", err)
	}
	path := fmt.Sprintf("/dev/loop%d", index)

	loop, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return "", err
	}
	defer loop.Close()

	f, err := os.OpenFile(src, os.O_RDWR, 0)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, loop.Fd(), uintptr(LOOP_SET_FD), f.Fd())
	if errno != 0 {
		return "", fmt.Errorf("error setting loop device backing file: %s", err)
	}

	return path, nil
}

type Version string

const Version1 Version = "v1"

type Image struct {
	Version Version  `json:"version"`
	Layers  []*Layer `json:"layers"`
}

type LayerType string

const LayerTypeSquashfs LayerType = "squashfs"

type Layer struct {
	ID   string    `json:"id"`
	Type LayerType `json:"type"`
}
