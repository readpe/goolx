// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package goolx

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unsafe"

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

// Client represents a new goolx api client. OlxAPI calls cannot be called in parallel,
// the underlying dll procedure calls share memory and do not support cuncurency.
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

// LoadDataFile loads *.olr file from disk. Opens read/write.
func (c *Client) LoadDataFile(name string) error {
	return c.olxAPI.LoadDataFile(name, false)
}

// LoadDataFile loads *.olr file from disk. Opens read only.
func (c *Client) LoadDataFileReadOnly(name string) error {
	return c.olxAPI.LoadDataFile(name, true)
}

// Returns the currently loaded olr filename.
func (c *Client) GetOlrFilename() string {
	return c.olxAPI.GetOlrFileName()
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
	return &handleIterator{
		f: func(hnd *int) error {
			return c.olxAPI.GetEquipment(eqType, hnd)
		},
	}
}

// NextBusEquipment returns an EquipmentIterator type. The EquipmentIterator will loop through all
// equipment handles at the provided bus in the case until it reaches the end. This is done using the Next() and Hnd() methods.
// See NextEquipment for more details.
func (c *Client) NextBusEquipment(busHnd, eqType int) HandleIterator {
	return &handleIterator{
		f: func(hnd *int) error {
			return c.olxAPI.GetBusEquipment(busHnd, eqType, hnd)
		},
	}
}

// BoundaryConfig represents the configuration parameters for the BoundaryEquivalent method.
type BoundaryConfig [3]float64

// BoundaryOption represents an option modifying function to be applied to a BoundaryConfig type.
type BoundaryOption func(*BoundaryConfig)

// BoundaryEliminationThreshold sets the per unit elimination threshold. Refer to Oneliner documentation for more details.
func BoundaryEliminationThreshold(pu float64) BoundaryOption {
	return func(bc *BoundaryConfig) {
		bc[0] = pu
	}
}

// BoundaryKeepEquipment enables the keep bus equipment configuration option
func BoundaryKeepEquipment() BoundaryOption {
	return func(bc *BoundaryConfig) {
		bc[1] = 1
	}
}

// BoundaryKeepAnnotations enables the keep annotations configuration option.
func BoundaryKeepAnnotations() BoundaryOption {
	return func(bc *BoundaryConfig) {
		bc[2] = 1
	}
}

// NewBoundaryConfig creates a new BoundaryConfig with the provided options functions applied.
func NewBoundaryConfig(options ...BoundaryOption) BoundaryConfig {
	var bc = BoundaryConfig{}
	for _, opt := range options {
		opt(&bc)
	}
	return bc
}

