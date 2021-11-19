// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package goolx

type faultConn struct {
	idx  int
	code int
}

// Fault connections.
var (
	ABC = faultConn{idx: 0, code: 1}
	ABG = faultConn{idx: 1, code: 1}
	AG  = faultConn{idx: 2, code: 1}
	AB  = faultConn{idx: 3, code: 1}
)

type OutageOption int

const (
	OutageOptionOnePer = iota // One at a time
	OutageOptionTwoPer        // Two at a time
	OutageOptionAll           // All at once
	OutageOptionBF            // Breaker failure
)

type FaultConfig struct {
	fltConn    [4]int
	fltOpt     [15]float64
	outageList []int
	outageOpt  [4]int
	fltR       float64
	fltX       float64
	clearPrev  bool
}

func (cfg *FaultConfig) Apply(configs ...faultConfig) {
	for _, f := range configs {
		f(cfg)
	}
}

func NewFaultConfig(configs ...faultConfig) *FaultConfig {
	fc := &FaultConfig{}
	fc.Apply(configs...)
	return fc
}

func New3LGFaultConfig() *FaultConfig {
	return NewFaultConfig(FaultConn(ABC))
}

func New1LGFaultConfig() *FaultConfig {
	return NewFaultConfig(FaultConn(AG))
}

type faultConfig func(*FaultConfig)

// FaultRX sets the fault impedance.
func FaultRX(r, x float64) faultConfig {
	return func(fc *FaultConfig) {
		fc.fltR = r
		fc.fltX = x
	}
}

// FaultClearPrev sets the clear previous flag. True will clear the
// previous fault results.
func FaultClearPrev(e bool) faultConfig {
	return func(fc *FaultConfig) {
		fc.clearPrev = e
	}
}

// FaultConn applies the provided fault connections. Overrides the previous fault connections.
func FaultConn(conn ...faultConn) faultConfig {
	var fltConn [4]int
	for _, c := range conn {
		fltConn[c.idx] = c.code
	}
	return func(fc *FaultConfig) {
		fc.fltConn = fltConn
	}
}

// withOutage is a middleware function to apply outage configuration options to an existing
// faultConfig function.
func withOutage(f faultConfig, outageList []int, otgOpt OutageOption) faultConfig {
	// Set the outage options array.
	var outageOptions [4]int
	switch otgOpt {
	case OutageOptionOnePer:
		outageOptions[0] = 1
	case OutageOptionTwoPer:
		outageOptions[1] = 1
	case OutageOptionAll:
		outageOptions[2] = 1
	case OutageOptionBF:
		outageOptions[3] = 1
	}

	return func(fc *FaultConfig) {
		fc.outageList = outageList
		fc.outageOpt = outageOptions
	}
}

func FaultCloseIn() faultConfig {
	return func(fc *FaultConfig) {
		fc.fltOpt[0] = 1
	}
}

func FaultCloseInWithOutage(outageList []int, otgOpt OutageOption) faultConfig {
	return withOutage(FaultCloseIn(), outageList, otgOpt)
}

func FaultCloseInWithEndOpen() faultConfig {
	return func(fc *FaultConfig) {
		fc.fltOpt[2] = 1
	}
}

func FaultCloseInWithEndOpenWithOutage(outageList []int, otgOpt OutageOption) faultConfig {
	return withOutage(FaultCloseInWithEndOpen(), outageList, otgOpt)
}
