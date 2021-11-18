// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// olxapi.dll is a win32 application, build constrained to 386 GOARCH
//go:build windows && 386
// +build windows,386

package olxapi

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

// OlxAPIDLLPath is the full path to the directory containing the olxapi.dll.
// default is `C:\Programs Files (x86)\ASPEN\1LPFv14\OlxAPI`
// override if location is different.
var OlxAPIDLLPath = `C:\Programs Files (x86)\ASPEN\1LPFv14\OlxAPI`

// OlxAPI represents a connection to the olxapi.dll. Provides method
// wrappers for each api function. Instantiate using New().
//
// It is unclear if the olxapi.dll can be called cuncurrently if loaded into different processes,
// e.g. instantiating a new Client in a goroutine.
// TODO(readpe): Test concurrent access of olxapi.dll
type OlxAPI struct {
	sync.Mutex
	dll *syscall.DLL // olxapi.dll

	// OlxAPI Procedures
	errorString       *syscall.Proc
	versionInfo       *syscall.Proc
	saveDataFile      *syscall.Proc
	loadDataFile      *syscall.Proc
	readChangeFile    *syscall.Proc
	getEquipment      *syscall.Proc
	equipmentType     *syscall.Proc
	getData           *syscall.Proc
	findbus           *syscall.Proc
	getEquipmentByTag *syscall.Proc
	findBusNo         *syscall.Proc
	setData           *syscall.Proc
	getBusEquipment   *syscall.Proc
}

// New loads the dll and procedures and returns a new instance of OlxAPI.
// It is the callers responsibility to Release the dll when done with use.
// Recommend use of defer to ensure release of dll. Any errors will panic since
// no part of the API will work without loading the dll correctly.
//
// Current directory is temporarily changed to OlxAPIDLLPath prior to loading dll, and
// immediately changed back.
func New() *OlxAPI {

	// Temporarily change directory to OlxAPIDLLPath before loading dll. Defer changeback.
	changeBack, err := tempChdir(OlxAPIDLLPath)
	if err != nil {
		panic(err)
	}
	defer changeBack()

	api := &OlxAPI{
		dll: syscall.MustLoadDLL("olxapi.dll"),
	}

	// OlxApI Procedures, panics if not found
	api.errorString = api.dll.MustFindProc("OlxAPIErrorString")
	api.versionInfo = api.dll.MustFindProc("OlxAPIVersionInfo")
	api.saveDataFile = api.dll.MustFindProc("OlxAPISaveDataFile")
	api.loadDataFile = api.dll.MustFindProc("OlxAPILoadDataFile")
	api.readChangeFile = api.dll.MustFindProc("OlxAPIReadChangeFile")
	api.getEquipment = api.dll.MustFindProc("OlxAPIGetEquipment")
	api.equipmentType = api.dll.MustFindProc("OlxAPIEquipmentType")
	api.getData = api.dll.MustFindProc("OlxAPIGetData")
	api.findbus = api.dll.MustFindProc("OlxAPIFindBus")
	api.getEquipmentByTag = api.dll.MustFindProc("OlxAPIFindEquipmentByTag")
	api.findBusNo = api.dll.MustFindProc("OlxAPIFindBusNo")
	api.setData = api.dll.MustFindProc("OlxAPISetData")
	api.getBusEquipment = api.dll.MustFindProc("OlxAPIGetBusEquipment")

	return api
}

// Release releases the api dll. Must be called when done with use of dll.
func (o *OlxAPI) Release() error {
	o.Lock()
	defer o.Unlock()
	return o.dll.Release()
}

// ErrOlxAPI represents an OLXAPIFailure error returned by any
// olxapi function.
type ErrOlxAPI struct {
	function string
	err      string
}

func (e ErrOlxAPI) Error() string {
	return fmt.Sprintf("OLXAPIFailure: %s: %s", e.function, e.err)
}

// ErrorString calls the OlxAPIErrorString function, returning the string.
func (o *OlxAPI) ErrorString() string {
	o.Lock()
	r, _, _ := o.errorString.Call()
	o.Unlock()
	return utf8StringFromPtr(r)
}

// VersionInfo calls the OlxAPIVersionInfo function, returning the string.
func (o *OlxAPI) VersionInfo() string {
	buf := make([]byte, 1028)
	o.Lock()
	o.versionInfo.Call(uintptr(unsafe.Pointer(&buf[0])))
	o.Unlock()
	return string(buf)
}

// SaveDataFile calls the OlxAPISaveDataFile function. Returns error if
// OLXAPIFailure is returned.
func (o *OlxAPI) SaveDataFile(name string) error {
	b, err := utf8NullFromString(name)
	if err != nil {
		return fmt.Errorf("SaveDataFile: %v", err)
	}
	o.Lock()
	r, _, _ := o.saveDataFile.Call(uintptr(unsafe.Pointer(&b[0])))
	o.Unlock()
	if r == OLXAPIFailure {
		return ErrOlxAPI{"SaveDataFile", o.ErrorString()}
	}
	return nil
}

// LoadDataFile calls the OlxAPILoadDataFile function. Returns error if
// OLXAPIFailure is returned.
func (o *OlxAPI) LoadDataFile(name string) error {
	b, err := utf8NullFromString(name)
	if err != nil {
		return fmt.Errorf("LoadDataFile: %v", err)
	}
	o.Lock()
	r, _, _ := o.loadDataFile.Call(uintptr(unsafe.Pointer(&b[0])))
	o.Unlock()
	if r == OLXAPIFailure {
		return ErrOlxAPI{"LoadDataFile", o.ErrorString()}
	}
	return nil
}

