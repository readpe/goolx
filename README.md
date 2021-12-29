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
Also see [OlxCLI](https://github.com/readpe/olxcli) for example usage. 

[Example: Bus faults](example/readme.go)
```go
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
```

Output:
```bash
go run .\example\readme.go
1. Bus Fault on:          28 ARIZONA          132. kV 3LG
1. Bus Fault on:           2 CLAYTOR          132. kV 3LG
1. Bus Fault on:           5 FIELDALE         132. kV 3LG
1. Bus Fault on:           6 NEVADA           132. kV 3LG
1. Bus Fault on:          10 NEW HAMPSHR      33.  kV 3LG
1. Bus Fault on:           7 OHIO             132. kV 3LG
1. Bus Fault on:           8 REUSENS          132. kV 3LG
1. Bus Fault on:          11 ROANOKE          13.8 kV 3LG
1. Bus Fault on:           4 TENNESSEE        132. kV 3LG
```

## Disclaimer
This project is not endorsed or affiliated with ASPEN Inc.

## Acknowledgements
Thanks to ASPEN Inc. for providing the Python API which inspired the development of this module.
