// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package olxapi

import (
	"fmt"
	"math"
	"os"
	"strings"
	"unsafe"
)

// utf8NullFromString returns UTF-8 string with a terminating NUL added.
// If s contains a NUL byte at any location, it returns (nil, error).
func UTF8NullFromString(s string) ([]byte, error) {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return nil, fmt.Errorf("unable to encode string %s to utf-8", s)
		}
	}
	return []byte(s + "\x00"), nil
}

// utf8NullToString returns the UTF-8 encoding of the UTF-8 sequence s,
// with a terminating NUL removed.
func UTF8NullToString(s []byte) string {
	for i, v := range s {
		if v == 0 {
			s = s[0:i]
			break
		}
	}
	return string(s)
}

// utf8PtrToString takes a pointer to a UTF-8 encoded null terminated,
// character byte array, example is a char* from C.
func utf8StringFromPtr(p uintptr) string {
	buf := strings.Builder{}
	// increment pointer 1 byte at a time until null character found.
	for p := p; ; p++ {
		// go vet shows as misuse of unsafe.Pointer, tested ok
		b := *(*byte)(unsafe.Pointer(p))
		if b == 0 {
			// null termination found
			break
		}
		buf.WriteByte(b)
	}
	return buf.String()
}

// float64ToUint32 converts a float64 to two uint32. This is needed in order to pass
// a C double (float64) to the 32 bit dll using uintptr.
func float64ToUint32(f float64) [2]uint32 {
	f64 := math.Float64bits(f)
	return *(*[2]uint32)(unsafe.Pointer(&f64))
}

// tempChdir temporarily changes the directory. Returns
// a callback function to return the directory to the original.
func tempChdir(dir string) (func() error, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	err = os.Chdir(dir)
	if err != nil {
		return nil, err
	}
	return func() error {
		return os.Chdir(cwd)
	}, nil
}
