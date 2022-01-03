// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/readpe/goolx"
)

func main() {
	api := goolx.NewClient()
	defer api.Release() // releases api dll at function return

	// Load a oneliner case into memory.
	if err := api.LoadDataFile(`local\SAMPLE09.OLR`); err != nil {
		log.Fatal(err)
	}

	// Define a 3LG and 1LG bus fault config for DoFault.
	fltCfg := goolx.NewFaultConfig(
		goolx.FaultCloseIn(),                 // Bus Handle + Close In = bus fault
		goolx.FaultConn(goolx.ABC, goolx.AG), // 3LG and 1LG (A-gnd)
		goolx.FaultClearPrev(true),           // Clear previous results
	)

	// Loop through all buses in case using NextEquipment iterator.
	for bi := api.NextEquipment(goolx.TCBus); bi.Next(); {
		hnd := bi.Hnd()

		// Run pre-defined fault config for bus.
		if err := api.DoFault(hnd, fltCfg); err != nil {
			log.Println(err)
			continue
		}

		// Define tabwriter for clean spacing of output (Optional).
		tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		defer tw.Flush()

		// Loop through fault results, using iterator type wrapping Pickfault method.
		for fi := api.NextFault(1); fi.Next(); {
			fltIndex := fi.Index()
			fd := api.FaultDescription(fltIndex)

			// Get bus fault duty in phase quantities. HNDSC is the handle for total short circuit current.
			ia, ib, ic, err := api.GetSCCurrentPhase(goolx.HNDSC)
			if err != nil {
				log.Fatal(err)
			}

			// Print result to stdout using tabwriter.
			fmt.Fprintf(tw, "%q\tIa:%v\tIb:%v\tIc:%v\t\n", fd, ia, ib, ic)
		}
	}
}
