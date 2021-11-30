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
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"io"
	"math"
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

var (
	ErrFaultNotRun    = errors.New("fault not simulated")
	ErrFaultNotPicked = errors.New("fault not picked")
)

// OlxAPI represents a connection to the olxapi.dll. Provides method
// wrappers for each api function. Instantiate using New().
//
// It is unclear if the olxapi.dll can be called cuncurrently if loaded into different processes,
// e.g. instantiating a new Client in a goroutine.
// TODO(readpe): Test concurrent access of olxapi.dll
type OlxAPI struct {
	sync.Mutex
	dll *syscall.DLL // olxapi.dll

	faultRun    bool
	faultPicked bool

	// OlxAPI Procedures
	errorString       *syscall.Proc
	versionInfo       *syscall.Proc
	saveDataFile      *syscall.Proc
	loadDataFile      *syscall.Proc
	getOlrFileName    *syscall.Proc
	closeDataFile     *syscall.Proc
	readChangeFile    *syscall.Proc
	getEquipment      *syscall.Proc
	deleteEquipment   *syscall.Proc
	equipmentType     *syscall.Proc
	getData           *syscall.Proc
	findBusByName     *syscall.Proc
	getEquipmentByTag *syscall.Proc
	findBusNo         *syscall.Proc
	setData           *syscall.Proc
	getBusEquipment   *syscall.Proc

	doFault            *syscall.Proc
	faultDescriptionEx *syscall.Proc
	doSteppedEvent     *syscall.Proc
	getSteppedEvent    *syscall.Proc
	getRelay           *syscall.Proc

	getObjTags  *syscall.Proc
	setObjTags  *syscall.Proc
	getObjMemo  *syscall.Proc
	setObjMemo  *syscall.Proc
	getObjGUID  *syscall.Proc
	getAreaName *syscall.Proc
	getZoneName *syscall.Proc

	pickFault    *syscall.Proc
	getSCVoltage *syscall.Proc
	getSCCurrent *syscall.Proc
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
	api.getOlrFileName = api.dll.MustFindProc("OlxAPIGetOlrFileName")
	api.closeDataFile = api.dll.MustFindProc("OlxAPICloseDataFile")
	api.readChangeFile = api.dll.MustFindProc("OlxAPIReadChangeFile")
	api.getEquipment = api.dll.MustFindProc("OlxAPIGetEquipment")
	api.deleteEquipment = api.dll.MustFindProc("OlxAPIDeleteEquipment")
	api.equipmentType = api.dll.MustFindProc("OlxAPIEquipmentType")
	api.getData = api.dll.MustFindProc("OlxAPIGetData")
	api.findBusByName = api.dll.MustFindProc("OlxAPIFindBusByName")
	api.getEquipmentByTag = api.dll.MustFindProc("OlxAPIFindEquipmentByTag")
	api.findBusNo = api.dll.MustFindProc("OlxAPIFindBusNo")
	api.setData = api.dll.MustFindProc("OlxAPISetData")
	api.getBusEquipment = api.dll.MustFindProc("OlxAPIGetBusEquipment")
	api.doFault = api.dll.MustFindProc("OlxAPIDoFault")
	api.faultDescriptionEx = api.dll.MustFindProc("OlxAPIFaultDescriptionEx")
	api.doSteppedEvent = api.dll.MustFindProc("OlxAPIDoSteppedEvent")
	api.getSteppedEvent = api.dll.MustFindProc("OlxAPIGetSteppedEvent")
	api.getRelay = api.dll.MustFindProc("OlxAPIGetRelay")
	api.getObjTags = api.dll.MustFindProc("OlxAPIGetObjTags")
	api.setObjTags = api.dll.MustFindProc("OlxAPISetObjTags")
	api.getObjMemo = api.dll.MustFindProc("OlxAPIGetObjMemo")
	api.setObjMemo = api.dll.MustFindProc("OlxAPISetObjMemo")
	api.getObjGUID = api.dll.MustFindProc("OlxAPIGetObjGUID")
	api.getAreaName = api.dll.MustFindProc("OlxAPIGetAreaName")
	api.getZoneName = api.dll.MustFindProc("OlxAPIGetZoneName")
	api.pickFault = api.dll.MustFindProc("OlxAPIPickFault")
	api.getSCVoltage = api.dll.MustFindProc("OlxAPIGetSCVoltage")
	api.getSCCurrent = api.dll.MustFindProc("OlxAPIGetObjGUID")

	return api
}

