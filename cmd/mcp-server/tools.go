package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/readpe/goolx"
)

// equipmentTypes maps user-facing type names to goolx type codes.
var equipmentTypes = map[string]int{
	"bus":    goolx.TCBus,
	"line":   goolx.TCLine,
	"xfmr":   goolx.TCXFMR,
	"xfmr3":  goolx.TCXFMR3,
	"gen":    goolx.TCGen,
	"load":   goolx.TCLoad,
	"shunt":  goolx.TCShunt,
	"svd":    goolx.TCSVD,
	"ps":     goolx.TCPS,
	"scap":   goolx.TCSCAP,
	"switch": goolx.TCSwitch,
	"fuse":   goolx.TCFuse,
	"rlygrp": goolx.TCRLYGroup,
}

func textResult(text string) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}, nil, nil
}

func errResult(err error) (*mcp.CallToolResult, any, error) {
	return nil, nil, err
}

// registerTools registers all MCP tools on the server.
func registerTools(server *mcp.Server, client *goolx.Client) {
	registerLoadCase(server, client)
	registerGetSystemInfo(server, client)
	registerFindBus(server, client)
	registerListEquipment(server, client)
	registerGetEquipment(server, client)
	registerRunFault(server, client)
	registerRunSteppedEvent(server, client)
}

// --- load_case ---

type loadCaseInput struct {
	FilePath string `json:"file_path,omitempty" jsonschema:"description=Path to .olr file to load (read-only). Omit to check what is currently loaded."`
}

func registerLoadCase(server *mcp.Server, client *goolx.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "load_case",
		Description: "Load an ASPEN Oneliner .olr case file (read-only) or check what is currently loaded.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in loadCaseInput) (*mcp.CallToolResult, any, error) {
		if in.FilePath == "" {
			fn := client.GetOlrFilename()
			if fn == "" {
				return textResult("No case file is currently loaded.")
			}
			return textResult(fmt.Sprintf("Currently loaded: %s\n%s", fn, client.Info()))
		}
		if err := client.LoadDataFileReadOnly(in.FilePath); err != nil {
			return errResult(fmt.Errorf("failed to load case: %w", err))
		}
		return textResult(fmt.Sprintf("Loaded: %s\n%s", in.FilePath, client.Info()))
	})
}

// --- get_system_info ---

type getSystemInfoInput struct{}

func registerGetSystemInfo(server *mcp.Server, client *goolx.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_system_info",
		Description: "Get system-level summary of the loaded case: equipment counts and base MVA.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in getSystemInfoInput) (*mcp.CallToolResult, any, error) {
		var baseMVA float64
		var nBus, nGen, nLoad, nShunt, nLine, nScap, nXfmr, nXfmr3, nPS, nMu int

		if err := client.GetData(goolx.HNDSYS, goolx.SYdBaseMVA).Scan(&baseMVA); err != nil {
			return errResult(fmt.Errorf("no case loaded or cannot read system data: %w", err))
		}

		client.GetData(goolx.HNDSYS, goolx.SYnNObus).Scan(&nBus)
		client.GetData(goolx.HNDSYS, goolx.SYnNOgen).Scan(&nGen)
		client.GetData(goolx.HNDSYS, goolx.SYnNOload).Scan(&nLoad)
		client.GetData(goolx.HNDSYS, goolx.SYnNOshunt).Scan(&nShunt)
		client.GetData(goolx.HNDSYS, goolx.SYnNOline).Scan(&nLine)
		client.GetData(goolx.HNDSYS, goolx.SYnNOseriescap).Scan(&nScap)
		client.GetData(goolx.HNDSYS, goolx.SYnNOxfmr).Scan(&nXfmr)
		client.GetData(goolx.HNDSYS, goolx.SYnNOxfmr3).Scan(&nXfmr3)
		client.GetData(goolx.HNDSYS, goolx.SYnNOps).Scan(&nPS)
		client.GetData(goolx.HNDSYS, goolx.SYnNOmutual).Scan(&nMu)

		var sb strings.Builder
		fmt.Fprintf(&sb, "System Info\n")
		fmt.Fprintf(&sb, "  Base MVA:        %0.1f\n", baseMVA)
		fmt.Fprintf(&sb, "  Buses:           %d\n", nBus)
		fmt.Fprintf(&sb, "  Generators:      %d\n", nGen)
		fmt.Fprintf(&sb, "  Loads:           %d\n", nLoad)
		fmt.Fprintf(&sb, "  Shunts:          %d\n", nShunt)
		fmt.Fprintf(&sb, "  Lines:           %d\n", nLine)
		fmt.Fprintf(&sb, "  Series Caps:     %d\n", nScap)
		fmt.Fprintf(&sb, "  Transformers:    %d\n", nXfmr)
		fmt.Fprintf(&sb, "  3W Transformers: %d\n", nXfmr3)
		fmt.Fprintf(&sb, "  Phase Shifters:  %d\n", nPS)
		fmt.Fprintf(&sb, "  Mutual Pairs:    %d\n", nMu)

		return textResult(sb.String())
	})
}

