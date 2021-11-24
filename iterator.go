// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package goolx

import "io"

// HandleIterator is a iterator interface for equipment handles.
type HandleIterator interface {
	Next() bool
	Hnd() int
}

// NextEquipment is an equipment handle iterator for getting
// the next equipment handle of the provided type. The Next() method
// will retrieve the next handle if available and populate it for access
// by Hnd(). If Next() returns false, then there was an error or the list
// was exhausted. Once the iterator is exhausted it cannot be reused.
type NextEquipment struct {
	c      *Client
	eqType int
	hnd    int
	done   bool
	err    error
}

// Next retrieves the next equipment handle of type. Returns
// true if successful, and false if not. Hnd() should not be used
// if Next() is false. This can be used in for loops.
func (n *NextEquipment) Next() bool {
	if n.done {
		return false
	}
	err := n.c.olxAPI.GetEquipment(n.eqType, &n.hnd)
	if err != nil {
		n.done = true
		if err == io.EOF {
			// EOF is not an error, so don't set n.err = err.
			return false
		}
		n.err = err
	}
	return true
}

// Hnd returns the current equipment handle, Next() must be called first.
func (n *NextEquipment) Hnd() int {
	return n.hnd
}

// NextEquipmentByTag is an equipment handle iterator for getting
// the next equipment handle of the provided type with the listed tags. The Next() method
// will retrieve the next handle if available and populate it for access
// by Hnd(). If Next() returns false, then there was an error or the list
// was exhausted. Once the iterator is exhausted it cannot be reused.
type NextEquipmentByTag struct {
	c      *Client
	eqType int
	tags   []string
	hnd    int
	err    error
}

// Next retrieves the next equipment handle of type. Returns
// true if successful, and false if not. Hnd() should not be used
// if Next() is false. This can be used in for loops.
func (n *NextEquipmentByTag) Next() bool {
	if n.err != nil {
		return false
	}
	err := n.c.olxAPI.FindEquipmentByTag(n.eqType, &n.hnd, n.tags...)
	if err != nil {
		n.hnd, n.err = 0, err
		return false
	}
	return true
}

// Hnd returns the current equipment handle, Next() must be called first.
func (n *NextEquipmentByTag) Hnd() int {
	return n.hnd
}

// NextRelay is an handle iterator for getting
// the next relay handle in the provided relay group. The Next() method
// will retrieve the next handle if available and populate it for access
// by Hnd(). If Next() returns false, then there was an error or the list
// was exhausted. Once the iterator is exhausted it cannot be reused.
type NextRelay struct {
	c           *Client
	rlyGroupHnd int
	hnd         int
	done        bool
	err         error
}

// Next retrieves the next relay handle int he relay group. Returns
// true if successful, and false if not. Hnd() should not be used
// if Next() is false. This can be used in for loops.
func (n *NextRelay) Next() bool {
	if n.done {
		return false
	}
	err := n.c.olxAPI.GetRelay(n.rlyGroupHnd, &n.hnd)
	if err != nil {
		n.done = true
		if err == io.EOF {
			// EOF is not an error, so don't set n.err = err.
			return false
		}
		n.err = err
	}
	return true
}

// Hnd returns the current relay handle, Next() must be called first.
func (n *NextRelay) Hnd() int {
	return n.hnd
}
