// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package goolx

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unsafe"

	"github.com/readpe/goolx/constants"
	"github.com/readpe/goolx/internal/olxapi"
)

// Supported Oneliner Version/Build
const (
	OnelinerVersionSupported = 15.4
	OnelinerBuildSupported   = 17321
)

// Byte size constants
const (
	cIntSize    = int(unsafe.Sizeof(int32(0)))
	cDoubleSize = int(unsafe.Sizeof(float64(0)))
	KiB         = 1 << (10 * 1)
	MiB         = 1 << (10 * 2)
)

// Client represents a new goolx api client.
type Client struct {
	olxAPI *olxapi.OlxAPI
}

// NewClient returns a new goolx Client instance.
func NewClient() *Client {
	return &Client{olxAPI: olxapi.New()}
}

// Release releases the api dll. Must be called when done with use of dll.
func (c *Client) Release() error {
	return c.olxAPI.Release()
}

// Info calls the OlxAPIVersionInfo function, returning
// the string
func (c *Client) Info() string {
	return c.olxAPI.VersionInfo()
}

// Version parses the version number from the olxapi.dll info function.
func (c *Client) Version() (string, error) {
	s := c.Info()
	ss := strings.Split(s, " ")
	if len(ss) < 3 {
		return "", fmt.Errorf("unable to parse api version")
	}
	return ss[2], nil
}

// BuildNumber parses the build number from the olxapi.dll info function.
func (c *Client) BuildNumber() (int, error) {
	s := c.Info()
	ss := strings.Split(s, " ")
	if len(ss) < 5 {
		return 0, fmt.Errorf("unable to parse api build number")
	}
	return strconv.Atoi(ss[4])
}

// SaveDataFile saves *.olr file to disk
func (c *Client) SaveDataFile(name string) error {
	return c.olxAPI.SaveDataFile(name)
}

// LoadDataFile loads *.olr file from disk
func (c *Client) LoadDataFile(name string) error {
	return c.olxAPI.LoadDataFile(name)
}

// CloseDataFile closes the currently loaded *.olr data file.
func (c *Client) CloseDataFile() error {
	return c.olxAPI.CloseDataFile()
}

// ReadChangeFile reads *.chf file from disk and applies to case
func (c *Client) ReadChangeFile(name string) error {
	return c.olxAPI.ReadChangeFile(name)
}

// DeleteEquipment deletes the equipment with the provided handle.
func (c *Client) DeleteEquipment(hnd int) error {
	return c.olxAPI.DeleteEquipment(hnd)
}

// NextEquipment returns an EquipmentIterator type. The EquipmentIterator will loop through all
// equipment handles in the case until it reaches the end. This is done using the Next() and Hnd() methods.
// Note: ASPEN equipment handle integers are not unique and are generated on data access. Therefore care
// should be taken when using handle across functions or applications. It is recommended to use the handle
// immediately after retrieving to get unique equipment identifiers.
func (c *Client) NextEquipment(eqType int) HandleIterator {
	return &NextEquipment{c: c, eqType: eqType}
}

// NextBusEquipment returns an EquipmentIterator type. The EquipmentIterator will loop through all
// equipment handles at the provided bus in the case until it reaches the end. This is done using the Next() and Hnd() methods.
// See NextEquipment for more details.
func (c *Client) NextBusEquipment(busHnd, eqType int) HandleIterator {
	return &NextBusEquipment{c: c, busHnd: busHnd, eqType: eqType}
}

// EquipmentType returns the equipment type code for the equipment with the provided handle
func (c *Client) EquipmentType(hnd int) (int, error) {
	return c.olxAPI.EquipmentType(hnd)
}

// Data represents data returned via the GetData method.
type Data struct {
	err    error
	tokens []int
	data   []interface{}
}

