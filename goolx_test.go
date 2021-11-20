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

func TestInfo(t *testing.T) {
	c := NewClient()
	got := c.Info()
	if got == "" {
		t.Errorf("info string is empty")
	}
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
}

func TestSaveDatafile(t *testing.T) {
	c := NewClient()
	tmp, err := ioutil.TempDir("", "goolx")
	if err != nil {
		t.Error(tmp)
	}
	err = c.SaveDataFile(path.Join(tmp, "test.olr"))
	if err != nil {
		t.Error(err)
	}
	err = c.SaveDataFile(path.Join(tmp, "temp", "test.olr"))
	if err == nil {
		t.Errorf("expected directory doesn't exist error, got nil")
	}
}

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
	err = os.WriteFile(path.Join(tmp, "test.olr"), b, 0700)
	if err == nil {
		t.Errorf("expected malformed case error, got nil")
	}
	// TODO(readpe): write embedded olr case to temp directory for successful loading test
}

func TestReadChangeFile(t *testing.T) {
	c := NewClient()
	tmp, err := ioutil.TempDir("", "goolx")
	if err != nil {
		t.Error(tmp)
	}
	err = c.ReadChangeFile(path.Join(tmp, "test.chf"))
	if err == nil {
		t.Errorf("expected file doesn't exist error, got nil")
	}
	// testing empty changefile
	var b []byte
	err = os.WriteFile(path.Join(tmp, "test.chf"), b, 0700)
	if err != nil {
		t.Error(err)
	}
	err = c.ReadChangeFile(path.Join(tmp, "test.chf"))
	if err == nil {
		t.Errorf("expected malformed changefile error, got nil")
	}
}

func TestNextEquipment(t *testing.T) {
	c := NewClient()
	defer c.Release()
	hi := c.NextEquipment(TCBus)
	for hi.Next() {
		hnd := hi.Hnd()
		if hnd <= 0 {
			t.Errorf("expected postive hnd, got %d", hnd)
		}
	}
}

func TestFindEquipmentByTag(t *testing.T) {
	c := NewClient()
	defer c.Release()
	hi := c.NextEquipmentByTag(TCBus, "Tag1", "Tag2", "Tag3")
	for hi.Next() {
		hnd := hi.Hnd()
		if hnd <= 0 {
			t.Errorf("expected postive hnd, got %d", hnd)
		}
	}
}

func TestDoFault(t *testing.T) {
	c := NewClient()
	defer c.Release()
	hnd, err := c.FindBus("TEMP", 115)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("Okay", func(t *testing.T) {
		config := NewFaultConfig()
		if err := c.DoFault(hnd, config); err != nil {
			t.Error(err)
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
				config: NewFaultConfig(FaultConn(ABC), FaultCloseIn()),
				want:   "Close-in 3LG fault description here",
			},
			{
				name:   "1LG, Close-in",
				config: NewFaultConfig(FaultConn(AG), FaultCloseIn()),
				want:   "Close-in 1LG fault description here",
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				if err := c.DoFault(hnd, test.config); err != nil {
					t.Error(err)
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
