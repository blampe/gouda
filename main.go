package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime/pprof"

	"github.com/blampe/gouda/gouda"
)

var (
	output = flag.String("o", "", "Name of the output image file.")
	size   = flag.Int("s", 800, "Width/height of the output image in pixels.")
	//radius        = flag.Float64("r", 2.0, "Maximum x/y value shown.")
	timeLimit = flag.Int("t", 10,
		"Number of seconds to spend rendering the Buddhabrot.",
	)
	minIterations = flag.Int(
		"i", 10,
		("Minimum number of iterations before a point's trajectory will " +
			"be recorded. Larger values reduce 'haziness' due to quickly " +
			"escaping points."),
	)
	maxIterations = flag.Int(
		"I", 10000,
		("Maximum number of iterations before a point is considered " +
			"outside of the Mandelbrot set. Larger values increase finer " +
			"details but make trajectories take longer to render."),
	)
	profile = flag.String("profile", "", "Write CPU profile to the given file.")
)

func main() {
	flag.Parse()

	if *maxIterations < *minIterations {
		*maxIterations = 10 * *minIterations
	}

	img := image.NewRGBA(image.Rect(0, 0, *size, *size))

	settings := gouda.NewSettings(
		img,
		*size,
		*size,
		2.0,
		*timeLimit,
		*maxIterations,
		*minIterations,
	)

	if *output == "" {
		*output = fmt.Sprintf(
			"-s %d -t %d -i %d -I %d.png",
			*size,
			*timeLimit,
			*minIterations,
			*maxIterations,
		)
	}

	if *profile != "" {
		f, err := os.Create(*profile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	gouda.Cheese(&settings)

	fmt.Println("Saving results to", *output)

	f, err := os.Create(*output)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	fileFormat := filepath.Ext(*output)

	switch fileFormat {
	case ".png":
		err = png.Encode(f, img)
	case ".jpg", ".jpeg":
		err = jpeg.Encode(f, img, nil)
	case ".gif":
		err = gif.Encode(f, img, nil)
	default:
		err = errors.New("unknown file format " + fileFormat)
	}

	if err != nil {
		log.Fatal(err)
	}
}
