// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package goolx

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/readpe/goolx/constants"
)

var testCase = `C:\Program Files (x86)\ASPEN\1LPFv15\SAMPLE09.OLR`

func TestInfo(t *testing.T) {
	c := NewClient()
	got := c.Info()
	if got == "" {
		t.Errorf("info string is empty")
	}
	t.Log(got)
}

func TestVersion(t *testing.T) {
	c := NewClient()
	got, err := c.Version()
	if err != nil {
		t.Error(err)
	}
	if got == "" {
		t.Error("version string is empty")
	}
	t.Log(got)
}

func TestBuildNumber(t *testing.T) {
	c := NewClient()
	got, err := c.BuildNumber()
	if err != nil {
		t.Error(err)
	}
	if got == 0 {
		t.Error("build number is empty")
	}
	var supportedBuild = 17321
	if got < supportedBuild {
		t.Errorf("only support build number > %d", supportedBuild)
	}
	t.Logf("Build Number: %d", got)
}

// func TestSaveDatafile(t *testing.T) {
// 	c := NewClient()
// 	tmp, err := ioutil.TempDir("", "goolx")
// 	if err != nil {
// 		t.Error(tmp)
// 	}
// 	err = c.SaveDataFile(path.Join(tmp, "test.olr"))
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	err = c.SaveDataFile(path.Join(tmp, "temp", "test.olr"))
// 	if err == nil {
// 		t.Errorf("expected directory doesn't exist error, got nil")
// 	}
// }

func TestLoadDatafile(t *testing.T) {
	c := NewClient()
	tmp, err := ioutil.TempDir("", "goolx")
	if err != nil {
		t.Error(tmp)
	}
	err = c.LoadDataFile(path.Join(tmp, "test.olr"))
	if err == nil {
		t.Errorf("expected file doesn't exist error, got nil")
	}
	// testing empty olr case
	var b []byte
	os.WriteFile(path.Join(tmp, "test.olr"), b, 0700)
	err = c.LoadDataFile(path.Join(tmp, "test.olr"))
	if err == nil {
		t.Errorf("expected 'Failed to read OLR file', got nil")
	}

	err = c.LoadDataFile(testCase)
	if err != nil {
		t.Error(err)
	}
}

func TestCloseDataFile(t *testing.T) {
	c := NewClient()
	err := c.CloseDataFile()
	if err != nil {
		t.Error(err)
	}
}

func TestReadChangeFile(t *testing.T) {
	c := NewClient()
	tmp, err := ioutil.TempDir("", "goolx")
	if err != nil {
		t.Error(tmp)
	}
	err = c.ReadChangeFile(path.Join(tmp, "test.chf"))
	if err == nil {
		t.Errorf("expected file doesn't exist error, got %v", err)
	}
	// testing empty changefile
	var b []byte
	err = os.WriteFile(path.Join(tmp, "test.chf"), b, 0700)
	if err != nil {
		t.Error(err)
	}
	err = c.ReadChangeFile(path.Join(tmp, "test.chf"))
	if err == nil {
		t.Errorf("expected malformed changefile error, got %v", err)
	}
}

func TestGetEquipment(t *testing.T) {
	c := NewClient()
	defer c.Release()
	err := c.LoadDataFile(testCase)
	if err != nil {
		t.Error(err)
	}
	var hnd int
	err = c.olxAPI.GetEquipment(constants.TCBus, &hnd)
	if err != nil {
		t.Error(err)
	}
	t.Log(err, hnd)

}

func TestGetEquipmentType(t *testing.T) {
	c := NewClient()
	defer c.Release()
	err := c.LoadDataFile(testCase)
	if err != nil {
		t.Error(err)
	}
	var hnd int
	err = c.olxAPI.GetEquipment(constants.TCBus, &hnd)
	if err != nil {
		t.Error(err)
	}
	eqType, err := c.EquipmentType(hnd)
	if err != nil {
		t.Error(err)
		t.Log(eqType, err)
	}
	if eqType != constants.TCBus {
		t.Errorf("expected eqType %d, got %d", constants.TCBus, eqType)
	}
	t.Log(err, hnd)
}

func TestDeleteEquipment(t *testing.T) {
	c := NewClient()
	err := c.DeleteEquipment(0)
	if err == nil {
		t.Errorf("expected 'DeleteObj failure: Invalid Device Handle' error, got %v", err)
	}
}

func TestNextEquipment(t *testing.T) {
	c := NewClient()
	defer c.Release()
	c.LoadDataFile(testCase)
	hi := c.NextEquipment(constants.TCBus)
	var handles []int
	for hi.Next() {
		hnd := hi.Hnd()
		handles = append(handles, hnd)
	}
	expected := 9
	got := len(handles)
	if got != expected {
		t.Errorf("expected %d bus handles got %d", expected, got)
	}
}