// Scan copies the data from the matched parameter token into the values pointed at by dest.
// The order of the destination pointers should match the parameters queried with GetData.
// Will return an error if any parameters produced an error during GetData call. Data will not
// be populated in this case.
func (d Data) Scan(dest ...interface{}) error {
	// If any errors during GetData call, Scan is returned without populating data.
	if d.err != nil {
		return d.err
	}
	// number of tokens must match data returned
	if len(d.tokens) != len(d.data) {
		return fmt.Errorf("Scan: token and data numbers don't match")
	}
	for i, p := range dest {
		if err := convertAssignData(p, d.data[i]); err != nil {
			return err
		}
	}
	return nil
}

// convertAssignData copies to dest the value in src, converting it if possible.
// An error is returned if the copy would result in loss of information.
// dest should be a pointer type.
func convertAssignData(dest, src interface{}) error {
	switch s := src.(type) {
	case string:
		switch d := dest.(type) {
		case *string:
			if d == nil {
				return fmt.Errorf("convertAssignData: nil pointer")
			}
			*d = s
			return nil
		}
	case float64:
		switch d := dest.(type) {
		case *float64:
			if d == nil {
				return fmt.Errorf("convertAssignData: nil pointer")
			}
			*d = s
			return nil
		}
	case int:
		switch d := dest.(type) {
		case *int:
			if d == nil {
				return fmt.Errorf("convertAssignData: nil pointer")
			}
			*d = s
			return nil
		}
	case []string:
		switch d := dest.(type) {
		case *[]string:
			if d == nil {
				return fmt.Errorf("convertAssignData: nil pointer")
			}
			*d = s
			return nil
		}
	case []float64:
		switch d := dest.(type) {
		case *[]float64:
			if d == nil {
				return fmt.Errorf("convertAssignData: nil pointer")
			}
			*d = s
			return nil
		}
	case []int:
		switch d := dest.(type) {
		case *[]int:
			if d == nil {
				return fmt.Errorf("convertAssignData: nil pointer")
			}
			*d = s
			return nil
		}
	}
	return fmt.Errorf("unsupported Scan, storing data type %T into type %T", src, dest)
}

// GetData returns data for the object handle, and all parameter tokens provided.
// The data for each token can be retrieved using the Scan method on the Data type.
// This is similar to the Row.Scan in the sql package.
func (c *Client) GetData(hnd int, tokens ...int) Data {
	var data = Data{tokens: tokens}
	for _, tkn := range tokens {
		d, err := c.getData(hnd, tkn)
		if err != nil {
			data.err = err
		}
		data.data = append(data.data, d)
	}
	return data
}

// getData returns the requested data for given equipment handle and field token.
// The returned data type is dependent on the token field data type, must inspect empty
// interface concrete type before use.
func (c *Client) getData(hnd, token int) (interface{}, error) {

	eqType, _ := c.olxAPI.EquipmentType(hnd)

	switch token / 100 {

	case constants.VTSTRING:
		// string
		buf := make([]byte, 10*KiB) // 10 KiB buffer for string data null terminated
		err := c.olxAPI.GetData(hnd, token, buf)
		if err != nil {
			return nil, err
		}

		s := olxapi.UTF8NullToString(buf)
		return s, nil

	case constants.VTDOUBLE:
		// double
		buf := make([]byte, 8) // 64 bit (8 byte) float64 buffer, equivalent to C Double
		err := c.olxAPI.GetData(hnd, token, buf)
		if err != nil {
			return nil, err
		}

		f := math.Float64frombits(binary.LittleEndian.Uint64(buf))
		return f, nil

	case constants.VTINTEGER:
		// integers
		buf := make([]byte, 4) // 32 bit (4 byte) int32 buffer
		err := c.olxAPI.GetData(hnd, token, buf)
		if err != nil {
			return nil, err
		}

		i := int(binary.LittleEndian.Uint32(buf)) // Convert []byte to int

		return i, nil

	case constants.VTARRAYSTRING:
		// string array
		buf := make([]byte, 10*KiB) // 10 KiB buffer
		err := c.olxAPI.GetData(hnd, token, buf)
		if err != nil {
			return nil, err
		}

		// tab delimited
		sa := strings.Split(string(buf), "\t")

		return sa, nil

	case constants.VTARRAYINT:
		// array length depends on token
		var length int

		length, ok := constants.ArrayLengths[eqType][token]
		if !ok {
			return nil, fmt.Errorf("array length not found for equipment type: %v; token: %v", eqType, token)
		}

		buf := make([]byte, cIntSize*int(length))
		err := c.olxAPI.GetData(hnd, token, buf)
		if err != nil {
			return nil, err
		}

		// convert []byte buf of type c int array to []int
		data := make([]int32, len(buf)/cIntSize)
		for i := range data {
			data[i] = int32(binary.LittleEndian.Uint32(buf[i*cIntSize : (i+1)*cIntSize]))
		}

		// returning []int
		return data, nil

	case constants.VTARRAYDOUBLE:
		// array length depends on token
		var length int

		length, ok := constants.ArrayLengths[eqType][token]
		if !ok {
			return nil, fmt.Errorf("array length not found for equipment type: %v; token: %v", eqType, token)
		}

		buf := make([]byte, cDoubleSize*length)
		err := c.olxAPI.GetData(hnd, token, buf)
		if err != nil {
			return nil, err
		}

		// convert []byte buf of type c double array to []float64
		data := make([]float64, len(buf)/cDoubleSize)
		for i := range data {
			data[i] = math.Float64frombits(binary.LittleEndian.Uint64(buf[i*cDoubleSize : (i+1)*cDoubleSize]))
		}

		// returning []float64
		return data, nil

	default:
		return nil, fmt.Errorf("GetData token type not found: %d", token)
	}
}

