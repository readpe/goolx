package main

import (
	"fmt"
	"strings"

	"github.com/readpe/goolx"
)

// formatBus formats a Bus for LLM-readable output.
func formatBus(b *goolx.Bus) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Bus: %s\n", b.Name)
	fmt.Fprintf(&sb, "  Handle:    %d\n", b.Hnd)
	fmt.Fprintf(&sb, "  kV:        %0.2f\n", b.KVNominal)
	fmt.Fprintf(&sb, "  Area:      %d\n", b.Area)
	fmt.Fprintf(&sb, "  Zone:      %d\n", b.Zone)
	fmt.Fprintf(&sb, "  Location:  %s\n", b.Location)
	return sb.String()
}

// formatLine formats a Line for LLM-readable output.
func formatLine(l *goolx.Line) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Line: %s\n", l)
	fmt.Fprintf(&sb, "  Handle:      %d\n", l.Hnd)
	fmt.Fprintf(&sb, "  Name:        %s\n", l.Name)
	fmt.Fprintf(&sb, "  In Service:  %d\n", l.InService)
	fmt.Fprintf(&sb, "  Length:       %0.2f %s\n", l.Length, l.LengthUnit)
	fmt.Fprintf(&sb, "  R/X:         %0.5f / %0.5f pu\n", l.R, l.X)
	fmt.Fprintf(&sb, "  R0/X0:       %0.5f / %0.5f pu\n", l.R0, l.X0)
	if l.Bus1 != nil {
		fmt.Fprintf(&sb, "  Bus1:        %s (handle %d)\n", l.Bus1, l.Bus1.Hnd)
	}
	if l.Bus2 != nil {
		fmt.Fprintf(&sb, "  Bus2:        %s (handle %d)\n", l.Bus2, l.Bus2.Hnd)
	}
	return sb.String()
}

// formatPhasors formats three phase phasors as a single line.
func formatPhasors(label string, a, b, c goolx.Phasor) string {
	return fmt.Sprintf("  %s:  A=%s  B=%s  C=%s", label, a, b, c)
}