// --- find_bus ---

type findBusInput struct {
	Name   string  `json:"name,omitempty" jsonschema:"description=Bus name to search for."`
	KV     float64 `json:"kv,omitempty" jsonschema:"description=Bus nominal kV (required with name)."`
	Number int     `json:"number,omitempty" jsonschema:"description=Bus number to search for (alternative to name+kv)."`
}

func registerFindBus(server *mcp.Server, client *goolx.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "find_bus",
		Description: "Look up a bus by name+kV or by bus number. Returns bus details and handle for use in other tools.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in findBusInput) (*mcp.CallToolResult, any, error) {
		var hnd int
		var err error

		if in.Number > 0 {
			hnd, err = client.FindBusNo(in.Number)
		} else if in.Name != "" {
			hnd, err = client.FindBusByName(in.Name, in.KV)
		} else {
			return errResult(fmt.Errorf("provide either 'name'+'kv' or 'number'"))
		}
		if err != nil {
			return errResult(err)
		}

		bus, err := client.GetBus(hnd)
		if err != nil {
			return errResult(err)
		}

		return textResult(formatBus(bus))
	})
}

// --- list_equipment ---

type listEquipmentInput struct {
	Type      string `json:"type" jsonschema:"description=Equipment type: bus line xfmr xfmr3 gen load shunt svd ps scap switch fuse rlygrp"`
	BusHandle int    `json:"bus_handle,omitempty" jsonschema:"description=Filter to equipment at this bus handle."`
	Limit     int    `json:"limit,omitempty" jsonschema:"description=Max items to return (default 50)."`
}

func registerListEquipment(server *mcp.Server, client *goolx.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_equipment",
		Description: "List equipment in the loaded case, optionally filtered by bus. Returns handles and names.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in listEquipmentInput) (*mcp.CallToolResult, any, error) {
		tc, ok := equipmentTypes[strings.ToLower(in.Type)]
		if !ok {
			return errResult(fmt.Errorf("unknown equipment type %q, valid types: bus line xfmr xfmr3 gen load shunt svd ps scap switch fuse rlygrp", in.Type))
		}

		limit := in.Limit
		if limit <= 0 {
			limit = 50
		}

		var iter goolx.HandleIterator
		if in.BusHandle > 0 {
			// For bus-level queries, use branch iterator for branch-like types
			branchType := tc
			if tc == goolx.TCLine || tc == goolx.TCXFMR || tc == goolx.TCXFMR3 || tc == goolx.TCPS || tc == goolx.TCSwitch || tc == goolx.TCSCAP {
				branchType = goolx.TCBranch
			}
			iter = client.NextBusEquipment(in.BusHandle, branchType)
		} else {
			iter = client.NextEquipment(tc)
		}

		var sb strings.Builder
		count := 0
		total := 0
		for iter.Next() {
			total++
			if count >= limit {
				continue // keep counting total
			}
			hnd := iter.Hnd()

			var name string
			if tc == goolx.TCBus {
				name = client.FullBusName(hnd)
			} else if in.BusHandle > 0 {
				// For bus equipment queries using branch iterator, show branch name
				name = client.FullBranchName(hnd)
			} else {
				name = client.FullBranchName(hnd)
				if name == "" {
					name = client.FullBusName(hnd)
				}
			}

			fmt.Fprintf(&sb, "  [%d] %s\n", hnd, name)
			count++
		}

		header := fmt.Sprintf("Equipment type: %s", in.Type)
		if total > limit {
			header += fmt.Sprintf(" (showing %d of %d)", limit, total)
		} else {
			header += fmt.Sprintf(" (%d found)", total)
		}
		return textResult(header + "\n" + sb.String())
	})
}

// --- get_equipment ---

