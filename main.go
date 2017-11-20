package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	"io"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	"google.golang.org/genproto/googleapis/cloud/vision/v1"

	api "cloud.google.com/go/vision/apiv1"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <path-to-image>\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "Pass a path to a local file.\n")
	}
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	path := flag.Arg(0)

	detectCropHints(os.Stdout, path)
}

func detectCropHints(w io.Writer, file string) error {
	ctx := context.Background()

	client, err := api.NewImageAnnotatorClient(ctx)
	if err != nil {
		return err
	}

	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	image, err := api.NewImageFromReader(f)
	if err != nil {
		return err
	}
	res, err := client.CropHints(ctx, image, nil)
	if err != nil {
		return err
	}

	fmt.Fprintln(w, "Crop hints:")
	for _, hint := range res.CropHints {
		for _, v := range hint.BoundingPoly.Vertices {
			fmt.Fprintf(w, "(%d,%d)\n", v.X, v.Y)
		}
	}

	err = crop(file, res.CropHints[0])
	if err != nil {
		return err
	}

	return nil
}
func crop(file string, cropHint *vision.CropHint) error {
	poly := cropHint.BoundingPoly
	rect := image.Rect(int(poly.Vertices[0].X), int(poly.Vertices[0].Y), int(poly.Vertices[2].X-1), int(poly.Vertices[2].Y-1))
	image, _ := imaging.Open(file)

	croppedIMage := imaging.Crop(image, rect)
	return imaging.Save(croppedIMage, "output-crop.jpg")
}