// FindBusByName returns the bus handle for the given bus name and kv, if found
func (c *Client) FindBusByName(name string, kv float64) (int, error) {
	return c.olxAPI.FindBusByName(name, kv)
}

// NextEquipmentByTag returns a NextEquipmentTag type which satisfies the HandleIterator interface.
func (c *Client) NextEquipmentByTag(eqType int, tags ...string) HandleIterator {
	return &NextEquipmentByTag{
		c:      c,
		eqType: eqType,
		tags:   tags,
	}
}

// FindBusNo returns the bus with the provided bus number. Or returns 0 and an error if not found.
func (c *Client) FindBusNo(n int) (int, error) {
	return c.olxAPI.FindBusNo(n)
}

// DoFault runs a fault for the given equipment handle with the providedfault configurations.
// PickFault or NextFault must be called prior to accessing results data.
func (c *Client) DoFault(hnd int, config *FaultConfig) error {
	if config == nil {
		return fmt.Errorf("DoFault: config must not be nil")
	}
	return c.olxAPI.DoFault(
		hnd,
		config.fltConn,
		config.fltOpt,
		config.outageOpt,
		config.outageList,
		config.fltR, config.fltX,
		config.clearPrev,
	)
}

// FaultDescription returns the fault description string for the specified index.
func (c *Client) FaultDescription(index int) string {
	return strings.TrimSpace(c.olxAPI.FaultDescriptionEx(index, 0))
}

// DoSteppedEvent runs a stepped event analysis for the given equipment with the provided config parameters.
func (c *Client) DoSteppedEvent(hnd int, cfg *SteppedEventConfig) error {
	return c.olxAPI.DoSteppedEvent(hnd, cfg.fltOpt, cfg.runOpt, cfg.nTiers)
}

// NextRelay returns an HandleIterator type. The HandleIterator will loop through all
// relay handles in the provided relay group until it reaches the end. This is done using the Next() and Hnd() methods.
// Note: ASPEN equipment handle integers are not unique and are generated on data access. Therefore care
// should be taken when using handle across functions or applications. It is recommended to use the handle
// immediately after retrieving to get unique equipment identifiers.
func (c *Client) NextRelay(rlyGroupHnd int) HandleIterator {
	return &NextRelay{c: c, rlyGroupHnd: rlyGroupHnd}
}

// GetObjTags returns a slice of tag strings for the equipment with the provided handle.
func (c *Client) GetObjTags(hnd int) (tags []string, err error) {
	s, err := c.olxAPI.GetObjTags(hnd)
	if err != nil {
		return
	}
	if s == "" {
		return tags, nil
	}
	tags = strings.Split(s, ",")
	return
}

