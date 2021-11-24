
[![Go Report Card](https://goreportcard.com/badge/github.com/readpe/goolx)](https://goreportcard.com/report/github.com/readpe/goolx)


# Overview

`Disclaimer: This project is not endorsed or affiliated with ASPEN Inc.`

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
```go
import (
	"fmt"
	"log"

	"github.com/readpe/goolx"
	"github.com/readpe/goolx/olxapi"
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
	buses := c.NextEquipment(olxapi.TCBus)
	for buses.Next() {
		hnd := buses.Hnd()

		// Get bus data
		b, err := goolx.GetBus(c, hnd)
		if err != nil {
			log.Println(fmt.Errorf("could not get bus data: %v", err))
			continue
		}
		fmt.Printf("found bus %s %fkV with handle: %d", b.Name, b.KVNominal, b.HND)
	}
}
```

## Acknowledgements
Thanks to ASPEN Inc. for providing the Python API which inspired the development of this module.
