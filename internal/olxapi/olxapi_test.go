package olxapi

import (
	"testing"

	"github.com/readpe/goolx/constants"
)

var testCase = `C:\Program Files (x86)\ASPEN\1LPFv15\SAMPLE09.OLR`

func TestOlxAPI_GetOlrFilename(t *testing.T) {
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
	err = api.GetEquipment(constants.TCRLYGroup, &hnd)
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
			err = api.GetEquipment(constants.TCBus, &hnd)
			if err != nil {
				break
			}
			otgs, err := api.MakeOutageList(hnd, 9, 1)
			if err != nil {
				t.Error(err)
			}
			if otgs[len(otgs)-1] != 0 {
				t.Errorf("outage list not zero terminated")
				t.Log(hnd, otgs)
			}
		}

	})
}
