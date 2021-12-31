package goolx

import "fmt"

// Bus represents a bus data object.
type Bus struct {
	Hnd       int
	Name      string
	Area      int
	Zone      int
	Tap       int
	KVNominal float64
	KV        float64
	Angle     float64
	Location  string
	Comment   string
}

func (b *Bus) String() string {
	return fmt.Sprintf("%s %0.2f", b.Name, b.KVNominal)
}

// GetBus loads the bus data at the provided handle into a new bus object. Returns error
// if the handle provided does not point to an equipment type TCBus.
func (c *Client) GetBus(hnd int) (*Bus, error) {
	return c.getBus(hnd)
}

func (c *Client) getBus(hnd int) (*Bus, error) {
	if eqType, _ := c.EquipmentType(hnd); eqType != TCBus {
		return nil, fmt.Errorf("getBus: equipment type must be TCBus")
	}
	var bus = Bus{Hnd: hnd}
	data := c.GetData(hnd,
		BUSsName,
		BUSnArea,
		BUSnZone,
		BUSnTapBus,
		BUSdKVnominal,
		BUSdKVP,
		BUSdAngleP,
		BUSsLocation,
		BUSsComment,
	)

	if err := data.Scan(
		&bus.Name,
		&bus.Area,
		&bus.Zone,
		&bus.Tap,
		&bus.KVNominal,
		&bus.KV,
		&bus.Angle,
		&bus.Location,
		&bus.Comment,
	); err != nil {
		return nil, fmt.Errorf("getBus: could not scan bus data %v", err)
	}
	return &bus, nil
}
