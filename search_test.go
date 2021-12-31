// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package goolx

import (
	"fmt"
	"testing"
)

func TestClient_FindLine(t *testing.T) {
	api := NewClient()
	defer api.Release()

	if err := api.LoadDataFile(testCase); err != nil {
		t.Error(err)
	}

	tests := []struct {
		name     string
		fName    string
		fKV      float64
		tName    string
		tKV      float64
		ckt      string
		expected string
	}{
		{name: "line", fName: "CLAYTOR", fKV: 132, tName: "NEVADA", tKV: 132, ckt: "1", expected: "CLAYTOR 132.00-NEVADA 132.00 ckt:1"},
		{name: "phase shifter", fName: "TENNESSEE", fKV: 132, tName: "NEVADA", tKV: 132, ckt: "1", expected: "<nil>"},
		{name: "xfmr3", fName: "NEVADA", fKV: 132, tName: "NEW HAMPSHR", tKV: 33, ckt: "1", expected: "<nil>"},
		{name: "line", fName: "FIELDALE", fKV: 132, tName: "OHIO", tKV: 132, ckt: "1", expected: "FIELDALE 132.00-OHIO 132.00 ckt:1"},
		{name: "line", fName: "OHIO", fKV: 132, tName: "FIELDALE", tKV: 132, ckt: "1", expected: "FIELDALE 132.00-OHIO 132.00 ckt:1"},
		{name: "not found", fName: "OHIO", fKV: 132, tName: "ERROR", tKV: 132, ckt: "1", expected: "<nil>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ln, err := api.FindLine(tt.fName, tt.fKV, tt.tName, tt.tKV, tt.ckt)
			if err != nil {
				t.Log(err)
			}
			got := fmt.Sprint(ln)
			if got != tt.expected {
				t.Errorf("got %s, expected %s", got, tt.expected)
			}
		})
	}

}
