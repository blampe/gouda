package gouda

import (
	"image"
	"time"
)

type settings struct {
	img           *image.RGBA
	Canvas        canvas
	timeLimit     time.Duration
	maxIterations int
	minIterations int
}

func NewSettings(
	img *image.RGBA,
	width int,
	height int,
	radius float64,
	timeLimit int,
	maxIterations int,
	minIterations int,
) settings {

	canvas := canvas{
		Width:  width,
		Height: height,
		minw:   -radius,
		maxw:   radius,
		minh:   -radius,
		maxh:   radius,
	}

	return settings{
		img:           img,
		Canvas:        canvas,
		timeLimit:     time.Second * time.Duration(timeLimit),
		maxIterations: maxIterations,
		minIterations: minIterations,
	}
}
