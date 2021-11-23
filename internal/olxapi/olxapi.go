// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// olxapi.dll is a win32 application, build constrained to 386 GOARCH
//go:build windows && 386
// +build windows,386

package olxapi

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

// OlxAPI return codes.
const (
	OLXAPIFailure = 0
	OLXAPIOk      = 1
)

// OlxAPIDLLPath is the full path to the directory containing the olxapi.dll.
// default is `C:\Program Files (x86)\ASPEN\1LPFv15`
// override if location is different.
var OlxAPIDLLPath = `C:\Program Files (x86)\ASPEN\1LPFv15`

// OlxAPI represents a connection to the olxapi.dll. Provides method
// wrappers for each api function. Instantiate using New().
//
// It is unclear if the olxapi.dll can be called cuncurrently if loaded into different processes,
// e.g. instantiating a new Client in a goroutine.
// TODO(readpe): Test concurrent access of olxapi.dll
type OlxAPI struct {
	sync.Mutex
	initialized bool
	dll         *syscall.DLL // olxapi.dll

	// OlxAPI Procedures
	errorString       *syscall.Proc
	versionInfo       *syscall.Proc
	saveDataFile      *syscall.Proc
	loadDataFile      *syscall.Proc
	readChangeFile    *syscall.Proc
	getEquipment      *syscall.Proc
	equipmentType     *syscall.Proc
	getData           *syscall.Proc
	findBusByName     *syscall.Proc
	getEquipmentByTag *syscall.Proc
	findBusNo         *syscall.Proc
	setData           *syscall.Proc
	getBusEquipment   *syscall.Proc

	doFault            *syscall.Proc
	faultDescriptionEx *syscall.Proc
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

	// hasp_rt.exe needs to be in same directory as executable. This appears to be a limitation
	// imposed by olxapi.dll, request feature to search PATH directories instead.
	if err := haspRTCopy(); err != nil {
		panic(err)
	}
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
	api.findBusByName = api.dll.MustFindProc("OlxAPIFindBusByName")
	api.getEquipmentByTag = api.dll.MustFindProc("OlxAPIFindEquipmentByTag")
	api.findBusNo = api.dll.MustFindProc("OlxAPIFindBusNo")
	api.setData = api.dll.MustFindProc("OlxAPISetData")
	api.getBusEquipment = api.dll.MustFindProc("OlxAPIGetBusEquipment")

	api.doFault = api.dll.MustFindProc("OlxAPIDoFault")
	api.faultDescriptionEx = api.dll.MustFindProc("OlxAPIFaultDescriptionEx")

	return api
}

// haspRTCopy copies the hasp_rt.exe from ASPEN program directory to the current executables directory, only if the hash sum are different.
// This appears to be a limitation of the olxapi.dll implementation.
func haspRTCopy() error {
	if haspRTShaSumDiff() {
		return nil
	}
	ex, err := os.Executable()
	if err != nil {
		return fmt.Errorf("haspRTCopy: could not locate executable path: %v ", err)
	}
	exPath := filepath.Dir(ex)
	srcFile, err := os.Open(filepath.Join(OlxAPIDLLPath, `hasp_rt.exe`))
	if err != nil {
		return fmt.Errorf("haspRTCopy: could not locate hasp_rt.exe: %v", err)
	}
	defer srcFile.Close()
	destFile, err := os.OpenFile(filepath.Join(exPath, `hasp_rt.exe`), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return fmt.Errorf("haspRTCopy: could not create hasp_rt.exe: %v", err)
	}
	defer destFile.Close()
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("haspRTCopy: could not create hasp_rt.exe: %v", err)
	}
	err = destFile.Sync()
	if err != nil {
		return fmt.Errorf("haspRTCopy: could not sync new hasp_rt.exe: %v", err)
	}

	return nil
}

func haspRTShaSumDiff() bool {
	ex, err := os.Executable()
	if err != nil {
		return false
	}
	exPath := filepath.Dir(ex)
	srcHash := sha1File(filepath.Join(OlxAPIDLLPath, `hasp_rt.exe`))
	exHash := sha1File(filepath.Join(exPath, `hasp_rt.exe`))

	return bytes.Equal(srcHash.Sum(nil), exHash.Sum(nil))
}

func sha1File(name string) hash.Hash {
	h := sha1.New()
	f, err := os.Open(name)
	if err != nil {
		return h
	}
	defer f.Close()

	io.Copy(h, f)

	return h
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
	r, _, _ := o.getEquipment.Call(uintptr(eqType), uintptr(unsafe.Pointer(hnd)))
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

// FindBusByName calls the OlxAPIFindBusByName function.
func (o *OlxAPI) FindBusByName(name string, kv float64) (int, error) {
	b, err := utf8NullFromString(name)
	if err != nil {
		return 0, fmt.Errorf("FindBus: %v", err)
	}
	// Cannot pass float64 by value as uintptr to 32bit dll using syscall directly.
	// Must convert to two uint32 and pass consecutively.
	// See https://github.com/golang/go/issues/29092
	f322 := float64ToUint32(kv)
	var hnd int
	o.Lock()
	r, _, _ := o.findBusByName.Call(uintptr(unsafe.Pointer(&b[0])), uintptr(f322[0]), uintptr(f322[1]), uintptr(unsafe.Pointer(&hnd)))
	o.Unlock()

	if r == OLXAPIFailure {
		return 0, ErrOlxAPI{"FindBusByName", o.ErrorString()}
	}
	return hnd, nil
}

// FindEquipmentByTag calls the OlxAPIFindEquipmentByTag function.
func (o *OlxAPI) FindEquipmentByTag(eqType int, hnd *int, tags ...string) error {
	bTags, err := utf8NullFromString(strings.Join(tags, ","))
	if err != nil {
		return err
	}
	o.Lock()
	r, _, _ := o.getEquipmentByTag.Call(uintptr(unsafe.Pointer(&bTags[0])), uintptr(eqType), uintptr(unsafe.Pointer(hnd)))
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

func (o *OlxAPI) DoFault(hnd int, fltConn [4]int, fltOpt [15]float64, outageOpt [4]int, outageLst []int, fltR, fltX float64, clearPrev bool) error {

	// Cannot pass float64 by value as uintptr to 32bit dll using syscall directly.
	// Must convert to two uint32 and pass consecutively.
	// See https://github.com/golang/go/issues/29092
	fltR322 := float64ToUint32(fltR)
	fltX322 := float64ToUint32(fltX)

	var clear int
	if clearPrev {
		clear = 1
	}

	o.Lock()
	r, _, _ := o.doFault.Call(
		uintptr(hnd),
		uintptr(unsafe.Pointer(&fltConn[0])),
		uintptr(unsafe.Pointer(&fltOpt[0])),
		uintptr(unsafe.Pointer(&outageOpt[0])),
		uintptr(unsafe.Pointer(&outageLst)),
		uintptr(unsafe.Pointer(&fltR322[0])), uintptr(unsafe.Pointer(&fltR322[1])),
		uintptr(unsafe.Pointer(&fltX322[0])), uintptr(unsafe.Pointer(&fltX322[1])),
		uintptr(clear),
	)
	o.Unlock()
	if r == OLXAPIFailure {
		return ErrOlxAPI{"DoFault", o.ErrorString()}
	}
	return nil
}

func (o *OlxAPI) FaultDescriptionEx(index, flag int) string {
	o.Lock()
	r, _, _ := o.faultDescriptionEx.Call(uintptr(index), uintptr(flag))
	o.Unlock()
	return utf8StringFromPtr(r)
}
