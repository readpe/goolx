// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package goolx

// FaultConn represents a fault connection for use with the DoFault procedure.
// The index and code are specified in ASPEN Oneliner documentation.
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

// OutageOption represents a the method outages are applied with use in the DoFault procedure.
type OutageOption int

const (
	OutageOptionOnePer OutageOption = iota // One at a time
	OutageOptionTwoPer                     // Two at a time
	OutageOptionAll                        // All at once
	OutageOptionBF                         // Breaker failure
)

// FaultConfig represents configuration parameters required to run the Oneliner DoFault procedure.
// Options are configured by passing one or more of the FaultOption functions provided into the
// NewFaultConfig function.
type FaultConfig struct {
	fltConn    [4]int
	fltOpt     [15]float64
	outageList []int
	outageOpt  [4]int
	fltR       float64
	fltX       float64
	clearPrev  bool
}

// Apply will apply the provided FaultOption functions to the existing config.
func (cfg *FaultConfig) Apply(options ...FaultOption) {
	for _, f := range options {
		f(cfg)
	}
}

// NewFaultConfig returns a pointer to a new instance of FaultConfig for use with the Oneliner
// DoFault procedure. Provide FaultOption functions to modify the underlying parameters.
func NewFaultConfig(options ...FaultOption) *FaultConfig {
	fc := &FaultConfig{}
	fc.Apply(options...)
	return fc
}

// New3LGFaultConfig is a helper function to return a new 3LG FaultConfig instance. See NewFaultConfig for
// more specifics.
func New3LGFaultConfig() *FaultConfig {
	return NewFaultConfig(FaultConn(ABC))
}

// New1LGFaultConfig is a helper function to return a new 1LG FaultConfig instance. See NewFaultConfig for
// more specifics.
func New1LGFaultConfig() *FaultConfig {
	return NewFaultConfig(FaultConn(AG))
}

// FaultOption represents configuration modification functions to apply to the FaultConfig data.
type FaultOption func(*FaultConfig)

// FaultRX sets the fault impedance in Ohms.
func FaultRX(r, x float64) FaultOption {
	return func(fc *FaultConfig) {
		fc.fltR = r
		fc.fltX = x
	}
}

// FaultClearPrev sets the clear previous flag. True will clear the
// previous fault results.
func FaultClearPrev(e bool) FaultOption {
	return func(fc *FaultConfig) {
		fc.clearPrev = e
	}
}

// FaultConn applies the provided fault connections. Overrides the previous fault connections.
func FaultConn(conn ...faultConn) FaultOption {
	var fltConn [4]int
	for _, c := range conn {
		fltConn[c.idx] = c.code
	}
	return func(fc *FaultConfig) {
		fc.fltConn = fltConn
	}
}

// withOutage is a middleware function to apply outage configuration options to an existing
// FaultOption function.
func withOutage(f FaultOption, outageList []int, otgOpt OutageOption) FaultOption {
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
		f(fc) // Call wrapped FaultOption function first.
		fc.outageList = outageList
		fc.outageOpt = outageOptions
	}
}

// FaultCloseIn applies a close-in fault.
func FaultCloseIn() FaultOption {
	return func(fc *FaultConfig) {
		fc.fltOpt[0] = 1
	}
}

// FaultCloseInOutage applies a close-in fault with outages on the system. An outage list and options are required.
func FaultCloseInOutage(outageList []int, otgOpt OutageOption) FaultOption {
	return withOutage(func(fc *FaultConfig) {
		fc.fltOpt[1] = 1
	}, outageList, otgOpt)
}

// FaultCloseInEndOpen applies a close-in fault with the end open as determined by the Oneliner algorithm for determining the
// end of the line.
func FaultCloseInEndOpen() FaultOption {
	return func(fc *FaultConfig) {
		fc.fltOpt[2] = 1
	}
}

// FaultCloseInEndOpenOutage applies a close-in fault with outages and the end open as determined by the Oneliner algorithm for determining the
// end of the line. An outage list and options are required.
func FaultCloseInEndOpenOutage(outageList []int, otgOpt OutageOption) FaultOption {
	return withOutage(func(fc *FaultConfig) {
		fc.fltOpt[3] = 1
	}, outageList, otgOpt)
}

// FaultRemoteBus applies a remote bus fault.
func FaultRemoteBus() FaultOption {
	return func(fc *FaultConfig) {
		fc.fltOpt[4] = 1
	}
}

// FaultRemoteBusOutage applies a remote bus fault with outages. An outage list and options are required.
func FaultRemoteBusOutage(outageList []int, otgOpt OutageOption) FaultOption {
	return withOutage(func(fc *FaultConfig) {
		fc.fltOpt[5] = 1
	}, outageList, otgOpt)
}

// FaultLineEnd applies a line end fault, as determined by Oneliners algorithm for determining line end.
func FaultLineEnd() FaultOption {
	return func(fc *FaultConfig) {
		fc.fltOpt[6] = 1
	}
}

// FaultLineEndOutage applies a line end fault with outages. Line end is determined by Oneliners algorithm for determining line end. An outage list and options are required.
func FaultLineEndOutage(outageList []int, otgOpt OutageOption) FaultOption {
	return withOutage(func(fc *FaultConfig) {
		fc.fltOpt[7] = 1
	}, outageList, otgOpt)
}

// FaultIntermediate applies an intermediate fault at the provided percentage.
func FaultIntermediate(percent float64) FaultOption {
	return func(fc *FaultConfig) {
		fc.fltOpt[8] = percent
	}
}

// FaultIntermediate applies an intermediate fault with outage at the provided percentage. An outage list and options are required.
func FaultIntermediateOutage(percent float64, outageList []int, otgOpt OutageOption) FaultOption {
	return withOutage(func(fc *FaultConfig) {
		fc.fltOpt[9] = percent
	}, outageList, otgOpt)
}

// FaultIntermediate applies an intermediate fault with at the provided percentage with the end open.
func FaultIntermediateEndOpen(percent float64) FaultOption {
	return func(fc *FaultConfig) {
		fc.fltOpt[10] = percent
	}
}

// FaultIntermediateEndOpenOutage applies an intermediate fault with at the provided percentage with the end open and outages. An outage list and options are required.
func FaultIntermediateEndOpenOutage(percent float64, outageList []int, otgOpt OutageOption) FaultOption {
	return withOutage(func(fc *FaultConfig) {
		fc.fltOpt[11] = percent
	}, outageList, otgOpt)
}

// FaultIntermediateAuto applies an auto sequencing intermediate fault between from and to at the specified step.
func FaultIntermediateAuto(step, from, to float64) FaultOption {
	return func(fc *FaultConfig) {
		fc.fltOpt[9] = step
		fc.fltOpt[12] = from
		fc.fltOpt[13] = to
	}
}