func TestNextBusEquipment(t *testing.T) {
	c := NewClient()
	defer c.Release()
	c.LoadDataFile(testCase)
	hi := c.NextEquipment(constants.TCBus)
	var handles []int
	var branches []int
	for hi.Next() {
		handles = append(handles, hi.Hnd())
		brs := c.NextBusEquipment(hi.Hnd(), constants.TCBranch)
		for brs.Next() {
			branches = append(branches, brs.Hnd())
		}
	}
	expected := 9
	got := len(handles)
	if got != expected {
		t.Errorf("expected %d bus handles got %d", expected, got)
	}
	expected = 23
	got = len(branches)
	if got != expected {
		t.Errorf("expected %d bus handles got %d", expected, got)
	}

}

// TODO (readpe): Get passing test.
// func TestFindEquipmentByTag(t *testing.T) {
// 	c := NewClient()
// 	defer c.Release()
// 	hi := c.NextEquipmentByTag(TCBus, "Tag1", "Tag2", "Tag3")
// 	var handles []int
// 	for hi.Next() {
// 		hnd := hi.Hnd()
// 		handles = append(handles, hnd)
// 	}
// 	expected := 0
// 	got := len(handles)
// 	if got != expected {
// 		t.Errorf("expected %d bus handles got %d", expected, got)
// 	}
// }

func TestDoFault(t *testing.T) {
	c := NewClient()
	defer c.Release()
	c.LoadDataFile(testCase)

	// Can't run many of the fault options on the bus handle, need to select branch or relay group.
	hnd, err := c.FindBusByName("TENNESSEE", 132)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("Okay", func(t *testing.T) {
		config := NewFaultConfig()
		if err := c.DoFault(hnd, config); err == nil {
			t.Errorf("expected `no fault connection selected` error, got %v", err)
		}
	})
	t.Run("nil config", func(t *testing.T) {
		if err := c.DoFault(hnd, nil); err == nil {
			t.Errorf("expected non-nil error")
		}
	})
	t.Run("Options", func(t *testing.T) {
		tests := []struct {
			name   string
			config *FaultConfig
			want   string
		}{
			{
				name:   "3LG,Close-in",
				config: NewFaultConfig(FaultConn(ABC), FaultCloseIn(), FaultClearPrev(true)),
				want:   "1. Bus Fault on:           4 TENNESSEE        132. kV 3LG",
			},
			{
				name:   "1LG,Close-in",
				config: NewFaultConfig(FaultConn(AG), FaultCloseIn(), FaultClearPrev(true)),
				want:   "1. Bus Fault on:           4 TENNESSEE        132. kV 1LG Type=A",
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				test.config.fltOpt[2] = 1
				if err := c.DoFault(hnd, test.config); err != nil {
					t.Error(err)
				}
				got := c.FaultDescription(0)
				if test.want != got {
					t.Errorf("expected '%s', got '%s'", test.want, got)
				}
			})
		}
	})
}

func TestDoSteppedEvent(t *testing.T) {
	c := NewClient()
	c.LoadDataFile(testCase)
	// Can't run many of the fault options on the bus handle, need to select branch or relay group.
	hnd, err := c.FindBusByName("TENNESSEE", 132)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("Invalid device handle", func(t *testing.T) {
		cfg := NewSteppedEvent()
		err := c.DoSteppedEvent(0, cfg)
		if err == nil {
			t.Errorf("expected 'invalid device handle' error, got %v", err)
		}
	})
	t.Run("Options", func(t *testing.T) {
		tests := []struct {
			name    string
			config  *SteppedEventConfig
			want    string
			wantErr error
		}{
			{
				name:    "3LG,Close-in",
				config:  NewSteppedEvent(SteppedEventConn(ABC), SteppedEventAll(), SteppedEventCloseIn()),
				want:    "1. Simultaneous Fault:\n     Bus Fault on:           4 TENNESSEE        132. kV 3LG",
				wantErr: nil,
			},
			{
				name:    "1LG,Close-in",
				config:  NewSteppedEvent(SteppedEventConn(AG), SteppedEventAll(), SteppedEventCloseIn()),
				want:    "1. Simultaneous Fault:\n     Bus Fault on:           4 TENNESSEE        132. kV 1LG Type=A",
				wantErr: nil,
			},
			{
				name:    "1LG,Intermediate-50",
				config:  NewSteppedEvent(SteppedEventConn(AG), SteppedEventOCGnd(), SteppedEventIntermediate(50)),
				want:    "1. Simultaneous Fault:\n     Bus Fault on:           4 TENNESSEE        132. kV 1LG Type=A",
				wantErr: nil,
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				if err := c.DoSteppedEvent(hnd, test.config); err != nil {
					t.Error(err)
				}
				if test.wantErr != err {
					t.Errorf("expected %v, got %v", test.wantErr, err)
				}
				fd := c.FaultDescription(1)
				if fd != test.want {
					t.Errorf("expected %q, got %q", test.want, fd)
					t.Logf("%q", fd)
				}
			})
		}
	})
}

