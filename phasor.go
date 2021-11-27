// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package goolx

import (
	"fmt"
	"math"
	"math/cmplx"
)

var (
	a1 = NewPhasor(1, 120)
	a2 = NewPhasor(1, 240)
)

// PhaseToSeq converts from phase values to sequential components. Three phase only.
func PhaseToSeq(a, b, c Phasor) (seq0, seq1, seq2 Phasor) {
	seq0 = (1.0 / 3.0) * (a + b + c)
	seq1 = (1.0 / 3.0) * (a + a1*b + a2*c)
	seq2 = (1.0 / 3.0) * (a + a2*b + a1*c)
	return
}

// SeqToPhase converts from sequential compoents to phase values. Three phase only.
func SeqToPhase(seq0, seq1, seq2 Phasor) (a, b, c Phasor) {
	a = seq0 + seq1 + seq2
	b = seq0 + a2*seq1 + a1*seq2
	c = seq0 + a1*seq1 + a2*seq2
	return
}

// Phasor represents a phasor value for common power system calculations.
type Phasor complex128

// NewPhasor returns a new phasor instance.
func NewPhasor(mag, ang float64) Phasor {
	return Phasor(cmplx.Rect(mag, ang*math.Pi/180.0))
}

// Mag returns the Phasor absolute magnitude.
func (p Phasor) Mag() float64 {
	r, _ := cmplx.Polar(complex128(p))
	return r
}

// Ang returns the Phasor angle in degrees.
func (p Phasor) Ang() float64 {
	if p.Mag() < 1e-6 {
		return 0
	}
	θ := cmplx.Phase(complex128(p))
	return θ * 180.0 / math.Pi
}

// Rect returns the complex128 representation.
func (p Phasor) Rect() complex128 {
	return complex128(p)
}

// String implements the stringer interface for the Phasor type.
func (p Phasor) String() string {
	return fmt.Sprintf("%0.2f\u2220%0.1f\u00B0", p.Mag(), p.Ang())
}