// ReadChangeFile calls the OlxAPIReadChangeFile function. Returns error if
// OLXAPIFailure is returned.
func (o *OlxAPI) ReadChangeFile(name string) error {
	b, err := utf8NullFromString(name)
	if err != nil {
		return fmt.Errorf("ReadChangeFile: %v", err)
	}
	o.Lock()
	r, _, _ := o.readChangeFile.Call(uintptr(unsafe.Pointer(&b[0])))
	o.Unlock()
	if r == OLXAPIFailure {
		return ErrOlxAPI{"ReadChangeFile", o.ErrorString()}
	}
	return nil
}

// GetEquipment calls the OlxAPIGetEquipment function. Returns
// the equipment handle. Returns an error if OLXAPIFailure
// is returned. Returns io.EOF error when iteration is exhausted.
func (o *OlxAPI) GetEquipment(eqType int, hnd *int) error {
	o.Lock()
	r, _, _ := o.getEquipment.Call(uintptr(eqType), uintptr(unsafe.Pointer(&hnd)))
	o.Unlock()
	switch int(r) {
	case -1:
		// OlxAPI returns -1 when GetEquipment is exhausted, returning EOF error.
		return io.EOF
	case OLXAPIFailure:
		return ErrOlxAPI{"GetEquipment", o.ErrorString()}
	}
	return nil
}

// EquipmentType calls the OlxAPIEquipmentType function. Returns
// the equipment type code. Returns error if OLXAPIFailure
// is returned.
func (o *OlxAPI) EquipmentType(hnd int) (int, error) {
	o.Lock()
	r, _, _ := o.getEquipment.Call(uintptr(hnd))
	o.Unlock()
	if r == OLXAPIFailure {
		return 0, ErrOlxAPI{"EquipmentType", o.ErrorString()}
	}
	return int(r), nil
}

// GetData calls the OlxAPIGetData function for the given handle and token.
// The buffer must be adequate size for the data type being returned.
func (o *OlxAPI) GetData(hnd, token int, buf []byte) error {
	o.Lock()
	r, _, _ := o.getData.Call(uintptr(hnd), uintptr(token), uintptr(unsafe.Pointer(&buf[0])))
	o.Unlock()
	if r == OLXAPIFailure {
		return ErrOlxAPI{"GetDataFloat64", o.ErrorString()}
	}
	return nil
}

// FindBus calls the OlxAPIFindbus function.
func (o *OlxAPI) FindBus(name string, kv float64) (hnd int, err error) {
	b, err := utf8NullFromString(name)
	if err != nil {
		return hnd, fmt.Errorf("FindBus: %v", err)
	}
	// Cannot pass float64 by value as uintptr to 32bit dll using syscall directly.
	// Must convert to two uint32 and pass consecutively.
	// See https://github.com/golang/go/issues/29092
	f322 := float64ToUint32(kv)
	o.Lock()
	r, _, _ := o.findbus.Call(uintptr(unsafe.Pointer(&b[0])), uintptr(f322[0]), uintptr(f322[1]))
	o.Unlock()
	if r == OLXAPIFailure {
		return hnd, ErrOlxAPI{"FindBus", o.ErrorString()}
	}
	return int(r), nil
}

// FindEquipmentByTag calls the OlxAPIFindEquipmentByTag function.
func (o *OlxAPI) FindEquipmentByTag(eqType int, hnd *int, tags ...string) error {
	bTags, err := utf8NullFromString(strings.Join(tags, ","))
	if err != nil {
		return err
	}
	o.Lock()
	r, _, _ := o.getEquipmentByTag.Call(uintptr(unsafe.Pointer(&bTags[0])), uintptr(eqType), uintptr(unsafe.Pointer(&hnd)))
	o.Unlock()
	if r == OLXAPIFailure {
		return ErrOlxAPI{"FindEquipmentByTag", o.ErrorString()}
	}
	return nil
}

// FindBusNo calls the OlxAPIFindBusNo function.
func (o *OlxAPI) FindBusNo(n int) (int, error) {
	o.Lock()
	r, _, _ := o.findBusNo.Call(uintptr(n))
	o.Unlock()
	if r == OLXAPIFailure {
		return 0, ErrOlxAPI{"FundBusNo", o.ErrorString()}
	}
	return int(r), nil
}

// SetDataInt calls the OlxAPISetData function. Data provided is of type int.
func (o *OlxAPI) SetDataInt(hnd, token int, data interface{}) error {
	o.Lock()
	r, _, _ := o.setData.Call(uintptr(hnd), uintptr(token), uintptr(unsafe.Pointer(&data)))
	o.Unlock()
	if r == OLXAPIFailure {
		return ErrOlxAPI{"SetDataInt", o.ErrorString()}
	}
	return nil
}

// SetDataFloat64 calls the OlxAPISetData function. Data provided is of type int.
func (o *OlxAPI) SetDataFloat64(hnd, token, data float64) error {
	o.Lock()
	r, _, _ := o.setData.Call(uintptr(hnd), uintptr(token), uintptr(unsafe.Pointer(&data)))
	o.Unlock()
	if r == OLXAPIFailure {
		return ErrOlxAPI{"SetDataInt", o.ErrorString()}
	}
	return nil
}

// GetBusEquipment returns the handle of the next equipment attached to the provided bus handle,
// of the specified type. Returns io.EOF error when iteration is exhausted.
func (o *OlxAPI) GetBusEquipment(busHnd, eqType int, hnd *int) error {
	o.Lock()
	r, _, _ := o.getBusEquipment.Call(uintptr(busHnd), uintptr(eqType), uintptr(unsafe.Pointer(hnd)))
	o.Unlock()

	switch int(r) {
	case -1:
		// OlxAPI returns -1 when GetBusEquipment is exhausted, returning EOF error.
		return io.EOF
	case OLXAPIFailure:
		return ErrOlxAPI{"GetBusEquipment", o.ErrorString()}
	}
	return nil
}
