package gouda

import (
	"errors"
	"math"
)

type canvas struct {
	Width, Height int
	minw, maxw    float64
	minh, maxh    float64
	// offset, zoom?
}

func (c *canvas) CoordinateFor(w, h int) (complex128, error) {
	if w < 0 || w > c.Width || h < 0 || h > c.Height {
		panic("shit")
		return 0, errors.New("out of bounds")
	}

	// TODO: fuzz

	x := (float64(w)/float64(c.Width))*(c.maxw-c.minw) + c.minw
	y := (float64(h)/float64(c.Height))*(c.maxh-c.minh) + c.minh

	return complex(x, y), nil
}

func (c *canvas) PixelFor(z complex128) (int, int, error) {
	x, y := real(z), imag(z)

	if (x < c.minw) || (x > c.maxw) || (y < c.minh) || (y > c.maxh) {
		return -1, -1, errors.New("out of bounds")
	}

	w := round((x - c.minw) / (c.maxw - c.minw) * float64(c.Width))
	h := round((y - c.minh) / (c.maxh - c.minh) * float64(c.Height))

	if w == c.Width {
		w -= 1
	}

	if h == c.Height {
		h -= 1
	}

	return w, h, nil
}

func (c *canvas) Coordinates() chan complex128 {
	var coords chan complex128 = make(chan complex128)

	go func() {
		for w := 0; w < c.Width; w++ {
			for h := 0; h < c.Height; h++ {
				coord, err := c.CoordinateFor(w, h)
				if err == nil {
					coords <- coord
				}
			}
		}
		close(coords)
	}()

	return coords
}

func (c *canvas) Size() int {
	return c.Width * c.Height
}

func round(n float64) int {
	if n < 0 {
		return int(math.Ceil(n - 0.5))
	}
	return int(math.Floor(n + 0.5))
}

func (c *canvas) unitsPerPixel() float64 {
	return (c.maxw - c.minw) / float64(c.Width)
}
