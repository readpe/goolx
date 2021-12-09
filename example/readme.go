// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"

	"github.com/readpe/goolx"
	"github.com/readpe/goolx/model"
)

func main() {
	c := goolx.NewClient()
	defer c.Release() // releases api dll at function return

	// Print the olxapi info.
	info := c.Info()
	fmt.Println(info)

	// Load a oneliner case into memory.
	err := c.LoadDataFile("system.olr")
	if err != nil {
		log.Fatal(err)
	}

	// Loop through all buses in case using NextEquipment iterator.
	buses := c.NextEquipment(goolx.TCBus)
	for buses.Next() {
		hnd := buses.Hnd()

		// Get bus data
		b, err := model.GetBus(c, hnd)
		if err != nil {
			log.Println(fmt.Errorf("could not get bus data: %v", err))
			continue
		}
		fmt.Printf("found bus %s %fkV with handle: %d", b.Name, b.KVNominal, b.HND)
	}
}
