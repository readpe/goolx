// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package goolx

import (
	"fmt"
	"strings"
)

// FindLine searches for a branch with the given branch data, returns the branch handle.
// Returns error if a branch cannot be found.
func (c *Client) FindBranch(fName string, fKV float64, tName string, tKV float64, ckt string) (int, error) {
	return c.findBranch(fName, fKV, tName, tKV, ckt)
}

// FindLine searches for a line with the given branch data. From and To can be swapped and should return the same Line object.
// Returns error if a line cannot be found, or if the branch specified points to a non-line object.
func (c *Client) FindLine(fName string, fKV float64, tName string, tKV float64, ckt string) (*Line, error) {
	brHnd, err := c.findBranch(fName, fKV, tName, tKV, ckt)
	if err != nil {
		return nil, fmt.Errorf("FindLine: could not find line: %v", err)
	}

	var lineHnd int
	if c.GetData(brHnd, BRnHandle).Scan(&lineHnd); err != nil {
		return nil, fmt.Errorf("FindLine: could not find line: %v", err)
	}

	if eqType, _ := c.EquipmentType(lineHnd); eqType != TCLine {
		return nil, fmt.Errorf("FindLine: branch is not of type TCLine %v", err)
	}

	return c.getLine(lineHnd)
}

func (c *Client) findBranch(fName string, fKV float64, tName string, tKV float64, ckt string) (int, error) {
	fHnd, err := c.FindBusByName(fName, fKV)
	if err != nil {
		return 0, err
	}

	tHnd, err := c.FindBusByName(tName, tKV)
	if err != nil {
		return 0, err
	}

	for bi := c.NextBusEquipment(fHnd, TCBranch); bi.Next(); {
		brHnd := bi.Hnd()

		var brBus2Hnd int
		var brEqHnd int
		if err := c.GetData(brHnd, BRnBus2Hnd, BRnHandle).Scan(&brBus2Hnd, &brEqHnd); err != nil {
			return 0, err
		}

		// Get branch equipment type in order to obtain circuit ID.
		brEqType, err := c.EquipmentType(brEqHnd)
		if err != nil {
			return 0, err
		}

		// Determin ckt id code dependent on equipment type.
		var sID int
		switch brEqType {
		case TCLine:
			sID = LNsID
		case TCXFMR:
			sID = XRsID
		case TCXFMR3:
			sID = X3sID
		case TCPS:
			sID = PSsID
		case TCSC:
			sID = SCsID
		case TCSwitch:
			sID = SWsID
		default:
			return 0, fmt.Errorf("findBranch: %s %0.2fkV-%s %0.2fkV ckt:%s unsupported equipment type %d", fName, fKV, tName, tKV, ckt, brEqHnd)
		}

		var cktID string
		if err := c.GetData(brEqHnd, sID).Scan(&cktID); err != nil {
			return 0, err
		}

		// Check to bus and ckt match, if match then branch found.
		if brBus2Hnd == tHnd && strings.TrimSpace(cktID) == strings.TrimSpace(ckt) {
			return brHnd, nil
		}

	}
	return 0, fmt.Errorf("findBranch: could not find %s %0.2fkV-%s %0.2fkV ckt:%s", fName, fKV, tName, tKV, ckt)
}