// haspRTCopy copies the hasp_rt.exe from ASPEN program directory to the current executables directory, only if the hash sum are different.
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

// resetFault resets the faultRun and faultPicked flags.
func (o *OlxAPI) resetFault() {
	o.Lock()
	defer o.Unlock()
	o.faultRun = false
	o.faultPicked = false
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

// GetOlrFileName returns the currently loaded olr file name.
func (o *OlxAPI) GetOlrFileName() string {
	o.Lock()
	r, _, _ := o.getOlrFileName.Call()
	o.Unlock()
	return utf8StringFromPtr(r)
}

func (o *OlxAPI) CloseDataFile() error {
	o.Lock()
	r, _, _ := o.closeDataFile.Call()
	o.Unlock()
	if r == OLXAPIFailure {
		return ErrOlxAPI{"CloseDataFile", o.ErrorString()}
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

// DeleteEquipment deletes the equipment with the provided handle.
func (o *OlxAPI) DeleteEquipment(hnd int) error {
	o.Lock()
	r, _, _ := o.deleteEquipment.Call(uintptr(hnd))
	o.Unlock()
	if r == OLXAPIFailure {
		return ErrOlxAPI{"DeleteEquipment", o.ErrorString()}
	}
	return nil
}

// EquipmentType calls the OlxAPIEquipmentType function. Returns
// the equipment type code. Returns error if OLXAPIFailure
// is returned.
func (o *OlxAPI) EquipmentType(hnd int) (int, error) {
	o.Lock()
	r, _, _ := o.equipmentType.Call(uintptr(hnd))
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
		return ErrOlxAPI{"GetData", o.ErrorString()}
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
	// Resets faultRun and faultPicked flags.
	o.resetFault()
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
	o.faultRun = true
	o.Unlock()
	if r == OLXAPIFailure {
		o.resetFault()
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

// DoSteppedEvent runs a stepped-event simulation utilizing the provided parameters.
// Refer to Oneliner scripting documentation for options details.
func (o *OlxAPI) DoSteppedEvent(hnd int, fltOpt [64]float64, runOpt [7]int, nTiers int) error {
	o.resetFault()
	o.Lock()
	r, _, _ := o.doSteppedEvent.Call(uintptr(hnd), uintptr(unsafe.Pointer(&fltOpt[0])), uintptr(unsafe.Pointer(&runOpt[0])), uintptr(nTiers))
	o.faultRun = true
	o.Unlock()
	if r == OLXAPIFailure {
		o.resetFault()
		return ErrOlxAPI{"DoSteppedEvent", o.ErrorString()}
	}
	return nil
}

// GetSteppedEvent gets the stepped event data for the provided step. Returns an error if step index is out of range.
func (o *OlxAPI) GetSteppedEvent(step int) (t, current float64, userEvent int, eventDesc, faultDesc string, err error) {
	var bufT, bufCurrent [8]byte    // double buffers
	var bufEventDesc [4 * 512]byte  // event description string buffer, 4*512 bytes per Samples.py
	var bufFaultDesc [50 * 512]byte // event description string buffer, 50*512 bytes per Samples.py

	o.Lock()
	r, _, _ := o.getSteppedEvent.Call(
		uintptr(step),
		uintptr(unsafe.Pointer(&bufT)),
		uintptr(unsafe.Pointer(&bufCurrent)),
		uintptr(unsafe.Pointer(&userEvent)),
		uintptr(unsafe.Pointer(&bufEventDesc)),
		uintptr(unsafe.Pointer(&bufFaultDesc)),
	)
	o.Unlock()
	if r == OLXAPIFailure {
		err = ErrOlxAPI{"GetSteppedEvent", o.ErrorString()}
		return
	}
	// Convert result variables
	t = math.Float64frombits(binary.LittleEndian.Uint64(bufT[:]))
	current = math.Float64frombits(binary.LittleEndian.Uint64(bufCurrent[:]))
	// userEvent set directly
	eventDesc = UTF8NullToString(bufEventDesc[:])
	faultDesc = UTF8NullToString(bufFaultDesc[:])
	return
}

// GetRelay calls the OlxAPIGetRelay function. Returns
// the relay handle. Returns an error if OLXAPIFailure
// is returned. Returns io.EOF error when iteration is exhausted.
func (o *OlxAPI) GetRelay(rlyGroupHnd int, hnd *int) error {
	o.Lock()
	r, _, _ := o.getRelay.Call(uintptr(rlyGroupHnd), uintptr(unsafe.Pointer(hnd)))
	o.Unlock()
	switch int(r) {
	case -1:
		// OlxAPI returns -1 when GetRelay is exhausted, returning EOF error.
		return io.EOF
	case OLXAPIFailure:
		return ErrOlxAPI{"GetRelay", o.ErrorString()}
	}
	return nil
}

// GetObjTags calls OlxAPIGetObjTags function. Returns a string of comma separated tags.
func (o *OlxAPI) GetObjTags(hnd int) (string, error) {
	o.Lock()
	r, _, _ := o.getObjTags.Call(uintptr(hnd))
	o.Unlock()
	s := strings.TrimSpace(utf8StringFromPtr(r))
	if strings.HasPrefix(s, "GetObjTags failure:") {
		return "", ErrOlxAPI{"GetObjTags", s}
	}
	return s, nil
}

// SetObjTags calls OlxAPISetObjTags function. Tags are joined into a comma separated string.
func (o *OlxAPI) SetObjTags(hnd int, tags ...string) error {
	bTags, err := utf8NullFromString(strings.Join(tags, ","))
	if err != nil {
		return err
	}
	o.Lock()
	r, _, _ := o.setObjTags.Call(uintptr(hnd), uintptr(unsafe.Pointer(&bTags[0])))
	o.Unlock()
	if r == OLXAPIFailure {
		return ErrOlxAPI{"SetObjTags", o.ErrorString()}
	}
	return nil
}

// GetObjMemo calls OlxAPIGetObjMemo function. Returns the object memo string.
func (o *OlxAPI) GetObjMemo(hnd int) (string, error) {
	o.Lock()
	r, _, _ := o.getObjMemo.Call(uintptr(hnd))
	o.Unlock()
	s := utf8StringFromPtr(r)
	if strings.HasPrefix(s, "GetObjMemo failure:") {
		return "", ErrOlxAPI{"GetObjMemo", s}
	}
	return s, nil
}

// SetObjMemo calls OlxAPISetObjMemo function. Sets the object memo field. Overwrites existing data.
func (o *OlxAPI) SetObjMemo(hnd int, memo string) error {
	bMemo, err := utf8NullFromString(memo)
	if err != nil {
		return err
	}
	o.Lock()
	r, _, _ := o.setObjMemo.Call(uintptr(hnd), uintptr(unsafe.Pointer(&bMemo[0])))
	o.Unlock()
	if r == OLXAPIFailure {
		return ErrOlxAPI{"SetObjMemo", o.ErrorString()}
	}
	return nil
}

// GetObjGUID returns the GUID of the given object. Returns empty string if error.
func (o *OlxAPI) GetObjGUID(hnd int) (string, error) {
	o.Lock()
	r, _, _ := o.getObjGUID.Call(uintptr(hnd))
	o.Unlock()
	s := utf8StringFromPtr(r)
	if strings.HasPrefix(s, "GetObjGUID failure:") {
		return "", ErrOlxAPI{"GetObjGUID", s}
	}
	return s, nil
}

// GetAreaName returns the area name given the area id.
func (o *OlxAPI) GetAreaName(area int) (string, error) {
	o.Lock()
	r, _, _ := o.getAreaName.Call(uintptr(area))
	o.Unlock()
	s := utf8StringFromPtr(r)
	if strings.HasPrefix(s, "GetAreaName failure") {
		return "", ErrOlxAPI{"GetAreaName", s}
	}
	return s, nil
}

// GetZoneName returns the area name given the zone id.
func (o *OlxAPI) GetZoneName(zone int) (string, error) {
	o.Lock()
	r, _, _ := o.getZoneName.Call(uintptr(zone))
	o.Unlock()
	s := utf8StringFromPtr(r)
	if strings.HasPrefix(s, "GetZoneName failure:") {
		return "", ErrOlxAPI{"GetZoneName", s}
	}
	return s, nil
}

// PickFault must be called before accessing short circuit simulation data. The given index and number of tiers
// to be calculated are provided. See NextFault for an iterator which automatically switches from SFFirst to SFNext
// after the first fault until the last.
//
//	The index codes are:
//		SFLast     = -1
//		SFNext     = -2
//		SFFirst    = 1
//		SFPrevious = -4
func (o *OlxAPI) PickFault(indx, tiers int) error {
	if !o.faultRun {
		return fmt.Errorf("PickFault: %v", ErrFaultNotRun)
	}
	o.Lock()
	r, _, _ := o.pickFault.Call(uintptr(indx), uintptr(tiers))
	o.faultPicked = true
	o.Unlock()
	if r == OLXAPIFailure {
		o.Lock()
		o.faultPicked = false
		o.Unlock()
		return ErrOlxAPI{"PickFault", o.ErrorString()}
	}
	return nil
}

// GetSCVoltage Retrieves post-fault voltage of a bus, or of connected buses of
// a line, transformer, switch or phase shifter.
//
// The returned array size depends on the equipment type, a bus handle will
// only return an array length of 3, whereas a line/transformer will return
// an array of length 9. An invalid access error may occur if the incorrect
// size is presented to the Call.
//
// 	Result style codes:
//		1: output 012 sequence voltage in rectangular form
//		2: output 012 sequence voltage in polar form
//		3: output ABC phase voltage in rectangular form
//		4: output ABC phase voltage in polar form
func (o *OlxAPI) GetSCVoltage(hnd, styleCode int) (vdOut1 [9]float64, vdOut2 [9]float64, err error) {
	switch {
	case !o.faultRun:
		return vdOut1, vdOut2, fmt.Errorf("GetSCVoltage: %v", ErrFaultNotRun)
	case !o.faultPicked:
		return vdOut1, vdOut2, fmt.Errorf("GetSCVoltage: %v", ErrFaultNotPicked)
	}
	o.Lock()
	r, _, _ := o.getSCVoltage.Call(uintptr(hnd), uintptr(unsafe.Pointer(&vdOut1[0])), uintptr(unsafe.Pointer(&vdOut2[0])), uintptr(styleCode))
	o.Unlock()
	if r == OLXAPIFailure {
		return vdOut1, vdOut2, ErrOlxAPI{"GetSCVoltage", o.ErrorString()}
	}
	return vdOut1, vdOut2, nil
}

// GetSCCurrent returns the post fault current for a generator, load, shunt, switched shunt,
// generating unit, load unit, shunt unit, transmission line, transformer,
// switch or phase shifter.
//
// You can get the total fault current by calling this function with the
// pre-defined handle of short circuit solution, HND_SC.
//
// 	Result style codes:
//		1: output 012 sequence voltage in rectangular form
//		2: output 012 sequence voltage in polar form
//		3: output ABC phase voltage in rectangular form
//		4: output ABC phase voltage in polar form
func (o *OlxAPI) GetSCCurrent(hnd, styleCode int) (vdOut1 [12]float64, vdOut2 [12]float64, err error) {
	switch {
	case !o.faultRun:
		return vdOut1, vdOut2, fmt.Errorf("GetSCCurrent: %v", ErrFaultNotRun)
	case !o.faultPicked:
		return vdOut1, vdOut2, fmt.Errorf("GetSCCurrent: %v", ErrFaultNotPicked)
	}
	o.Lock()
	r, _, _ := o.getSCCurrent.Call(uintptr(hnd), uintptr(unsafe.Pointer(&vdOut1[0])), uintptr(unsafe.Pointer(&vdOut2[0])), uintptr(styleCode))
	o.Unlock()
	if r == OLXAPIFailure {
		return vdOut1, vdOut2, ErrOlxAPI{"GetSCCurrent", o.ErrorString()}
	}
	return vdOut1, vdOut2, nil
}