type getEquipmentInput struct {
	Handle      int     `json:"handle,omitempty" jsonschema:"description=Equipment handle (from find_bus or list_equipment)."`
	FromBusName string  `json:"from_bus_name,omitempty" jsonschema:"description=From bus name (for branch search)."`
	FromBusKV   float64 `json:"from_bus_kv,omitempty" jsonschema:"description=From bus kV (for branch search)."`
	ToBusName   string  `json:"to_bus_name,omitempty" jsonschema:"description=To bus name (for branch search)."`
	ToBusKV     float64 `json:"to_bus_kv,omitempty" jsonschema:"description=To bus kV (for branch search)."`
	CircuitID   string  `json:"circuit_id,omitempty" jsonschema:"description=Circuit ID (for branch search)."`
}

func registerGetEquipment(server *mcp.Server, client *goolx.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_equipment",
		Description: "Get detailed data for a specific piece of equipment by handle or by branch search (from/to bus + circuit ID).",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in getEquipmentInput) (*mcp.CallToolResult, any, error) {
		// Branch search path
		if in.FromBusName != "" && in.ToBusName != "" {
			line, err := client.FindLine(in.FromBusName, in.FromBusKV, in.ToBusName, in.ToBusKV, in.CircuitID)
			if err != nil {
				return errResult(err)
			}
			return textResult(formatLine(line))
		}

		if in.Handle == 0 {
			return errResult(fmt.Errorf("provide 'handle' or branch search fields (from_bus_name, from_bus_kv, to_bus_name, to_bus_kv, circuit_id)"))
		}

		eqType, err := client.EquipmentType(in.Handle)
		if err != nil {
			return errResult(fmt.Errorf("invalid handle %d: %w", in.Handle, err))
		}

		switch eqType {
		case goolx.TCBus:
			bus, err := client.GetBus(in.Handle)
			if err != nil {
				return errResult(err)
			}
			return textResult(formatBus(bus))
		case goolx.TCLine:
			line, err := client.GetLine(in.Handle)
			if err != nil {
				return errResult(err)
			}
			return textResult(formatLine(line))
		default:
			// Generic fallback: show full name and basic data
			name := client.FullBranchName(in.Handle)
			if name == "" {
				name = client.FullBusName(in.Handle)
			}
			return textResult(fmt.Sprintf("Equipment (type code %d): %s\n  Handle: %d", eqType, name, in.Handle))
		}
	})
}

// --- run_fault ---

type runFaultInput struct {
	BusHandle           int     `json:"bus_handle" jsonschema:"description=Bus handle to fault on (from find_bus)."`
	FaultType           string  `json:"fault_type,omitempty" jsonschema:"description=Fault type: 3LG 1LG LL LLG (default 3LG)."`
	Location            string  `json:"location,omitempty" jsonschema:"description=Fault location: close_in remote_bus line_end intermediate (default close_in)."`
	IntermediatePercent float64 `json:"intermediate_percent,omitempty" jsonschema:"description=Percent along line for intermediate faults (0-100)."`
	FaultR              float64 `json:"fault_r,omitempty" jsonschema:"description=Fault resistance in ohms (default 0)."`
	FaultX              float64 `json:"fault_x,omitempty" jsonschema:"description=Fault reactance in ohms (default 0)."`
	Tiers               int     `json:"tiers,omitempty" jsonschema:"description=Number of tiers for result collection (default 5)."`
}

