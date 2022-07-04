
package unixcsl

import (
	"fmt"
)

type CBreakErr byte

func (e CBreakErr)Error()(string){
	if (byte)(e) == 0 {
		return "NUL"
	}
	if (byte)(e) == DEL_CH {
		return "DEL"
	}
	if e <= 0x1f {
		return fmt.Sprintf("Ctrl-%c", (byte)(e) + 'A' - 1)
	}
	return fmt.Sprintf("CBreak-%x", (byte)(e))
}
