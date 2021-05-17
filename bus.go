// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package goolx

import "github.com/readpe/goolx/olxapi"

// Bus represents a bus equipment data structure. This does not represent all fields from
// ASPEN model, future fields may be added as needed.
type Bus struct {
	HND       int     // ASPEN Oneliner equipment handle
	Name      string  // BUSsName
	KVNominal float64 // BUSdKVnominal
	Number    int     // BUSnNumber
	Area      int     // BUSnArea
	Zone      int     // BUSnZone
	Tap       int     // BUSnTapBus
	Comment   string  // BUSsComment (aka Memo field)
}

// GetBus retrieves the bus with the given handle using the provided api client. Data is
// Scanned into a new bus object and returned if no errors.
func GetBus(c *Client, hnd int) (*Bus, error) {
	data := c.GetData(hnd,
		olxapi.BUSsName,
		olxapi.BUSdKVnominal,
		olxapi.BUSnNumber,
		olxapi.BUSnArea,
		olxapi.BUSnZone,
		olxapi.BUSnTapBus,
		olxapi.BUSsComment,
	)

	// Scan data into bus instance. Similar to sql.Rows.Scan
	b := Bus{HND: hnd}
	err := data.Scan(
		&b.Name,
		&b.KVNominal,
		&b.Number,
		&b.Area,
		&b.Zone,
		&b.Tap,
		&b.Comment,
	)
	if err != nil {
		return nil, err
	}
	return &b, nil
}