func registerRunFault(server *mcp.Server, client *goolx.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "run_fault",
		Description: "Run a fault simulation on a bus and return all fault currents and voltages.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in runFaultInput) (*mcp.CallToolResult, any, error) {
		if in.BusHandle == 0 {
			return errResult(fmt.Errorf("bus_handle is required"))
		}

		// Build fault connection
		var conn goolx.FltConn
		switch strings.ToUpper(in.FaultType) {
		case "", "3LG":
			conn = goolx.ABC
		case "1LG":
			conn = goolx.AG
		case "LL":
			conn = goolx.AB
		case "LLG":
			conn = goolx.ABG
		default:
			return errResult(fmt.Errorf("unknown fault_type %q, use: 3LG 1LG LL LLG", in.FaultType))
		}

		// Build fault location option
		var locOpt goolx.FaultOption
		switch strings.ToLower(in.Location) {
		case "", "close_in":
			locOpt = goolx.FaultCloseIn()
		case "remote_bus":
			locOpt = goolx.FaultRemoteBus()
		case "line_end":
			locOpt = goolx.FaultLineEnd()
		case "intermediate":
			pct := in.IntermediatePercent
			if pct <= 0 {
				pct = 50
			}
			locOpt = goolx.FaultIntermediate(pct)
		default:
			return errResult(fmt.Errorf("unknown location %q, use: close_in remote_bus line_end intermediate", in.Location))
		}

		opts := []goolx.FaultOption{
			goolx.FaultConn(conn),
			locOpt,
			goolx.FaultClearPrev(true),
		}
		if in.FaultR != 0 || in.FaultX != 0 {
			opts = append(opts, goolx.FaultRX(in.FaultR, in.FaultX))
		}

		cfg := goolx.NewFaultConfig(opts...)
		if err := client.DoFault(in.BusHandle, cfg); err != nil {
			return errResult(fmt.Errorf("DoFault failed: %w", err))
		}

		tiers := in.Tiers
		if tiers <= 0 {
			tiers = 5
		}

		var sb strings.Builder
		sb.WriteString("Fault Results\n")
		sb.WriteString(strings.Repeat("-", 60) + "\n")

		for fi := client.NextFault(tiers); fi.Next(); {
			idx := fi.Index()
			desc := client.FaultDescription(idx)
			fmt.Fprintf(&sb, "\n%s\n", desc)

			// Total fault current
			ia, ib, ic, err := client.GetSCCurrentPhase(goolx.HNDSC)
			if err != nil {
				fmt.Fprintf(&sb, "  (error reading current: %v)\n", err)
				continue
			}
			sb.WriteString(formatPhasors("I(total)", ia, ib, ic) + "\n")

			// Bus voltage at faulted bus
			va, vb, vc, err := client.GetSCVoltagePhase(in.BusHandle)
			if err == nil {
				sb.WriteString(formatPhasors("V(bus)  ", va, vb, vc) + "\n")
			}
		}

		return textResult(sb.String())
	})
}

// --- run_stepped_event ---

type runSteppedEventInput struct {
	BusHandle int    `json:"bus_handle" jsonschema:"description=Bus handle to run stepped event on."`
	FaultType string `json:"fault_type,omitempty" jsonschema:"description=Fault type: 3LG 1LG LL LLG (default 3LG)."`
	Location  string `json:"location,omitempty" jsonschema:"description=Fault location: close_in or intermediate percent (default close_in)."`
	Tiers     int    `json:"tiers,omitempty" jsonschema:"description=Number of tiers (default 3)."`
}

func registerRunSteppedEvent(server *mcp.Server, client *goolx.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "run_stepped_event",
		Description: "Run stepped event (protection coordination) analysis and return relay operation timeline.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, in runSteppedEventInput) (*mcp.CallToolResult, any, error) {
		if in.BusHandle == 0 {
			return errResult(fmt.Errorf("bus_handle is required"))
		}

		var conn goolx.FltConn
		switch strings.ToUpper(in.FaultType) {
		case "", "3LG":
			conn = goolx.ABC
		case "1LG":
			conn = goolx.AG
		case "LL":
			conn = goolx.AB
		case "LLG":
			conn = goolx.ABG
		default:
			return errResult(fmt.Errorf("unknown fault_type %q, use: 3LG 1LG LL LLG", in.FaultType))
		}

		tiers := in.Tiers
		if tiers <= 0 {
			tiers = 3
		}

		opts := []goolx.SteppedEventOption{
			goolx.SteppedEventConn(conn),
			goolx.SteppedEventAll(),
			goolx.SteppedEventTiers(tiers),
		}

		if strings.ToLower(in.Location) == "close_in" || in.Location == "" {
			opts = append(opts, goolx.SteppedEventCloseIn())
		}

		cfg := goolx.NewSteppedEvent(opts...)
		if err := client.DoSteppedEvent(in.BusHandle, cfg); err != nil {
			return errResult(fmt.Errorf("DoSteppedEvent failed: %w", err))
		}

		var sb strings.Builder
		sb.WriteString("Stepped Event Results\n")
		sb.WriteString(fmt.Sprintf("%-6s %-10s %-12s %-40s %s\n", "Step", "Time(s)", "Current(A)", "Event", "Fault"))
		sb.WriteString(strings.Repeat("-", 100) + "\n")

		for sei := client.NextSteppedEvent(); sei.Next(); {
			se := sei.Data()
			fmt.Fprintf(&sb, "%-6d %-10.4f %-12.1f %-40s %s\n",
				se.Step, se.Time, se.Current, se.EventDescription, se.FaultDescription)
		}

		return textResult(sb.String())
	})
}