// SetObjTags replaces the object tag with the provided tags. Will override existing tags, use GetObjTags to retrieve existing tags and append to if
// the wanting to keep existing tags.
func (c *Client) SetObjTags(hnd int, tags ...string) error {
	err := c.olxAPI.SetObjTags(hnd, tags...)
	if err != nil {
		return err
	}
	return nil
}

// AppendObjTags appends the provided tags to the object tag string. Does not check for duplicate tags
func (c *Client) AppendObjTags(hnd int, newTags ...string) error {
	tags, err := c.GetObjTags(hnd)
	if err != nil {
		return err
	}
	return c.SetObjTags(hnd, append(tags, newTags...)...)
}

// ReplaceObjTag replaces all occurences of the old tag with the new tag provided.
func (c *Client) ReplaceObjTag(hnd int, oldTag, newTag string) error {
	tags, err := c.GetObjTags(hnd)
	if err != nil {
		return err
	}
	for i, tag := range tags {
		if tag == oldTag {
			tags[i] = newTag
		}
	}
	return c.SetObjTags(hnd, tags...)
}

// GetObjMemo returns the object memo field string.
func (c *Client) GetObjMemo(hnd int) (string, error) {
	s, err := c.olxAPI.GetObjMemo(hnd)
	if err != nil {
		return "", err
	}
	return s, nil
}

// SetObjMemo sets the object memo field, overwrites existing data.
func (c *Client) SetObjMemo(hnd int, memo string) error {
	err := c.olxAPI.SetObjMemo(hnd, memo)
	if err != nil {
		return err
	}
	return nil
}

// AppendObjMemo appends a new line followed by s to the object memo field.
func (c *Client) AppendObjMemo(hnd int, s string) error {
	memo, err := c.GetObjMemo(hnd)
	if err != nil {
		return err
	}
	memo = fmt.Sprintf("%s\n%s", memo, s)
	return c.SetObjMemo(hnd, memo)
}

// ObjMemoContains reports whether substr is within the objects memo field.
func (c *Client) ObjMemoContains(hnd int, substr string) bool {
	memo, err := c.GetObjMemo(hnd)
	if err != nil {
		return false
	}
	return strings.Contains(memo, substr)
}

// ReplaceAllObjMemo replaces all non-overlapping instances of old replaced by new
// in the object memo field.
func (c *Client) ReplaceAllObjMemo(hnd int, old, new string) error {
	memo, err := c.GetObjMemo(hnd)
	if err != nil {
		return err
	}
	memo = strings.ReplaceAll(memo, old, new)
	return c.SetObjMemo(hnd, memo)
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
func (c *Client) PickFault(indx, tiers int) error {
	return c.olxAPI.PickFault(indx, tiers)
}

// GetSCVoltagePhase gets the short circuit phase voltage for the equipment with the provided handle.
// Returns Va, Vb, Vc Phasor types.
func (c *Client) GetSCVoltagePhase(hnd int) (Va, Vb, Vc Phasor, err error) {
	vdOut1, vdOut2, err := c.olxAPI.GetSCVoltage(hnd, 3)
	if err != nil {
		return Va, Vb, Vc, err
	}
	Va = Phasor(complex(vdOut1[0], vdOut2[0]))
	Vb = Phasor(complex(vdOut1[1], vdOut2[1]))
	Vc = Phasor(complex(vdOut1[2], vdOut2[2]))
	return
}

// GetSCVoltageSeq gets the short circuit sequence voltagse for the equipment with the provided handle.
// Returns V0, V1, V2 Phasor types.
func (c *Client) GetSCVoltageSeq(hnd int) (V0, V1, V2 Phasor, err error) {
	vdOut1, vdOut2, err := c.olxAPI.GetSCVoltage(hnd, 1)
	if err != nil {
		return V0, V1, V2, err
	}
	V0 = Phasor(complex(vdOut1[0], vdOut2[0]))
	V1 = Phasor(complex(vdOut1[1], vdOut2[1]))
	V2 = Phasor(complex(vdOut1[2], vdOut2[2]))
	return
}
