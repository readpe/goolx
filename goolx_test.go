// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package goolx

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
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
	err = c.olxAPI.GetEquipment(TCBus, &hnd)
	if err != nil {
		t.Error(err)
	}
	t.Log(err, hnd)

}
func TestNextEquipment(t *testing.T) {
	c := NewClient()
	defer c.Release()
	c.LoadDataFile(testCase)
	hi := c.NextEquipment(TCBus)
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

// Examples

func ExampleData_Scan() {
	var busHnd int
	api := NewClient()
	data := api.GetData(busHnd, BUSsName, BUSdKVP)

	// Scan loads the data into the pointers provided populating the Bus structure in this example.
	bus := Bus{}
	data.Scan(&bus.Name, &bus.KVNominal)
}
