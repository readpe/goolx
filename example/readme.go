// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"

	"github.com/readpe/goolx"
)

func main() {
	api := goolx.NewClient()
	defer api.Release() // releases api dll at function return

	// Load a oneliner case into memory.
	if err := api.LoadDataFile(`local\SAMPLE09.OLR`); err != nil {
		log.Fatal(err)
	}

	// Define a 3LG bus fault config for DoFault.
	fltCfg := goolx.NewFaultConfig(goolx.FaultCloseIn(), goolx.FaultConn(goolx.ABC), goolx.FaultClearPrev(true))

	// Loop through all buses in case using NextEquipment iterator.
	for bi := api.NextEquipment(goolx.TCBus); bi.Next(); {
		hnd := bi.Hnd()

		// Run 3LG bus fault with defined fault config.
		if err := api.DoFault(hnd, fltCfg); err != nil {
			log.Println(err)
			continue
		}

		fd := api.FaultDescription(1)
		fmt.Println(fd)
	}
}
