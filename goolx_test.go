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

func TestGetEquipmentType(t *testing.T) {
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
	eqType, err := c.EquipmentType(hnd)
	if err != nil {
		t.Error(err)
		t.Log(eqType, err)
	}
	if eqType != TCBus {
		t.Errorf("expected eqType %d, got %d", TCBus, eqType)
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

func TestNextBusEquipment(t *testing.T) {
	c := NewClient()
	defer c.Release()
	c.LoadDataFile(testCase)
	hi := c.NextEquipment(TCBus)
	var handles []int
	var branches []int
	for hi.Next() {
		handles = append(handles, hi.Hnd())
		brs := c.NextBusEquipment(hi.Hnd(), TCBranch)
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

func TestClient_GetSteppedEvent(t *testing.T) {
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
	t.Run("Okay", func(t *testing.T) {
		err := c.DoSteppedEvent(busHnd, NewSteppedEvent(
			SteppedEventConn(ABC),
			SteppedEventCloseIn(),
			SteppedEventAll(),
		))
		if err != nil {
			t.Error(err)
		}
		se, err := c.GetSteppedEvent(2)
		if err != nil {
			t.Error(err)
		}
		expected := `0.12, 3872.7, false, "Event no. 1 at time= 0.124s\n7 OHIO 132.kV-6 NEVADA 132.kV 1L tripped by OC phase relay OH-P1\n", "Bus Fault on:           4 TENNESSEE        132. kV 3LG  "`
		got := fmt.Sprintf("%0.2f, %0.1f, %t, %q, %q", se.Time, se.Current, se.UserEvent, se.EventDescription, se.FaultDescription)
		if expected != got {
			t.Errorf("expected %s\ngot %s", expected, got)
		}
		var gotSE []SteppedEvent
		steppedEvents := c.NextSteppedEvent()
		for steppedEvents.Next() {
			gotSE = append(gotSE, steppedEvents.Data())
		}
		expectedLen := 4
		gotLen := len(gotSE)
		if expectedLen != gotLen {
			t.Errorf("expected %d steps, got, %d", expectedLen, gotLen)
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
	xfmrs := c.NextEquipment(TCXFMR)
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
			token:     BUSsName,
			wantValue: "TENNESSEE",
		},
		{
			name:      "BUSsLocation",
			handle:    busHnd,
			token:     BUSsLocation,
			wantValue: "TENNESSE",
		},
		{
			name:      "BUSsComment",
			handle:    busHnd,
			token:     BUSsComment,
			wantValue: "",
		},
		{
			name:      "BUSdKVnominal",
			handle:    busHnd,
			token:     BUSdKVnominal,
			wantValue: 132.00,
		},
		{
			name:      "BUSdKVP",
			handle:    busHnd,
			token:     BUSdKVP,
			wantValue: 0.00,
		},
		{
			name:      "BUSdSPCx",
			handle:    busHnd,
			token:     BUSdSPCx,
			wantValue: 0.0,
		},
		{
			name:      "BUSdSPCy",
			handle:    busHnd,
			token:     BUSdSPCy,
			wantValue: 0.0,
		},
		{
			name:      "BUSnNumber",
			handle:    busHnd,
			token:     BUSnNumber,
			wantValue: 4,
		},
		{
			name:      "BUSnArea",
			handle:    busHnd,
			token:     BUSnArea,
			wantValue: 1,
		},
		{
			name:      "BUSnZone",
			handle:    busHnd,
			token:     BUSnZone,
			wantValue: 1,
		},
		{
			name:      "BUSnTapBus",
			handle:    busHnd,
			token:     BUSnTapBus,
			wantValue: 0,
		},
		{
			name:      "BUSnSubGroup",
			handle:    busHnd,
			token:     BUSnSubGroup,
			wantValue: 0,
		},
		{
			name:      "BUSnSlack",
			handle:    busHnd,
			token:     BUSnSlack,
			wantValue: 0,
		},
		{
			name:      "BUSnVisible",
			handle:    busHnd,
			token:     BUSnVisible,
			wantValue: 1,
		},
		{
			name:      "XRsName",
			handle:    xfmrHnd,
			token:     XRsName,
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
	rlyGroups := c.NextEquipment(TCRLYGroup)
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
		_, err = c.TagsGet(0)
		if err == nil {
			t.Errorf("expected 'TagsGet failure: Invalid Device Handle', got %v", err)
		}
	})
	t.Run("Empty", func(t *testing.T) {
		tags, err := c.TagsGet(busHnd)
		if err != nil {
			t.Error(err)
		}
		if len(tags) != 0 {
			t.Errorf("expected 0, got %d tags", len(tags))
		}
	})
	t.Run("Set", func(t *testing.T) {
		err := c.TagsSet(busHnd, "ABCD", "EFG")
		if err != nil {
			t.Error(err)
		}
		tags, err := c.TagsGet(busHnd)
		if err != nil {
			t.Error(err)
			t.Log(tags)
		}
		err = c.TagsSet(busHnd, "HIJK", "LMNOP")
		if err != nil {
			t.Error(err)
		}
		tags, err = c.TagsGet(busHnd)
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
		err := c.TagsAppend(busHnd, "ABCD", "EFG")
		if err != nil {
			t.Error(err)
		}
		tags, err := c.TagsGet(busHnd)
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
		err := c.TagReplace(busHnd, "EFG", "Hello World")
		if err != nil {
			t.Error(err)
		}
		tags, err := c.TagsGet(busHnd)
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
		_, err = c.MemoGet(0)
		if err == nil {
			t.Errorf("expected 'MemoGet failure: Invalid Device Handle', got %v", err)
		}
	})
	t.Run("Get empty", func(t *testing.T) {
		s, err := c.MemoGet(busHnd)
		if err != nil {
			t.Error(err)
		}
		t.Logf("%q", s)
	})
	t.Run("Set invalid handle", func(t *testing.T) {
		err = c.MemoSet(0, "Hello World!")
		if err == nil {
			t.Errorf("expected 'MemoSet failure: Invalid Device Handle', got %v", err)
		}
	})
	t.Run("Set okay", func(t *testing.T) {
		err = c.MemoSet(busHnd, "Hello World!\nNew Line")
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("Get okay", func(t *testing.T) {
		s, err := c.MemoGet(busHnd)
		if err != nil {
			t.Error(err)
			t.Log(s)
		}
		expected := "Hello World!\nNew Line"
		if s != expected {
			t.Errorf("expected %q, got %q", expected, s)
		}
	})

	t.Run("Contains", func(t *testing.T) {
		if ok := c.MemoContains(busHnd, "World"); !ok {
			t.Errorf("expected contains World true, got false")
			t.Log(c.MemoGet(busHnd))
		}
		if ok := c.MemoContains(busHnd, "Universe"); ok {
			t.Errorf("expected contains Universe false, got true")
			t.Log(c.MemoGet(busHnd))
		}
	})
	t.Run("ReplaceAll", func(t *testing.T) {
		err := c.MemoReplaceAll(busHnd, "World", "Universe")
		if err != nil {
			t.Error(err)
		}
		if ok := c.MemoContains(busHnd, "Universe"); !ok {
			t.Errorf("expected contains Universe true, got false")
			t.Log(c.MemoGet(busHnd))
		}
	})
}

func TestClient_GetSCVoltage(t *testing.T) {
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
	t.Run("Invalid handle", func(t *testing.T) {
		_, _, _, err := c.GetSCVoltagePhase(0)
		if err == nil {
			t.Errorf("expected 'GetVoltage Failure: Invalid Device Handle', got %v", err)
		}
	})
	t.Run("No Fault", func(t *testing.T) {
		err = c.PickFault(SFFirst, 1)
		if err == nil {
			t.Errorf("expected 'PickFault: fault not simulated', got %v", err)
		}
		_, _, _, err := c.GetSCVoltagePhase(busHnd)
		if err == nil {
			t.Errorf("expected 'GetSCVoltage: fault not simulated', got %v", err)
		}
	})
	t.Run("1LG", func(t *testing.T) {
		err := c.DoFault(busHnd, NewFaultConfig(FaultConn(AG), FaultCloseIn()))
		if err != nil {
			t.Fatal(err)
		}
		err = c.PickFault(SFFirst, 1)
		if err != nil {
			t.Fatal(err)
		}
		va, vb, vc, err := c.GetSCVoltagePhase(busHnd)
		if err != nil {
			t.Error(err)
		}
		got := fmt.Sprint(va)
		expected := "0.00∠0.0°"
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
		got = fmt.Sprint(vb)
		expected = "82.91∠-125.0°"
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
		got = fmt.Sprint(vc)
		expected = "81.78∠128.8°"
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}

		v0, v1, v2, err := c.GetSCVoltageSeq(busHnd)
		if err != nil {
			t.Error(err)
		}
		got = fmt.Sprint(v0)
		expected = "32.96∠-177.6°"
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
		got = fmt.Sprint(v1)
		expected = "54.50∠1.8°"
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
		got = fmt.Sprint(v2)
		expected = "21.54∠-179.1°"
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})
}

func TestClient_GetSCCurrent(t *testing.T) {
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
	t.Run("Invalid handle", func(t *testing.T) {
		_, _, _, err := c.GetSCCurrentPhase(0)
		if err == nil {
			t.Errorf("expected 'GetVoltage Failure: Invalid Device Handle', got %v", err)
		}
	})
	t.Run("No Fault", func(t *testing.T) {
		err = c.PickFault(SFFirst, 1)
		if err == nil {
			t.Errorf("expected 'PickFault: fault not simulated', got %v", err)
		}
		_, _, _, err := c.GetSCCurrentPhase(HNDSC)
		if err == nil {
			t.Errorf("expected 'GetSCCurrent: fault not simulated', got %v", err)
		}
	})
	t.Run("1LG", func(t *testing.T) {
		err := c.DoFault(busHnd, NewFaultConfig(FaultConn(AG), FaultCloseIn()))
		if err != nil {
			t.Fatal(err)
		}
		err = c.PickFault(SFFirst, 1)
		if err != nil {
			t.Fatal(err)
		}
		ia, ib, ic, err := c.GetSCCurrentPhase(HNDSC)
		if err != nil {
			t.Error(err)
		}
		got := fmt.Sprint(ia)
		expected := "3690.63∠-79.4°"
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
		got = fmt.Sprint(ib)
		expected = "0.00∠0.0°"
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
		got = fmt.Sprint(ic)
		expected = "0.00∠0.0°"
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}

		i0, i1, i2, err := c.GetSCCurrentSeq(HNDSC)
		if err != nil {
			t.Error(err)
		}
		got = fmt.Sprint(i0)
		expected = "1230.21∠-79.4°"
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
		got = fmt.Sprint(i1)
		expected = "1230.21∠-79.4°"
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
		got = fmt.Sprint(i2)
		expected = "1230.21∠-79.4°"
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})
}

func TestClient_NextFault(t *testing.T) {
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
	t.Run("No Fault", func(t *testing.T) {
		faults := c.NextFault(5)
		for faults.Next() {
			t.Log(faults.Index(), c.FaultDescription(faults.Index()))
		}
	})
	t.Run("Faults", func(t *testing.T) {
		err := c.DoFault(busHnd, NewFaultConfig(FaultConn(AG), FaultConn(ABC), FaultConn(ABG), FaultCloseIn()))
		if err != nil {
			t.Fatal(err)
		}
		var got []int
		faults := c.NextFault(5)
		for faults.Next() {
			got = append(got, faults.Index())
		}
		expectedLen := 3
		gotLen := len(got)
		if expectedLen != gotLen {
			t.Errorf("expected %d, got %d faults", expectedLen, gotLen)
		}
	})
}

func TestClient_SetData(t *testing.T) {
	c := NewClient()
	defer c.Release()
	err := c.LoadDataFile(testCase)
	if err != nil {
		t.Fatal(err)
	}
	busHnd, err := c.FindBusByName("TENNESSEE", 132)
	if err != nil {
		t.Fatal(err)
	}
	_ = busHnd
	t.Run("string", func(t *testing.T) {
		expected := "TESTING"
		err := c.SetData(busHnd, BUSsName, expected)
		if err != nil {
			t.Error(err)
		}

		err = c.PostData(busHnd)
		if err != nil {
			t.Error(err)
		}

		var got string
		if err := c.GetData(busHnd, BUSsName).Scan(&got); err != nil {
			t.Error(err)
		}

		if got != expected {
			t.Errorf("expected %s, got %s", expected, got)
		}
	})
	t.Run("float64", func(t *testing.T) {
		expected := 45.0
		err := c.SetData(busHnd, BUSdSPCx, expected)
		if err != nil {
			t.Error(err)
		}

		err = c.PostData(busHnd)
		if err != nil {
			t.Error(err)
		}

		var got float64
		if err := c.GetData(busHnd, BUSdSPCx).Scan(&got); err != nil {
			t.Error(err)
		}

		if got != expected {
			t.Errorf("expected %f, got %f", expected, got)
		}
	})
	t.Run("int", func(t *testing.T) {
		expected := 10
		err := c.SetData(busHnd, BUSnArea, expected)
		if err != nil {
			t.Error(err)
		}

		err = c.PostData(busHnd)
		if err != nil {
			t.Error(err)
		}

		var got int
		if err := c.GetData(busHnd, BUSnArea).Scan(&got); err != nil {
			t.Error(err)
		}

		if got != expected {
			t.Errorf("expected %d, got %d", expected, got)
		}
	})
}

func TestClient_MakeOutageList(t *testing.T) {
	api := NewClient()
	defer api.Release()

	err := api.LoadDataFile(testCase)
	if err != nil {
		t.Fatal(err)
	}

	hnd, err := api.FindBusByName("NEVADA", 132.0)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("Okay", func(t *testing.T) {
		otgs, err := api.MakeOutageList(hnd, 0, OtgLine)
		if err != nil {
			t.Error(err)
		}
		expected := 4 + 1
		got := len(otgs)
		if expected != got {
			t.Errorf("expected %d, got %d outages", expected, got)
			t.Log(otgs)
		}

		otgs, err = api.MakeOutageList(hnd, 0, OtgLine|OtgPhaseShift)
		if err != nil {
			t.Error(err)
		}
		expected = 5 + 1
		got = len(otgs)
		if expected != got {
			t.Errorf("expected %d, got %d outages", expected, got)
			t.Log(otgs)
		}

		otgs, err = api.MakeOutageList(hnd, 0, OtgLine|OtgPhaseShift|OtgXfmr)
		if err != nil {
			t.Error(err)
		}
		expected = 6 + 1
		got = len(otgs)
		if expected != got {
			t.Errorf("expected %d, got %d outages", expected, got)
			t.Log(otgs)
		}

		otgs, err = api.MakeOutageList(hnd, 0, OtgLine|OtgPhaseShift|OtgXfmr|OtgXfmr3)
		if err != nil {
			t.Error(err)
		}
		expected = 7 + 1
		got = len(otgs)
		if expected != got {
			t.Errorf("expected %d, got %d outages", expected, got)
			t.Log(otgs)
		}
	})
	t.Run("Run", func(t *testing.T) {
		t.Skip("Unexpected return from MakeOutageList in this test")
		otgs, err := api.MakeOutageList(hnd, 0, OtgLine)
		if err != nil {
			t.Error(err)
		}
		cfg := NewFaultConfig(
			FaultCloseInOutage(otgs, OutageOptionOnePer),
			FaultClearPrev(true),
			FaultConn(ABC),
		)
		err = api.DoFault(hnd, cfg)
		if err != nil {
			t.Error(err)
		}
		expected := 4
		var got int
		flts := api.NextFault(5)
		for flts.Next() {
			got++
		}
		if expected != got {
			t.Errorf("expected %d, got %d outages", expected, got)
		}
	})
}

func TestClient_GetObjGUID(t *testing.T) {
	api := NewClient()
	defer api.Release()

	if err := api.LoadDataFile(testCase); err != nil {
		t.Error(err)
	}

	hnd, err := api.FindBusByName("NEVADA", 132.0)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Okay", func(t *testing.T) {
		got, err := api.GetGUID(hnd)
		if err != nil {
			t.Error(err)
		}
		expected := "{ad5860b5-f146-4dd5-9a11-5aadf06d907b}"
		if expected != got {
			t.Errorf("expected %q, got %q", expected, got)
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
	data := api.GetData(busHnd, BUSsName, BUSdKVnominal)

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

func TestClient_GetRelayTime(t *testing.T) {
	api := NewClient()
	defer api.Release()

	err := api.LoadDataFile(testCase)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Okay", func(t *testing.T) {
		rlyGroups := api.NextEquipment(TCRLYGroup)
		for rlyGroups.Next() {
			rgHnd := rlyGroups.Hnd()

			err := api.DoFault(rgHnd, NewFaultConfig(FaultConn(AG), FaultLineEnd(), FaultClearPrev(true)))
			if err != nil {
				t.Error(err)
			}

			relays := api.NextRelay(rgHnd)
			for relays.Next() {
				rlyHnd := relays.Hnd()
				var rid string
				if err := api.GetData(rlyHnd, RDsID).Scan(&rid); err != nil {
					t.Error(err)
				}
				faults := api.NextFault(5)
				for faults.Next() {
					idx := faults.Index()
					fd := api.FaultDescription(idx)
					opTime, opText, err := api.GetRelayTime(rlyHnd, 1, true)
					if err != nil {
						t.Error(err)
						t.Log(rgHnd, rlyHnd, opTime, opText)
					}
					expected_rid := "CL-P1"
					expected := "TOC=1558.29"
					if rid == expected_rid && opText != expected {
						t.Errorf("relay %q expected %q, got %q", rid, expected, opText)
						t.Log(fd, rid, opTime, opText)
					}
					expected_rid = "OH-G1"
					expected = "TOC=1513.43"
					if rid == expected_rid && opText != expected {
						t.Errorf("relay %q expected %q, got %q", rid, expected, opText)
						t.Log(fd, rid, opTime, opText)
					}
					expected_rid = "Clator_NV G1"
					expected = "ZG2"
					if rid == expected_rid && opText != expected {
						t.Errorf("relay %q expected %q, got %q", rid, expected, opText)
						t.Log(fd, rid, opTime, opText)
					}
				}
			}
		}
	})

}

func TestClient_Nextlogicscheme(t *testing.T) {
	api := NewClient()
	defer api.Release()

	if err := api.LoadDataFile(`C:\Users\rpe\Desktop\SAMPLE09.OLR`); err != nil {
		t.Fatal(err)
	}

	for rg := api.NextEquipment(TCRLYGroup); rg.Next(); {
		for l := api.NextLogicScheme(rg.Hnd()); l.Next(); {
			var lsID string
			if err := api.GetData(l.Hnd(), LSsID).Scan(&lsID); err != nil {
				t.Error(err)
			}
		}
	}
}
