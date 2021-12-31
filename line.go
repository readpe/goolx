package goolx

import "fmt"

// Line represents a line data object.
type Line struct {
	Hnd          int
	Bus1         *Bus
	Bus2         *Bus
	CktID        string
	Name         string
	InService    int
	RelayGrp1Hnd int
	RelayGrp2Hnd int
	MuPairHnd    int
	Length       float64
	LengthUnit   string

	// Line parameters.
	R, X     float64
	R0, X0   float64
	B1, G1   float64
	B10, G10 float64
	B2, G2   float64
	B20, G20 float64
}

func (l *Line) String() string {
	return fmt.Sprintf("%s-%s ckt:%s", l.Bus1, l.Bus2, l.CktID)
}

// GetLine loads the line data at the provided handle into a new line object. Returns error
// if the handle provided does not point to an equipment type TCLine.
func (c *Client) GetLine(hnd int) (*Line, error) {
	return c.getLine(hnd)
}

// getLine loads line data into a Line object.
func (c *Client) getLine(hnd int) (*Line, error) {
	if eqType, _ := c.EquipmentType(hnd); eqType != TCLine {
		return nil, fmt.Errorf("getLine: equipment type must be TCLine")
	}
	var ln = Line{Hnd: hnd}
	data := c.GetData(hnd,
		LNnBus1Hnd,
		LNnBus2Hnd,
		LNsID,
		LNsName,
		LNnInService,
		LNnMuPairHnd,
		LNdLength,
		LNsLengthUnit,
		LNdR, LNdX,
		LNdR0, LNdX0,
		LNdB1, LNdG1,
		LNdB10, LNdG10,
		LNdB2, LNdG2,
		LNdB20, LNdG20,
	)

	var bus1Hnd, bus2Hnd int
	if err := data.Scan(
		&bus1Hnd,
		&bus2Hnd,
		&ln.CktID,
		&ln.Name,
		&ln.InService,
		&ln.MuPairHnd,
		&ln.Length,
		&ln.LengthUnit,
		&ln.X, &ln.R,
		&ln.X0, &ln.R0,
		&ln.B1, &ln.G1,
		&ln.B10, &ln.G10,
		&ln.B2, &ln.G2,
		&ln.B20, &ln.G20,
	); err != nil {
		return nil, fmt.Errorf("getLine: could not scan line data %v", err)
	}

	// Ignoring error on relaygroup lookup. OlxAPI throws error if relay groups not present, we can default to zero value.
	c.GetData(hnd, LNnRlyGr1Hnd, LNnRlyGr2Hnd).Scan(&ln.RelayGrp1Hnd, &ln.RelayGrp2Hnd)

	// Get bus1 data.
	if b, _ := c.getBus(bus1Hnd); b != nil {
		ln.Bus1 = b
	}

	// Get bus2 data.
	if b, _ := c.getBus(bus2Hnd); b != nil {
		ln.Bus2 = b
	}

	return &ln, nil
}