func TestClient_GetData(t *testing.T) {
	c := NewClient()
	err := c.LoadDataFile(testCase)
	if err != nil {
		t.Fatal(err)
	}
	busHnd, err := c.FindBusByName("TENNESSEE", 132)
	if err != nil {
		t.Fatal(err)
	}
	xfmrs := c.NextEquipment(constants.TCXFMR)
	if !xfmrs.Next() {
		t.Fatal("could not find transformer")
	}
	xfmrHnd := xfmrs.Hnd()
	tests := []struct {
		name      string
		handle    int
		token     int
		wantValue interface{}
	}{
		{
			name:      "BUSsName",
			handle:    busHnd,
			token:     constants.BUSsName,
			wantValue: "TENNESSEE",
		},
		{
			name:      "BUSsLocation",
			handle:    busHnd,
			token:     constants.BUSsLocation,
			wantValue: "TENNESSE",
		},
		{
			name:      "BUSsComment",
			handle:    busHnd,
			token:     constants.BUSsComment,
			wantValue: "",
		},
		{
			name:      "BUSdKVnominal",
			handle:    busHnd,
			token:     constants.BUSdKVnominal,
			wantValue: 132.00,
		},
		{
			name:      "BUSdKVP",
			handle:    busHnd,
			token:     constants.BUSdKVP,
			wantValue: 0.00,
		},
		{
			name:      "BUSdSPCx",
			handle:    busHnd,
			token:     constants.BUSdSPCx,
			wantValue: 0.0,
		},
		{
			name:      "BUSdSPCy",
			handle:    busHnd,
			token:     constants.BUSdSPCy,
			wantValue: 0.0,
		},
		{
			name:      "BUSnNumber",
			handle:    busHnd,
			token:     constants.BUSnNumber,
			wantValue: 4,
		},
		{
			name:      "BUSnArea",
			handle:    busHnd,
			token:     constants.BUSnArea,
			wantValue: 1,
		},
		{
			name:      "BUSnZone",
			handle:    busHnd,
			token:     constants.BUSnZone,
			wantValue: 1,
		},
		{
			name:      "BUSnTapBus",
			handle:    busHnd,
			token:     constants.BUSnTapBus,
			wantValue: 0,
		},
		{
			name:      "BUSnSubGroup",
			handle:    busHnd,
			token:     constants.BUSnSubGroup,
			wantValue: 0,
		},
		{
			name:      "BUSnSlack",
			handle:    busHnd,
			token:     constants.BUSnSlack,
			wantValue: 0,
		},
		{
			name:      "BUSnVisible",
			handle:    busHnd,
			token:     constants.BUSnVisible,
			wantValue: 1,
		},
		{
			name:      "XRsName",
			handle:    xfmrHnd,
			token:     constants.XRsName,
			wantValue: "NV-NH",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := c.GetData(test.handle, test.token)
			var got interface{}
			switch test.wantValue.(type) {
			case string:
				var dest string
				err = data.Scan(&dest)
				got = dest
			case int:
				var dest int
				err = data.Scan(&dest)
				got = dest
			case float64:
				var dest float64
				err = data.Scan(&dest)
				got = dest
			default:
				t.Errorf("%T type not implemented", test.wantValue)
			}
			if err != nil {
				t.Error(err)
			}
			if got != test.wantValue {
				t.Errorf("expected %#v, got %#v", test.wantValue, got)
			}

		})
	}
}

func TestClient_NextRelay(t *testing.T) {
	c := NewClient()
	err := c.LoadDataFile(testCase)
	if err != nil {
		t.Fatal(err)
	}
	rlyGroups := c.NextEquipment(constants.TCRLYGroup)
	if !rlyGroups.Next() {
		t.Fatal("could not find relay group")
	}
	rlyGroupHnd := rlyGroups.Hnd()

	relays := c.NextRelay(rlyGroupHnd)
	for relays.Next() {
		t.Log(relays.Hnd())
	}
}

