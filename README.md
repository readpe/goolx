[![Go Reference](https://pkg.go.dev/badge/github.com/readpe/goolx.svg)](https://pkg.go.dev/github.com/readpe/goolx)
[![Go Report Card](https://goreportcard.com/badge/github.com/readpe/goolx)](https://goreportcard.com/report/github.com/readpe/goolx)

---

**_Update 12-29-2021:_**  

goolx is not under active development at this time, feel free to submit issues or pull requests, however they may not be promptly addressed.

---

# Overview

goolx is an **_unofficial_** Go wrapper around ASPEN's Oneliner API. It can be utilized to create programs which interact with Oneliner through the provided API functions. Additionally, it provides helper functions for common analysis tasks.

The goal of this module is to provide generic unopinionated abstractions for fault analysis. These functions can be used as a building block for useful fault analysis application development.

**This library is in Version 0.x.x, and may have breaking API changes during development.** 

# Installation
It is required users of this Go module have authorized licenses for the underlying ASPEN software v15.4 or newer. Obtaining this third-party software is outside the scope of this README. 

The module can be installed using the following `go get` command.
```bash
go get -u github.com/readpe/goolx
```

This library is only designed for use on Windows 386 architecture. To properly build, ensure your `GOOS` and `GOARCH` environment variables are set appropriately:

```
GOOS=windows
GARCH=386
```

# Usage Example
For a more practical usage example, please refer to the demonstration project: [OlxCLI](https://github.com/readpe/olxcli)

[Example: Bus faults](example/readme.go)

The below example runs a 3LG and 1LG fault on every bus in the provided model, outputting the total fault current calculated.
```go
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
```

Output:
```bash
go run .\example\readme.go
"1. Bus Fault on:           4 TENNESSEE        132. kV 3LG"         Ia:4342.50∠-78.8°  Ib:4342.50∠161.2°  Ic:4342.50∠41.2°  
"2. Bus Fault on:           4 TENNESSEE        132. kV 1LG Type=A"  Ia:3690.63∠-79.4°  Ib:0.00∠0.0°       Ic:0.00∠0.0°
"1. Bus Fault on:          11 ROANOKE          13.8 kV 3LG"         Ia:50997.07∠-120.3°  Ib:50997.07∠119.7°  Ic:50997.07∠-0.3°
"2. Bus Fault on:          11 ROANOKE          13.8 kV 1LG Type=A"  Ia:47521.65∠-120.4°  Ib:0.00∠-123.2°     Ic:0.00∠-117.5°
"1. Bus Fault on:           8 REUSENS          132. kV 3LG"         Ia:5122.25∠-86.0°  Ib:5122.25∠154.0°  Ic:5122.25∠34.0°
"2. Bus Fault on:           8 REUSENS          132. kV 1LG Type=A"  Ia:5106.25∠-86.2°  Ib:0.00∠0.0°       Ic:0.00∠0.0°
"1. Bus Fault on:           7 OHIO             132. kV 3LG"         Ia:4286.88∠-80.9°  Ib:4286.88∠159.1°  Ic:4286.88∠39.1°
"2. Bus Fault on:           7 OHIO             132. kV 1LG Type=A"  Ia:4297.44∠-80.9°  Ib:0.00∠-32.8°     Ic:0.00∠-153.2°
"1. Bus Fault on:          10 NEW HAMPSHR      33.  kV 3LG"         Ia:8599.81∠-90.6°  Ib:8599.81∠149.4°  Ic:8599.81∠29.4°
"2. Bus Fault on:          10 NEW HAMPSHR      33.  kV 1LG Type=A"  Ia:8864.15∠-90.6°  Ib:0.00∠-29.7°     Ic:0.00∠-153.9°
"1. Bus Fault on:           6 NEVADA           132. kV 3LG"         Ia:5791.65∠-85.7°  Ib:5791.65∠154.3°  Ic:5791.65∠34.3°
"2. Bus Fault on:           6 NEVADA           132. kV 1LG Type=A"  Ia:5797.66∠-85.9°  Ib:0.00∠-32.4°     Ic:0.00∠-152.7°
"1. Bus Fault on:           5 FIELDALE         132. kV 3LG"         Ia:6657.30∠-86.8°  Ib:6657.30∠153.2°  Ic:6657.30∠33.2°
"2. Bus Fault on:           5 FIELDALE         132. kV 1LG Type=A"  Ia:6670.71∠-86.7°  Ib:0.00∠144.9°     Ic:0.00∠34.2°
"1. Bus Fault on:           2 CLAYTOR          132. kV 3LG"         Ia:5395.45∠-85.0°  Ib:5395.45∠155.0°  Ic:5395.45∠35.0°
"2. Bus Fault on:           2 CLAYTOR          132. kV 1LG Type=A"  Ia:5333.50∠-85.1°  Ib:0.00∠0.0°       Ic:0.00∠0.0°
"1. Bus Fault on:          28 ARIZONA          132. kV 3LG"         Ia:3659.65∠-82.4°  Ib:3659.65∠157.6°  Ic:3659.65∠37.6°
"2. Bus Fault on:          28 ARIZONA          132. kV 1LG Type=A"  Ia:2890.36∠-81.2°  Ib:0.00∠-42.4°     Ic:0.00∠-143.8°
```

## Disclaimer
This project is not endorsed or affiliated with ASPEN Inc.

## Acknowledgements
Thanks to ASPEN Inc. for providing the Python API which inspired the development of this module.
