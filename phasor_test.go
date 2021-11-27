// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package goolx

import "testing"

func Test_PhaseToSeq(t *testing.T) {
	va := NewPhasor(0, 0)
	vb := NewPhasor(82.909, -125.0)
	vc := NewPhasor(81.778, 128.8)
	v0, v1, v2 := PhaseToSeq(va, vb, vc)
	vaCalc, vbCalc, vcCalc := SeqToPhase(v0, v1, v2)
	if va.String() != vaCalc.String() {
		t.Errorf("expected %s, got %s", va, vaCalc)
	}
	if vb.String() != vbCalc.String() {
		t.Errorf("expected %s, got %s", vb, vbCalc)
	}
	if vc.String() != vcCalc.String() {
		t.Errorf("expected %s, got %s", vc, vcCalc)
	}
}
