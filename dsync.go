// Written by Eric Crosson
// 2017-05-20

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/codeskyblue/go-sh"
	"github.com/docopt/docopt-go"
)

type Image struct {
	path string
	name string
}

// Synchronize specified docker image from src to dest.
func syncImage(images []Image, dest string, wg *sync.WaitGroup) {
	defer wg.Done()
	for _, image := range images {
		remoteTemporaryImageName := path.Join("/tmp", fmt.Sprintf("dsync-%s", image.name))
		sh.Command("rsync", "-Prahz", image.path, fmt.Sprintf("%s:%s", dest, remoteTemporaryImageName)).Run()
		sh.Command("ssh", dest, "docker", "load", "-i", image.path).Run()
		sh.Command("ssh", dest, "rm", "-f", remoteTemporaryImageName).Run()
	}
}

// Save specified image to local temp-file.
func saveImage(image string) Image {
	file, err := ioutil.TempFile(os.TempDir(), "dsync")
	if err != nil {
		panic(err)
	}
	localImage := Image{path: file.Name(), name: image}
	sh.Command("docker", "save", "-o", localImage.path, image).Run()
	return localImage
}

// Remove specified image.
func (i Image) remove() {
	os.Remove(i.path)
}

func main() {
	var wg sync.WaitGroup
	// FIXME: allow multiple images to be specified with this syntax
	usage := `Docker-Image Synchronization.

Usage:
  dsync <image> to <dest>...
  dsync <dest> gets <image>...
  dsync --help
  dsync --version

Options:
  -h --help     Show this screen.
  -v --version  Show version.`

	arguments, _ := docopt.Parse(usage, nil, true,
		"Docker-Image Synchronization 1.0.0", false)

	dests := arguments["<dest>"].([]string)
	wg.Add(len(dests))

	rawImageNames := arguments["<image>"].([]string)
	var images []Image
	for _, imageName := range rawImageNames {
		image := saveImage(imageName)
		images = append(images, image)
		defer image.remove()
	}

	for _, dest := range dests {
		go syncImage(images, dest, &wg)
	}
	wg.Wait()
}