func TestClient_ObjTags(t *testing.T) {
	c := NewClient()
	err := c.LoadDataFile(testCase)
	if err != nil {
		t.Fatal(err)
	}
	busHnd, err := c.FindBusByName("TENNESSEE", 132)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("Invalid handle", func(t *testing.T) {
		_, err = c.GetObjTags(0)
		if err == nil {
			t.Errorf("expected 'GetObjTags failure: Invalid Device Handle', got %v", err)
		}
	})
	t.Run("Empty", func(t *testing.T) {
		tags, err := c.GetObjTags(busHnd)
		if err != nil {
			t.Error(err)
		}
		if len(tags) != 0 {
			t.Errorf("expected 0, got %d tags", len(tags))
		}
	})
	t.Run("Set", func(t *testing.T) {
		err := c.SetObjTags(busHnd, "ABCD", "EFG")
		if err != nil {
			t.Error(err)
		}
		tags, err := c.GetObjTags(busHnd)
		if err != nil {
			t.Error(err)
			t.Log(tags)
		}
		err = c.SetObjTags(busHnd, "HIJK", "LMNOP")
		if err != nil {
			t.Error(err)
		}
		tags, err = c.GetObjTags(busHnd)
		if err != nil {
			t.Error(err)
			t.Log(tags)
		}
		expectedLen := 2
		gotLen := len(tags)
		if gotLen != expectedLen {
			t.Fatalf("expected %d, got %d", expectedLen, gotLen)
			t.Log(tags)
		}
		expectedTag := "HIJK"
		gotTag := tags[0]
		if gotTag != expectedTag {
			t.Errorf("expected %q, got %q", expectedTag, gotTag)
			t.Log(tags)
		}
	})
	t.Run("Append", func(t *testing.T) {
		err := c.AppendObjTags(busHnd, "ABCD", "EFG")
		if err != nil {
			t.Error(err)
		}
		tags, err := c.GetObjTags(busHnd)
		if err != nil {
			t.Error(err)
		}
		gotLen := len(tags)
		expectedLen := 4
		if len(tags) != expectedLen {
			t.Fatalf("expected %d, got %d", expectedLen, gotLen)
			t.Log(tags)
		}
		expectedTag := "EFG"
		gotTag := tags[3]
		if gotTag != expectedTag {
			t.Errorf("expected %q, got %q", expectedTag, gotTag)
			t.Log(tags)
		}
	})
	t.Run("Replace", func(t *testing.T) {
		err := c.ReplaceObjTag(busHnd, "EFG", "Hello World")
		if err != nil {
			t.Error(err)
		}
		tags, err := c.GetObjTags(busHnd)
		if err != nil {
			t.Error(err)
		}
		gotLen := len(tags)
		expectedLen := 4
		if len(tags) != expectedLen {
			t.Fatalf("expected %d, got %d", expectedLen, gotLen)
			t.Log(tags)
		}
		expectedTag := "Hello World"
		gotTag := tags[3]
		if gotTag != expectedTag {
			t.Errorf("expected %q, got %q", expectedTag, gotTag)
			t.Log(tags)
		}
	})
}

func TestClient_ObjMemo(t *testing.T) {
	c := NewClient()
	err := c.LoadDataFile(testCase)
	if err != nil {
		t.Fatal(err)
	}
	busHnd, err := c.FindBusByName("TENNESSEE", 132)
	if err != nil {
		t.Fatal(err)
	}
	_ = busHnd
	t.Run("Get invalid handle", func(t *testing.T) {
		_, err = c.GetObjMemo(0)
		if err == nil {
			t.Errorf("expected 'GetObjMemo failure: Invalid Device Handle', got %v", err)
		}
	})
	t.Run("Get empty", func(t *testing.T) {
		s, err := c.GetObjMemo(busHnd)
		if err != nil {
			t.Error(err)
		}
		t.Logf("%q", s)
	})
	t.Run("Set invalid handle", func(t *testing.T) {
		err = c.SetObjMemo(0, "Hello World!")
		if err == nil {
			t.Errorf("expected 'SetObjMemo failure: Invalid Device Handle', got %v", err)
		}
	})
	t.Run("Set okay", func(t *testing.T) {
		err = c.SetObjMemo(busHnd, "Hello World!\nNew Line")
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("Get okay", func(t *testing.T) {
		s, err := c.GetObjMemo(busHnd)
		if err != nil {
			t.Error(err)
			t.Log(s)
		}
		expected := "Hello World!\nNew Line"
		if s != expected {
			t.Errorf("expected %q, got %q", expected, s)
		}
	})

}

// Examples

func ExampleData_Scan() {
	// Create API client.
	api := NewClient()

	// Load data file, and find bus handle.
	api.LoadDataFile(testCase)
	busHnd, err := api.FindBusByName("TENNESSEE", 132)
	if err != nil {
		return
	}

	// Get bus name and kv data.
	data := api.GetData(busHnd, constants.BUSsName, constants.BUSdKVnominal)

	// Scan loads the data into the pointers provided. Types must match the tokens provided.
	var name string
	var kV float64
	data.Scan(&name, &kV)

	fmt.Printf("Name: %s\n", name)
	fmt.Printf("kV: %0.2f\n", kV)

	// Output:
	// Name: TENNESSEE
	// kV: 132.00
}
