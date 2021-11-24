package constants

// Equipment Types
//
// OlxAPI equipment type codes are used to accessing equipment data from Oneliner.
const (
	TCNothing   = 0
	TCBus       = 1
	TCLoad      = 2
	TCLoadUnit  = 3
	TCShunt     = 4
	TCShuntUnit = 5
	TCGen       = 6
	TCGenUnit   = 7
	TCSVD       = 8
	TCBranch    = 9
	TCLine      = 10
	TCXFMR      = 11
	TCXFMR3     = 12
	TCPS        = 13
	TCSCAP      = 14
	TCMU        = 15
	TCArea      = 16
	TCZone      = 17
	TCNote      = 18
	TCSYS       = 19
	TCRLYGroup  = 20
	TCRLYOC     = 21
	TCRLYOCG    = 21
	TCRLYOCP    = 22
	TCRLYDS     = 23
	TCRLYDSG    = 23
	TCRLYDSP    = 24
	TCFuse      = 25
	TCPF        = 26
	TCSC        = 27
	TCSwitch    = 28
	TCRECLSR    = 29
	TCRECLSRP   = 29
	TCRECLSRG   = 30
	TCScheme    = 31
	TCBreaker   = 32
	TCCCGEN     = 33
	TCRLYD      = 34
	TCRLYV      = 35
	TCPILOT     = 36
	TCZCorrect  = 37
	TCBlob      = 38
	TCDCLine2   = 39
	TCLineKink  = 40
	TCRLYLink   = 41
	TCLTC       = 42
	TCLTC3      = 43
	TCSettings  = 44
	TCCount     = 44
	TCPicked    = 100
	TCPicked1   = 100
	TCPicked2   = 101
	TCPicked3   = 102

	// Special obj handles
	HNDSYS = 1
	HNDPF  = 2
	HNDSC  = 3

	OLXAPIFailure = 0
	OLXAPIOk      = 1

	// Fault browser code
	SFLast     = -1
	SFNext     = -2
	SFFirst    = 1
	SFPrevious = -4

	// Array size
	MXMSGLEN   = 512
	MAXPATH    = 260
	MXDSPARAMS = 255
	MXZONE     = 8
	MAXCCV     = 10
	MXSBKF     = 10
)
