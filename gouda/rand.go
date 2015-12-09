package gouda

import (
	"math/rand"

	srand "github.com/phil-mansfield/gotetra/math/rand"
)

type pointGenerator interface {
	// Generate a random complex number in the range
	// [0, 1] x [0, 1]
	Next() complex128
}

type randGenerator struct {
}

func (r *randGenerator) Next() complex128 {
	return complex(rand.Float64(), rand.Float64())
}

type sobolGenerator struct {
	// An initialized SobolSequence
	seq *srand.SobolSequence
}

func (r *sobolGenerator) Next() complex128 {
	r.seq.Init()

	v := r.seq.Next(2)
	return complex(v[0], v[1])
}
