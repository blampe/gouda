package gouda

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/cmplx"
	"math/rand"
	"os"
	"sync"
	"time"
)

func clamp(v, min, max float64) float64 {
	if v > max {
		return max
	}
	if v < min {
		return min
	}
	return v
}

func Cheese(s *settings) *image.RGBA {
	samples := samplePoints(s)

	var wg sync.WaitGroup

	counts := make([]int, s.Canvas.Width*s.Canvas.Height)

	for i := 0; i < 16; i += 1 {

		wg.Add(1)

		go func(i int, points <-chan complex128, countRef *[]int) {
			path := make([]complex128, s.maxIterations)

			for c := range points {
				recordPath(c, countRef, &path, &s.Canvas)
			}

			wg.Done()
		}(i, samples, &counts)
	}

	wg.Wait()

	min, max := ^0, 0

	for _, v := range counts {
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
	}

	colorFunc := func(x int) float64 {
		return float64(x)
		//return math.Log2(float64(x-min) + 1)
	}

	scale := 255 / colorFunc(max)

	for x := 0; x < s.Canvas.Width; x++ {
		for y := 0; y < s.Canvas.Height; y++ {
			count := counts[x*s.Canvas.Width+y]

			col := uint8(scale * colorFunc(count))
			// Rotate 90° by switching x/y.
			s.img.SetRGBA(y, x,
				color.RGBA{
					R: col,
					G: col,
					B: col,
					A: 255,
				},
			)
		}
	}

	return s.img
}

func recordPath(c complex128, countRef *[]int, path *[]complex128, can *canvas) {

	var z complex128
	atIteration := 0
	maxIterations := len(*path)

	for ; !hasEscaped(z) && atIteration < maxIterations; atIteration += 1 {
		z = z*z + c
		(*path)[atIteration] = z
	}

	if atIteration == maxIterations {
		return
	}

	atIteration -= 1

	for ; atIteration > 0; atIteration -= 1 {
		w, h, err := can.PixelFor((*path)[atIteration])

		if err == nil {
			(*countRef)[w*can.Width+h] += 1
		}
	}
}

func path(c complex128, maxIterations int) []complex128 {

	var z complex128
	path := []complex128{}

	for iter := 0; iter < maxIterations; iter++ {
		z = z*z + c
		path = append(path, z)
		if hasEscaped(z) {
			return path
		}
	}

	return []complex128{}
}

// Take settings
// Sample coordinates from canvas
// Iterate to find boundary of mandelbrot
// Randomly select points around the boundary
func samplePoints(s *settings) chan complex128 {

	samples := make(chan complex128)

	candidates := make([]complex128, 0)

	fmt.Println("Calculating candidates for trajectory tracing...")

	// We need to reduce our candidate set to points that have not
	// escaped after MIN iterations. From this set we will plot
	// trajectories for points that DO escape within MAX iterations.
	img := image.NewRGBA(image.Rect(0, 0, s.Canvas.Width, s.Canvas.Height))

	for c := range s.Canvas.Coordinates() {
		x, y, err := s.Canvas.PixelFor(c)
		img.SetRGBA(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 255})

		// Now see if we can iterate MIN times without it escaping.
		z := complex(0, 0)

		for i := 0; i < s.minIterations; i++ {
			z = z*z + c
			if hasEscaped(z) {
				break
			}
		}

		if !hasEscaped(z) {
			candidates = append(candidates, c)
			if err == nil {
				img.SetRGBA(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
			} else {
				fmt.Println("WTF")
			}
		}
	}

	f, _ := os.Create("foo.png")
	png.Encode(f, img)
	defer f.Close()

	fmt.Println("Found", len(candidates), "candidates based on", s.minIterations, "minimum iterations.")

	go func() {
		var c complex128
		sampled := 0

		unitsPerPixel := s.Canvas.unitsPerPixel()

		fmt.Println("Rendering for the next", s.timeLimit, "...")

		done := time.After(s.timeLimit)

		candidateNumber := 0

		for {
			select {

			default:
				candidateNumber = (candidateNumber + 1) % len(candidates)

				//if candidateNumber == 0 {
				//fmt.Println("Done with one with one full round of traces")
				//}

				c = candidates[candidateNumber]

				fuzzVector := complex(rand.Float64(), rand.Float64())
				fuzzVector -= complex(0.5, 0.5)         // [0, 1] -> [-1/2, 1/2]
				fuzzVector *= complex(unitsPerPixel, 0) // scale to canvas resolution

				c += fuzzVector

				if hasEscaped(c) || willNeverEscape(c) {
					continue
				}

				sampled += 1
				samples <- c

			case <-done:
				fmt.Println("Recorded", sampled, "trajectories!", "(", float64(sampled)/float64(s.Canvas.Size()), "per pixel)")
				close(samples)
				return
			}

		}
	}()

	return samples
}

func hasEscaped(z complex128) bool {
	return cmplx.Abs(z) >= 2
}

func willNeverEscape(z complex128) bool {
	// Cartiod filtering. We know these points will never escape.
	re := real(z)
	imsq := imag(z) * imag(z)

	if (re+1)*(re+1)+imsq < 1.0/16.0 {
		return true
	}

	var q float64 = (re-1.0/4.0)*(re-1.0/4.0) + imsq

	return 4*q*(q+re-1.0/4.0) < imsq
}
