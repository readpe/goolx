// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package goolx

import (
	"fmt"
	"strings"
	"unsafe"
)

// uTF8NullFromString returns UTF-8 string with a terminating NUL added.
// If s contains a NUL byte at any location, it returns (nil, error).
func uTF8NullFromString(s string) ([]byte, error) {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return nil, fmt.Errorf("unable to encode string %s to utf-8", s)
		}
	}
	return []byte(s + "\x00"), nil
}

// uTF8NullToString returns the UTF-8 encoding of the UTF-8 sequence s,
// with a terminating NUL removed.
func uTF8NullToString(s []byte) string {
	for i, v := range s {
		if v == 0 {
			s = s[0:i]
			break
		}
	}
	return string(s)
}

// uTF8PtrToString takes a pointer to a UTF-8 encoded null terminated,
// character byte array, example is a char* from C
func uTF8StringFromPtr(p uintptr) string {
	buf := strings.Builder{}
	for {
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
