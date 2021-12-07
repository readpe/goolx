package olxapi

import "testing"

var testCase = `C:\Program Files (x86)\ASPEN\1LPFv15\SAMPLE09.OLR`

func TestOlxAPI_GetOlrFilename(t *testing.T) {
	api := New()
	err := api.LoadDataFile(testCase)
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
	err := api.LoadDataFile(testCase)
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
	err := api.LoadDataFile(testCase)
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
	err := api.LoadDataFile(testCase)
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
