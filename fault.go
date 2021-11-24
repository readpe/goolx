// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package goolx

// FaultConn represents a fault connection for use with the DoFault procedure.
// The index and code are specified in ASPEN Oneliner documentation.
type faultConn struct {
	idx    int
	code   int
	seCode float64 // stepped event connection code
}

// Fault connections.
// TODO(readpe): Populate remaining connection codes.
var (
	ABC = faultConn{idx: 0, code: 1, seCode: 1}
	ABG = faultConn{idx: 1, code: 1, seCode: 4}
	AG  = faultConn{idx: 2, code: 1, seCode: 5}
	AB  = faultConn{idx: 3, code: 1, seCode: 8}
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
	fc := &FaultConfig{
		outageList: make([]int, 1),
	}
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

// SteppedEventConfig represents the configuration options for running stepped
// event analysis, for use with DoSteppedEvent function.
type SteppedEventConfig struct {
	fltOpt [64]float64
	runOpt [7]int
	nTiers int
}

// NewSteppedEvent returns a new SteppedEventConfig instance with the applied options.
// The default number of tiers is 3, this can be changed using the SteppedEventTiers option.
func NewSteppedEvent(options ...SteppedEventOption) *SteppedEventConfig {
	cfg := &SteppedEventConfig{nTiers: 3}
	SteppedEventOptions(options...)(cfg)
	return cfg
}

// SteppedEventOption represents a function to modify the SteppedEventConfig options.
type SteppedEventOption func(cfg *SteppedEventConfig)

// SteppedEventOptions consolidates a list of SteppedEventOption's into a single SteppedEventOption.
func SteppedEventOptions(options ...SteppedEventOption) SteppedEventOption {
	return func(cfg *SteppedEventConfig) {
		for _, opt := range options {
			opt(cfg)
		}
	}
}

// SteppedEventConn sets the fault connection code.
func SteppedEventConn(conn faultConn) SteppedEventOption {
	return func(cfg *SteppedEventConfig) {
		cfg.fltOpt[0] = conn.seCode
	}
}

// SteppedEventCloseIn applies a close-in fault by setting the intermediate to 0.
func SteppedEventCloseIn() SteppedEventOption {
	return SteppedEventIntermediate(0)
}

// SteppedEventIntermediate applies an intermediate fault at the given percentage.
func SteppedEventIntermediate(percent float64) SteppedEventOption {
	return func(cfg *SteppedEventConfig) {
		cfg.fltOpt[1] = percent
	}
}

// SteppedEventRX sets the stepped event fault impedance.
func SteppedEventRX(r, x float64) SteppedEventOption {
	return func(cfg *SteppedEventConfig) {
		cfg.fltOpt[2] = r
		cfg.fltOpt[3] = x
	}
}

// SteppedEventAll option enabled checking all relay types during stepped event simulation.
func SteppedEventAll() SteppedEventOption {
	return func(cfg *SteppedEventConfig) {
		for i := range cfg.runOpt {
			cfg.runOpt[i] = 1
		}
	}
}

// SteppedEventOCGnd option enables ground overcurrent relays during stepped event simulation.
func SteppedEventOCGnd() SteppedEventOption {
	return func(cfg *SteppedEventConfig) {
		cfg.runOpt[0] = 1
	}
}

// SteppedEventOCPh option enables phase overcurrent relays during stepped event simulation.
func SteppedEventOCPh() SteppedEventOption {
	return func(cfg *SteppedEventConfig) {
		cfg.runOpt[1] = 1
	}
}

// SteppedEventDSGnd option enables ground distance relays during stepped event simulation.
func SteppedEventDSGnd() SteppedEventOption {
	return func(cfg *SteppedEventConfig) {
		cfg.runOpt[2] = 1
	}
}

// SteppedEventDSPh option enables phase distance relays during stepped event simulation.
func SteppedEventDSPh() SteppedEventOption {
	return func(cfg *SteppedEventConfig) {
		cfg.runOpt[3] = 1
	}
}

// SteppedEventLogicScheme option enables protection schemes during stepped event simulation.
func SteppedEventLogicScheme() SteppedEventOption {
	return func(cfg *SteppedEventConfig) {
		cfg.runOpt[4] = 1
	}
}

// SteppedEventLogicVoltRelay option enables voltage relays during stepped event simulation.
func SteppedEventLogicVoltRelay() SteppedEventOption {
	return func(cfg *SteppedEventConfig) {
		cfg.runOpt[5] = 1
	}
}

// SteppedEventDiffRelay option enables differential relays during stepped event simulation.
func SteppedEventDiffRelay() SteppedEventOption {
	return func(cfg *SteppedEventConfig) {
		cfg.runOpt[6] = 1
	}
}

// SteppedEventTiers option sets the number of tiers from the initiating equipment to be evaluated.
func SteppedEventTiers(n int) SteppedEventOption {
	return func(cfg *SteppedEventConfig) {
		cfg.nTiers = n
	}
}