// BoundaryEquivalent created an equivalent case with the provided bus list and config.
func (c *Client) BoundaryEquivalent(file string, busList []int, cfg BoundaryConfig) error {
	return c.olxAPI.BoundaryEquivalent(file, busList, [3]float64(cfg))
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

	case VTSTRING:
		// string
		buf := make([]byte, 10*KiB) // 10 KiB buffer for string data null terminated
		err := c.olxAPI.GetData(hnd, token, buf)
		if err != nil {
			return nil, err
		}

		s := olxapi.UTF8NullToString(buf)
		return s, nil

	case VTDOUBLE:
		// double
		buf := make([]byte, 8) // 64 bit (8 byte) float64 buffer, equivalent to C Double
		err := c.olxAPI.GetData(hnd, token, buf)
		if err != nil {
			return nil, err
		}

		f := math.Float64frombits(binary.LittleEndian.Uint64(buf))
		return f, nil

	case VTINTEGER:
		// integers
		buf := make([]byte, 4) // 32 bit (4 byte) int32 buffer
		err := c.olxAPI.GetData(hnd, token, buf)
		if err != nil {
			return nil, err
		}

		i := int(binary.LittleEndian.Uint32(buf)) // Convert []byte to int

		return i, nil

	case VTARRAYSTRING:
		// string array
		buf := make([]byte, 10*KiB) // 10 KiB buffer
		err := c.olxAPI.GetData(hnd, token, buf)
		if err != nil {
			return nil, err
		}

		// tab delimited
		sa := strings.Split(string(buf), "\t")

		return sa, nil

	case VTARRAYINT:
		// array length depends on token
		var length int

		length, ok := ArrayLengths[eqType][token]
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

	case VTARRAYDOUBLE:
		// array length depends on token
		var length int

		length, ok := ArrayLengths[eqType][token]
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

// SetData sets the provided data to the specified equipment parameter as determined by the token value.
// Only string, float64, and int data is currently supported. PostData must be called after SetData.
// TODO: Figure out how to set array data fields.
func (c *Client) SetData(hnd, token int, data interface{}) error {
	switch d := data.(type) {
	case string:
		if token/100 != VTSTRING {
			return fmt.Errorf("SetData: incorrect data type provided for token %d: %T", token, d)
		}
		buf, _ := olxapi.UTF8NullFromString(d)
		err := c.olxAPI.SetData(hnd, token, buf)
		if err != nil {
			return err
		}
	case float64:
		if token/100 != VTDOUBLE {
			return fmt.Errorf("SetData: incorrect data type provided for token %d: %T", token, d)
		}
		var buf = bytes.Buffer{}
		binary.Write(&buf, binary.LittleEndian, d)
		err := c.olxAPI.SetData(hnd, token, buf.Bytes())
		if err != nil {
			return err
		}
	case int:
		if token/100 != VTINTEGER {
			return fmt.Errorf("SetData: incorrect data type provided for token %d: %T", token, d)
		}
		var buf = bytes.Buffer{}
		binary.Write(&buf, binary.LittleEndian, int32(d))
		err := c.olxAPI.SetData(hnd, token, buf.Bytes())
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("SetData: data type %T not supported", data)
	}
	return nil
}

// PostData will post data for the provided equipment handle that was previously set using the SetData method.
func (c *Client) PostData(hnd int) error {
	return c.olxAPI.PostData(hnd)
}

// FindBusByName returns the bus handle for the given bus name and kv, if found
func (c *Client) FindBusByName(name string, kv float64) (int, error) {
	hnd, err := c.olxAPI.FindBusByName(name, kv)
	if err != nil {
		return 0, fmt.Errorf("FindBusByName: could not find bus %s %0.2f", name, kv)
	}
	return hnd, nil
}

// NextEquipmentByTag returns a NextEquipmentTag type which satisfies the HandleIterator interface.
func (c *Client) NextEquipmentByTag(eqType int, tags ...string) HandleIterator {
	return &handleIterator{
		f: func(hnd *int) error {
			return c.olxAPI.FindEquipmentByTag(eqType, hnd, tags...)
		},
	}
}

// FindBusNo returns the bus with the provided bus number. Or returns 0 and an error if not found.
func (c *Client) FindBusNo(n int) (int, error) {
	return c.olxAPI.FindBusNo(n)
}

// OtgTypeMask represents a bit mask for use with the MakeOutageList to provide
// the desired outage type code, bitwise or the desired codes.
type OtgTypeMask uint8

// Outage type bit masks.
const (
	OtgLine OtgTypeMask = 1 << iota
	OtgXfmr
	OtgPhaseShift
	OtgXfmr3
	OtgSwitch
)

// MakeOutageList creates an outage list for use in the DoFault fault simulation
// analysis. Select the outaged branch types by bitwise or of OtgTypeMask's.
func (c *Client) MakeOutageList(hnd, tiers int, otgType OtgTypeMask) ([]int, error) {
	return c.olxAPI.MakeOutageList(hnd, tiers, int(otgType))
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

// GetSteppedEvent gets the stepped event data for the provided step. Returns an error if step index is out of range.
func (c *Client) GetSteppedEvent(step int) (SteppedEvent, error) {
	var userEvent bool
	t, current, userEventInt, eventDesc, faultDesc, err := c.olxAPI.GetSteppedEvent(step)
	if err != nil {
		return SteppedEvent{}, err
	}
	if userEventInt == 1 {
		userEvent = true
	}
	return SteppedEvent{
		Step:             step,
		UserEvent:        userEvent,
		Time:             t,
		Current:          current,
		EventDescription: eventDesc,
		FaultDescription: faultDesc,
	}, nil
}

func (c *Client) NextSteppedEvent() SteppedEventIterator {
	return &steppedEventIterator{
		f: func(step *int) (SteppedEvent, error) {
			*step++
			se, err := c.GetSteppedEvent(*step)
			if err != nil {
				return se, err
			}
			return se, nil
		},
	}
}

// NextRelay returns an HandleIterator type. The HandleIterator will loop through all
// relay handles in the provided relay group until it reaches the end. This is done using the Next() and Hnd() methods.
// Note: ASPEN equipment handle integers are not unique and are generated on data access. Therefore care
// should be taken when using handle across functions or applications. It is recommended to use the handle
// immediately after retrieving to get unique equipment identifiers.
func (c *Client) NextRelay(rlyGroupHnd int) HandleIterator {
	return &handleIterator{
		f: func(hnd *int) error {
			return c.olxAPI.GetRelay(rlyGroupHnd, hnd)
		},
	}
}

// GetRelayTime returns the relay operation time and operation text for the specified relay. TripOnly will only consider tripping relays if true.
// The mult factor multiplies the relay current by the factor provided, this should normally be set to 1.0, an error will be
// returned if mult == 0 which will just result in NOP results.
func (c *Client) GetRelayTime(rlyHnd int, mult float64, tripOnly bool) (float64, string, error) {
	if mult == 0 {
		return 0, "", fmt.Errorf("GetRelayTime: mult factor should be greater than 0")
	}
	return c.olxAPI.GetRelayTime(rlyHnd, mult, tripOnly)
}

// ComputeRelayTimeParams represents input parameters for use with the ComputeRelayTime method.
type ComputeRelayTimeParams struct {
	// Relay phase current.
	Ia, Ib, Ic Phasor
	// Relay neutral winding currents, if applicable, winding P and S.
	In1, In2 Phasor
	// Relay phase voltages.
	Va, Vb, Vc Phasor
	// Relau pre-fault voltage.
	VPre Phasor
}

// ComputeRelayTime computes operating time for a fuse, recloser, an overcurrent relay (phase or ground),
// or a distance relay (phase or ground) at given currents and voltages.
func (c *Client) ComputeRelayTime(hnd int, p ComputeRelayTimeParams) (float64, string, error) {
	return c.olxAPI.ComputeRelayTime(
		hnd,
		[5]float64{p.Ia.Mag(), p.Ib.Mag(), p.Ic.Mag(), p.In1.Mag(), p.In2.Mag()},
		[5]float64{p.Ia.Ang(), p.Ib.Ang(), p.Ic.Ang(), p.In1.Ang(), p.In2.Ang()},
		[3]float64{p.Va.Mag(), p.Vb.Mag(), p.Vc.Mag()},
		[3]float64{p.Va.Ang(), p.Vb.Ang(), p.Vc.Ang()},
		p.VPre.Mag(),
		p.VPre.Ang(),
	)
}

// NextLogicScheme returns the next Logic Scheme handle available in the provided relay group.
// Utilize the Next() method to loop through the available handles. The Hnd() method returns the selected handle.
func (c *Client) NextLogicScheme(rlyGroupHnd int) HandleIterator {
	return &handleIterator{
		f: func(hnd *int) error {
			return c.olxAPI.GetLogicScheme(rlyGroupHnd, hnd)
		},
	}
}

// TagsGet returns a slice of tag strings for the equipment with the provided handle.
func (c *Client) TagsGet(hnd int) (tags []string, err error) {
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

// TagsSet replaces the object tag with the provided tags. Will override existing tags, use GetObjTags to retrieve existing tags and append to if
// the wanting to keep existing tags.
func (c *Client) TagsSet(hnd int, tags ...string) error {
	err := c.olxAPI.SetObjTags(hnd, tags...)
	if err != nil {
		return err
	}
	return nil
}

// TagsAppend appends the provided tags to the object tag string. Does not check for duplicate tags
func (c *Client) TagsAppend(hnd int, newTags ...string) error {
	tags, err := c.TagsGet(hnd)
	if err != nil {
		return err
	}
	return c.TagsSet(hnd, append(tags, newTags...)...)
}

// TagReplace replaces all occurences of the old tag with the new tag provided.
func (c *Client) TagReplace(hnd int, oldTag, newTag string) error {
	tags, err := c.TagsGet(hnd)
	if err != nil {
		return err
	}
	for i, tag := range tags {
		if tag == oldTag {
			tags[i] = newTag
		}
	}
	return c.TagsSet(hnd, tags...)
}

// MemoGet returns the object memo field string.
func (c *Client) MemoGet(hnd int) (string, error) {
	s, err := c.olxAPI.GetObjMemo(hnd)
	if err != nil {
		return "", err
	}
	return s, nil
}

// MemoSet sets the object memo field, overwrites existing data.
func (c *Client) MemoSet(hnd int, memo string) error {
	err := c.olxAPI.SetObjMemo(hnd, memo)
	if err != nil {
		return err
	}
	return nil
}

// MemoAppend appends a new line followed by s to the object memo field.
func (c *Client) MemoAppend(hnd int, s string) error {
	memo, err := c.MemoGet(hnd)
	if err != nil {
		return err
	}
	memo = fmt.Sprintf("%s\n%s", memo, s)
	return c.MemoSet(hnd, memo)
}

// MemoContains reports whether substr is within the objects memo field.
func (c *Client) MemoContains(hnd int, substr string) bool {
	memo, err := c.MemoGet(hnd)
	if err != nil {
		return false
	}
	return strings.Contains(memo, substr)
}

// MemoReplaceAll replaces all non-overlapping instances of old replaced by new
// in the object memo field.
func (c *Client) MemoReplaceAll(hnd int, old, new string) error {
	memo, err := c.MemoGet(hnd)
	if err != nil {
		return err
	}
	memo = strings.ReplaceAll(memo, old, new)
	return c.MemoSet(hnd, memo)
}

// GetGUID returns the GUID for the provided object.
func (c *Client) GetGUID(hnd int) (string, error) {
	return c.olxAPI.GetObjGUID(hnd)
}

// Journal represents a Oneliner object journal record. Obtained from GetJournal method.
type Journal struct {
	CreatedAt  string
	CreatedBy  string
	ModifiedAt string
	ModifiedBy string
}

// GetJournal returns the object journal record for the provided handle.
func (c *Client) GetJournal(hnd int) Journal {
	s := c.olxAPI.GetObjJournalRecord(hnd)
	ss := strings.Split(s, "\n")
	j := Journal{}
	if len(ss) == 4 {
		j.CreatedAt = ss[0]
		j.CreatedBy = ss[1]
		j.ModifiedAt = ss[2]
		j.ModifiedBy = ss[3]
	}
	return j
}

// GetUDF returns the user defined field at the provided equipment with the specified field name.
func (c *Client) GetUDF(hnd int, field string) (string, error) {
	s, err := c.olxAPI.GetObjUDF(hnd, field)
	if err != nil {
		return "", err
	}
	return s, nil
}

// GetUDFByIndex returns the user defined field at the provided equipment with the specified field index.
func (c *Client) GetUDFByIndex(hnd, i int) (field, value string, err error) {
	field, value, err = c.olxAPI.GetObjUDFByIndex(hnd, i)
	if err != nil {
		return "", "", err
	}
	return field, value, nil
}

// SetUDF sets the user defined field at the provided equipment with the specified field name and value.
// SetUDF does not create a new user defined field if it does not exist. User defined fields must be created
// in Oneliner GUI.
func (c *Client) SetUDF(hnd int, field, value string) error {
	err := c.olxAPI.SetObjUDF(hnd, field, value)
	if err != nil {
		return err
	}
	return nil
}

// Find1LPF returns the handle for the object with the provided string id.
func (c *Client) Find1LPF(id string) (int, error) {
	return c.olxAPI.FindObj1LPF(id)
}

// Print1LPF returns the object string id.
func (c *Client) Print1LPF(hnd int) (string, error) {
	return c.olxAPI.PrintObj1LPF(hnd)
}

// GetAreaName returns the area name for the provided area id.
func (c *Client) GetAreaName(area int) (string, error) {
	return c.olxAPI.GetAreaName(area)
}

// GetZoneName returns the zone name for the provided zone id.
func (c *Client) GetZoneName(zone int) (string, error) {
	return c.olxAPI.GetZoneName(zone)
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

// NextFault returns a fault index iterator for looping through fault results. Will perform a PickFault function
// call for each fault simulation result.
func (c *Client) NextFault(tiers int) FaultIterator {
	return &faultIterator{
		f: func(i *int) error {
			*i++
			return c.olxAPI.PickFault(*i, tiers)
		},
	}
}

// GetPSCVoltageKV returns the pre-fault voltage for the provided bus or equipment in kV.
// See Oneliner documentation for returned array structure details.
func (c *Client) GetPSCVoltageKV(hnd int) ([]Phasor, error) {
	vdOut1, vdOut2, err := c.olxAPI.GetPSCVoltage(hnd, 1)
	if err != nil {
		return nil, err
	}
	return []Phasor{
		NewPhasor(vdOut1[0], vdOut2[0]),
		NewPhasor(vdOut1[1], vdOut2[1]),
		NewPhasor(vdOut1[2], vdOut2[2]),
	}, nil
}

// GetPSCVoltagePU returns the pre-fault voltage for the provided bus or equipment in PU.
// See Oneliner documentation for returned array structure details.
func (c *Client) GetPSCVoltagePU(hnd int) ([]Phasor, error) {
	vdOut1, vdOut2, err := c.olxAPI.GetPSCVoltage(hnd, 2)
	if err != nil {
		return nil, err
	}
	return []Phasor{
		NewPhasor(vdOut1[0], vdOut2[0]),
		NewPhasor(vdOut1[1], vdOut2[1]),
		NewPhasor(vdOut1[2], vdOut2[2]),
	}, nil
}

// GetSCVoltagePhase gets the short circuit phase voltage for the equipment with the provided handle.
// Returns Va, Vb, Vc Phasor types. PickFault must be called first.
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

// GetSCCurrentPhase gets the short circuit phase current for the equipment with the provided handle.
// Returns Ia, Ib, Ic Phasor types. PickFault must be called first.
func (c *Client) GetSCCurrentPhase(hnd int) (Ia, Ib, Ic Phasor, err error) {
	vdOut1, vdOut2, err := c.olxAPI.GetSCCurrent(hnd, 3)
	if err != nil {
		return Ia, Ib, Ic, err
	}
	Ia = Phasor(complex(vdOut1[0], vdOut2[0]))
	Ib = Phasor(complex(vdOut1[1], vdOut2[1]))
	Ic = Phasor(complex(vdOut1[2], vdOut2[2]))
	return
}

// GetSCCurrentSeq gets the short circuit sequence current for the equipment with the provided handle.
// Returns I0, I1, I2 Phasor types. PickFault must be called first.
func (c *Client) GetSCCurrentSeq(hnd int) (I0, I1, I2 Phasor, err error) {
	vdOut1, vdOut2, err := c.olxAPI.GetSCCurrent(hnd, 1)
	if err != nil {
		return I0, I1, I2, err
	}
	I0 = Phasor(complex(vdOut1[0], vdOut2[0]))
	I1 = Phasor(complex(vdOut1[1], vdOut2[1]))
	I2 = Phasor(complex(vdOut1[2], vdOut2[2]))
	return
}

// FullBusName returns the full bus name for the provided handle, returns empty string on error.
func (c *Client) FullBusName(hnd int) string {
	return c.olxAPI.FullBusName(hnd)
}

// FullBranchName returns the full branch name for the provided handle, returns empty string on error.
func (c *Client) FullBranchName(hnd int) string {
	return c.olxAPI.FullBranchName(hnd)
}

// FullRelayName returns the full relay name for the provided handle, returns empty string on error.
func (c *Client) FullRelayName(hnd int) string {
	return c.olxAPI.FullRelayName(hnd)
}
