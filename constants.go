
// See <https://man7.org/linux/man-pages/man4/console_codes.4.html>

package unixcsl

const (
	BEL_CH = '\x07'
	BS_CH  = '\x08'
	HT_CH  = '\x09'
	LF_CH  = '\x0a'
	VT_CH  = '\x0b'
	FF_CH  = '\x0c'
	CR_CH  = '\x0d'
	CAN_CH = '\x18'
	SUB_CH = '\x1a'
	ESC_CH = '\x1b'
	DEL_CH = '\x7f'
	CSI_CH = '\x9b'
)

// BEGIN ESC sequences
const (
	RIS_SEQ_L = 'c'
	IND_SEQ_L = 'D'
	NEL_SEQ_L = 'E'
	HTS_SEQ_L = 'H'
	RI_SEQ_L = 'M'
	DECID_SEQ_L = 'Z'
	DECSC_SEQ_L = '7'
	DECRC_SEQ_L = '8'
	CSI_SEQ_L = '['
	DECONM_SEQ_L = '>'
	DECPAM_SEQ_L = '='
	OSC_SEQ_L = ']'
)

var (
	RIS_SEQ = []byte{ESC_CH, RIS_SEQ_L} // Reset
	IND_SEQ = []byte{ESC_CH, IND_SEQ_L} // Linefeed
	NEL_SEQ = []byte{ESC_CH, NEL_SEQ_L} // Newline
	HTS_SEQ = []byte{ESC_CH, HTS_SEQ_L} // Set tab stop at current column.
	RI_SEQ = []byte{ESC_CH, RI_SEQ_L} // Reverse linefeed.
	DECID_SEQ = []byte{ESC_CH, DECID_SEQ_L} // DEC private identification. The kernel returns the string ESC[?6c', claiming that it is a VT102.
	DECSC_SEQ = []byte{ESC_CH, DECSC_SEQ_L} // Save current state
	DECRC_SEQ = []byte{ESC_CH, DECRC_SEQ_L} // Restore state most recently saved by DECSC_SEQ
	CSI_SEQ = []byte{ESC_CH, CSI_SEQ_L} // Control sequence introducer
	DECPNM_SEQ = []byte{ESC_CH, DECONM_SEQ_L} // Set numeric keypad mode
	DECPAM_SEQ = []byte{ESC_CH, DECPAM_SEQ_L} // Set application keypad mode
	OSC_SEQ = []byte{ESC_CH, OSC_SEQ_L} /* (Should be: Operating system command)
		'ESC ] P nrrggbb': set palette, with parameter given in 7 hexadecimal digits after the final P :-(.
		Here n is the color (0–15), and rrggbb indicates the red/green/blue values (0–255).
		'ESC ] R': reset palette
	*/
)

// AFTER ESC sequences

const ( // CSI chars
	ICH_CH = '@'
	CUU_CH = 'A' // Move cursor up the indicated
	CUD_CH = 'B' // Move cursor down the indicated
	CUF_CH = 'C' // Move cursor right the indicated
	CUB_CH = 'D' // Move cursor left the indicated
	CNL_CH = 'E' // Move cursor down the indicated, and go to the colume 1
	CPL_CH = 'F' // Move cursor up the indicated, and go to the colume 1
	CHA_CH = 'G' // Move cursor to indicated column in current row.
	CUP_CH = 'H' // Move cursor to the indicated row, column. default 1, 1
	ED_CH = 'J' /* Erase display (default: from cursor to end of display).
		* 1: erase from start to cursor.
		* 2: erase whole display.
		* 3: erase whole display including scroll-back buffer (since Linux 3.0).
	*/
	EL_CH = 'K' /* Erase line (default: from cursor to end of line).
		* 1: erase from start of line to cursor.
		* 2: erase whole line.
	*/
	IL_CH = 'L' // Insert the indicated. num of blank lines.
	DL_CH = 'M' // Delete the indicated. num of lines.
	DCH_CH = 'P' // Delete the indicated. num of characters on current line.
	ECH_CH = 'X' // Erase the indicated. num of characters on current line.
	HPR_CH = 'a' // Move cursor right the indicated. num of columns.
	DA_CH = 'c' // Answer 'ESC[?6c': "I am a VT102".
	VPA_CH = 'd' // Move cursor to the indicated row, current column.
	VPR_CH = 'e' // Move cursor down the indicated. num of rows.
	HVP_CH = 'f' // Move cursor to the indicated row, column.
	TBC_CH = 'g' // Without parameter: clear tab stop at current position.
	SM_CH = 'h' // Set Mode
	RM_CH = 'l' // Reset Mode
	SGR_CH = 'm' // Set attributes. see bellow 'Select Graphic Rendition'
	DSR_CH = 'n' // Status report. see bellow 'Status Report Commands'
	DECLL_CH = 'q' /* Set keyboard LEDs.
		0: clear all LEDs
		1: set Scroll Lock LED
		2: set Num Lock LED
		3: set Caps Lock LED
	*/
	DECSTBM_CH = 'r' // 
	SAVE_CURSOR_LOCATION_CH = 's' // Save cursor location. (# No short name)
	RESTORE_CURSOR_LOCATION_CH = 'u' // Restore cursor location. (# No short name)
	HPA_CH = '`' // Move cursor to indicated column in current row.
)

const ( // SGR(Select Graphic Rendition)'s parameters
	SGR_RESET = 0 // reset all attributes to their defaults
	SGR_BOLD = 1 // set bold
	SGR_HALF_BRIGHT = 2 // set half-bright
	SGR_UNDERSCORE = 4 // set underscore
	SGR_BLINK = 5 // set blink
	SGR_REVERSE = 7 // reverse video
	SGR_UNDERLINE = 21 // set underline. before Linux 4.17, this value set normal intensity (as is done in many other terminals)
	// TODO
)

const ( // DSR(Status Report Commands)'s parameters
	DSR_DSR = 5 // Device status report (DSR): Answer is 'ESC[0n' (Terminal OK).
	DSR_CPR = 6 // Cursor position report (CPR): Answer is 'ESC[y;xR'
	DSR_CPR_RE = 'R'
)

