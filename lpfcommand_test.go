// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package goolx

import (
	"fmt"
	"os"
	"path"
	"testing"
)

func TestClient_Run1LPFCommand(t *testing.T) {
	tmpDir := os.TempDir()
	tmpFile := path.Join(tmpDir, "report.csv")
	api := NewClient()
	defer api.Release()

	if err := api.LoadDataFile(testCase); err != nil {
		t.Fatal(err)
	}

	xmlString := fmt.Sprintf(`<BUSFAULTSUMMARY REPORTPATHNAME="%s" BUSNOLIST="10,20,60"/>`, tmpFile)

	err := api.Run1LPFCommand(xmlString)
	if err != nil {
		t.Error(err)
	}
}
