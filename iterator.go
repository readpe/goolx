// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package goolx

// HandleIterator is a iterator interface for equipment handles.
type HandleIterator interface {
	Next() bool
	Hnd() int
}

type handleIterator struct {
	hnd  int
	done bool
	f    func(hnd *int) error
}

// Next advances to the next available handle, returns false if iteration exhausted.
func (h *handleIterator) Next() bool {
	if h.done {
		return false
	}
	if err := h.f(&h.hnd); err != nil {
		h.done = true
		return false
	}
	return true
}

// Hnd returns the currently selected equipment handle.
func (h *handleIterator) Hnd() int {
	return h.hnd
}

// FaultIterator is a fault result iterator for iterating through the available fault results,
// utilizing the PickFault function.
type FaultIterator interface {
	Next() bool
	Index() int
}

type faultIterator struct {
	i    int
	done bool
	f    func(idx *int) error
}

// Next advances to the next fault available.
func (f *faultIterator) Next() bool {
	if f.done {
		return false
	}

	if err := f.f(&f.i); err != nil {
		f.done = true
		return false
	}

	return true
}

// Index returns the index number of the currently selected fault.
func (f *faultIterator) Index() int {
	return f.i
}

// SteppedEventIterator is a stepped event result iterator for iterating through the available fault results,
// utilizing the GetSteppedEvent function.
type SteppedEventIterator interface {
	Next() bool
	Data() SteppedEvent
}

type steppedEventIterator struct {
	step int
	done bool
	data SteppedEvent
	f    func(step *int) (SteppedEvent, error)
}

// Next retrieves the next step available in the stepped event result. If no step is available, returns false.
func (s *steppedEventIterator) Next() bool {
	if s.done {
		return false
	}
	data, err := s.f(&s.step)
	if err != nil {
		s.done = true
		return false
	}
	s.data = data
	return true
}

// Data returns the current step data for the stepped event.
func (s *steppedEventIterator) Data() SteppedEvent {
	return s.data
}
