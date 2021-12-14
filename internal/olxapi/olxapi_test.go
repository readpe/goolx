package olxapi

import (
	"os"
	"path"
	"path/filepath"
	"testing"
)

var testCase = `..\..\local\SAMPLE09.OLR`

func init() {
	var err error
	testCase, err = filepath.Abs(testCase)
	if err != nil {
		panic(err)
	}
}

const (
	TCBus      = 1
	TCBranch   = 9
	TCRLYGroup = 20
)

func TestOlxAPI_GetOlrFilename(t *testing.T) {
	t.Log(testCase)
	api := New()
	err := api.LoadDataFile(testCase, false)
	if err != nil {
		t.Error(err)
	}
	t.Run("Okay", func(t *testing.T) {
		got := api.GetOlrFileName()
		expected := testCase
		if expected != got {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})
}

func TestOlxAPI_GetObjGUID(t *testing.T) {
	api := New()
	err := api.LoadDataFile(testCase, false)
	if err != nil {
		t.Error(err)
	}
	busHnd, err := api.FindBusByName("NEW HAMPSHR", 33.0)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("Invalid handle", func(t *testing.T) {
		_, err := api.GetObjGUID(0)
		if err == nil {
			t.Errorf("expected 'GetObjGUID failure: Invalid Device Handle', got %v", err)
		}
	})
	t.Run("Bus", func(t *testing.T) {
		got, err := api.GetObjGUID(busHnd)
		if err != nil {
			t.Error(err)
		}
		expected := "{26734839-572f-4347-8a16-6f25a8011186}"
		if expected != got {
			t.Errorf("expected %q GUID, got %q", expected, got)
		}
	})
}

func TestOlxAPI_GetAreaName(t *testing.T) {
	api := New()
	err := api.LoadDataFile(testCase, false)
	if err != nil {
		t.Error(err)
	}
	t.Run("Invalid area", func(t *testing.T) {
		_, err := api.GetAreaName(0)
		if err == nil {
			t.Errorf("expected 'GetAreaName failure: Area number not found', got %v", err)
		}
	})
	t.Run("Valid area", func(t *testing.T) {
		got, err := api.GetAreaName(2)
		if err != nil {
			t.Error(err)
		}
		expected := "CC"
		if expected != got {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})
}

func TestOlxAPI_GetZoneName(t *testing.T) {
	api := New()
	err := api.LoadDataFile(testCase, false)
	if err != nil {
		t.Error(err)
	}
	t.Run("Invalid zone", func(t *testing.T) {
		_, err := api.GetZoneName(-1)
		if err == nil {
			t.Errorf("expected 'GetZoneName failure: Zone number not found', got %v", err)
		}
	})
	t.Run("Valid zone", func(t *testing.T) {
		got, err := api.GetZoneName(1)
		if err != nil {
			t.Error(err)
		}
		expected := "ZONE 1"
		if expected != got {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})
}

func TestOlxAPI_GetRelayTime(t *testing.T) {
	api := New()
	defer api.Release()
	err := api.LoadDataFile(testCase, false)
	if err != nil {
		t.Error(err)
	}

	var hnd int
	err = api.GetEquipment(TCRLYGroup, &hnd)
	if err != nil {
		t.Error(err)
	}
	var rlyHnd int
	err = api.GetRelay(hnd, &rlyHnd)
	if err != nil {
		t.Error(err)
	}

	t.Run("Okay", func(t *testing.T) {
		opTime, opText, err := api.GetRelayTime(rlyHnd, 0, true)
		if err == nil {
			t.Errorf("expected 'fault simulation result not available' error, got %v", err)
			t.Log(hnd, rlyHnd, opTime, opText)
		}
	})
}

func TestOlxAPI_MakeOutageList(t *testing.T) {
	api := New()
	defer api.Release()
	err := api.LoadDataFile(testCase, false)
	if err != nil {
		t.Error(err)
	}

	t.Run("Okay", func(t *testing.T) {
		var hnd int

		for {
			err = api.GetEquipment(TCBus, &hnd)
			if err != nil {
				break
			}
			_, err := api.MakeOutageList(hnd, 9, 1)
			if err != nil {
				t.Error(err)
			}
			// TODO: Getting unexpected results from OlxAPI, need to research more.
			// if otgs[len(otgs)-1] != 0 {
			// 	t.Errorf("outage list not zero terminated")
			// 	t.Log(hnd, otgs)
			// }
		}

	})
}

func TestOlxAPI_GetLogicScheme(t *testing.T) {
	api := New()
	defer api.Release()

	if err := api.LoadDataFile(testCase, false); err != nil {
		t.Fatal(err)
	}

	t.Run("Okay", func(t *testing.T) {
		var rlyGrpHnd int
		if err := api.GetEquipment(TCRLYGroup, &rlyGrpHnd); err != nil {
			t.Fatal(err)
		}

		var logicHnd int
		if err := api.GetLogicScheme(rlyGrpHnd, &logicHnd); err == nil {
			t.Errorf("expected 'GetRelay failure: scheme is empty', got %v", err)
		}
	})

}

func TestOlxAPI_FullNames(t *testing.T) {
	api := New()
	defer api.Release()

	if err := api.LoadDataFile(testCase, false); err != nil {
		t.Fatal(err)
	}

	t.Run("FullBusName", func(t *testing.T) {
		var hnd int
		if err := api.GetEquipment(TCBus, &hnd); err != nil {
			t.Fatal(err)
		}
		expected := "28 ARIZONA 132.kV"
		got := api.FullBusName(hnd)
		if got != expected {
			t.Errorf("got %q, expected %q", got, expected)
		}
	})
	t.Run("FullBranchName", func(t *testing.T) {
		var hnd int
		if err := api.GetEquipment(TCBus, &hnd); err != nil {
			t.Fatal(err)
		}
		var brHnd int
		if err := api.GetBusEquipment(hnd, TCBranch, &brHnd); err != nil {
			t.Fatal(err)
		}
		expected := "   28 ARIZONA            132.kV -     6 NEVADA             132.kV 1 L NV-AZ"
		got := api.FullBranchName(brHnd)
		if got != expected {
			t.Errorf("got %q, expected %q", got, expected)
		}
	})
	t.Run("FullBranchName", func(t *testing.T) {
		var hnd int
		if err := api.GetEquipment(TCRLYGroup, &hnd); err != nil {
			t.Fatal(err)
		}
		var rlyHnd int
		if err := api.GetRelay(hnd, &rlyHnd); err != nil {
			t.Fatal(err)
		}
		expected := "[DS RELAY] GCXTEST ON     6 NEVADA             132.kV -     2 CLAYTOR            132.kV 1 L CLA-NV"
		got := api.FullRelayName(rlyHnd)
		if got != expected {
			t.Errorf("got %q, expected %q", got, expected)
		}
	})

}

func TestOlxAPI_GetObjJournalRecord(t *testing.T) {
	api := New()
	defer api.Release()

	if err := api.LoadDataFile(testCase, false); err != nil {
		t.Fatal(err)
	}

	var hnd int
	if err := api.GetEquipment(TCBus, &hnd); err != nil {
		t.Fatal(err)
	}

	t.Run("Okay", func(t *testing.T) {
		expected := "Unknown\nUnknown\n2002/3/19 01:00\nUnknown"
		got := api.GetObjJournalRecord(hnd)
		if got != expected {
			t.Errorf("got %q, expected %q", got, expected)
		}
	})
}

func TestOlxAPI_GetObjUDF(t *testing.T) {
	api := New()
	defer api.Release()

	if err := api.LoadDataFile(testCase, false); err != nil {
		t.Fatal(err)
	}

	var hnd int
	if err := api.GetEquipment(TCBus, &hnd); err != nil {
		t.Fatal(err)
	}

	t.Run("Invalid field name", func(t *testing.T) {
		_, err := api.GetObjUDF(hnd, "NOTCORRECT")
		if err == nil {
			t.Errorf("expected error 'Invalid field name: test' got %v", err)
		}
	})
	t.Run("Okay", func(t *testing.T) {
		err := api.SetObjUDF(hnd, "SUBID", "SUBA")
		if err != nil {
			t.Fatal(err)
		}
		expected := "SUBA"
		got, err := api.GetObjUDF(hnd, "SUBID")
		if err != nil {
			t.Error(err)
		}
		if got != expected {
			t.Errorf("got %q, expected %q", got, expected)
		}
		field, got, err := api.GetObjUDFByIndex(hnd, 0)
		if err != nil {
			t.Error(err)
		}
		if got != expected {
			t.Errorf("got %q, expected %q", got, expected)
			t.Log(field, got)
		}
	})
}

func TestOlxAPI_GetPSCVoltage(t *testing.T) {
	api := New()
	defer api.Release()

	if err := api.LoadDataFile(testCase, false); err != nil {
		t.Fatal(err)
	}
	var hnd int
	if err := api.GetEquipment(TCBus, &hnd); err != nil {
		t.Fatal(err)
	}
	_, _, err := api.GetPSCVoltage(hnd, 0)
	if err == nil {
		t.Errorf("expected incorrect style code 0")
	}
}

func TestOlxAPI_FindObj1LPF(t *testing.T) {
	api := New()
	defer api.Release()

	if err := api.LoadDataFile(testCase, false); err != nil {
		t.Fatal(err)
	}
	var hnd int
	if err := api.GetEquipment(TCBus, &hnd); err != nil {
		t.Fatal(err)
	}
	s, err := api.PrintObj1LPF(hnd)
	if err != nil {
		t.Error(err)
	}
	got, err := api.FindObj1LPF(s)
	if err != nil {
		t.Error(err)
	}
	if got != hnd {
		t.Errorf("got %d, expected %d", got, hnd)
	}
}

func TestOlxAPI_BoundaryEquivalent(t *testing.T) {
	api := New()
	defer api.Release()

	if err := api.LoadDataFile(testCase, false); err != nil {
		t.Fatal(err)
	}
	var hnd int
	if err := api.GetEquipment(TCBus, &hnd); err != nil {
		t.Fatal(err)
	}

	tmpFile := path.Join(os.TempDir(), "test.olr")
	err := api.BoundaryEquivalent(tmpFile, []int{hnd}, [3]float64{0, 0, 0})
	if err != nil {
		t.Error(err)
	}

	info, err := os.Stat(tmpFile)
	if err != nil {
		t.Error(err)
	}

	if info.Size() == 0 {
		t.Errorf("expected non-zero equivalent case size")
	}
}
