// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package goolx

// Run1LPFCommand runs a Oneliner command using xml input string.
func (c *Client) Run1LPFCommand(s string) error {
	return c.olxAPI.Run1LPFCommand(s)
}
