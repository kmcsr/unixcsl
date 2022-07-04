
package unixcsl

import (
	"io"
	"os"
	"fmt"
	"sync"
	"unicode/utf8"

	term "golang.org/x/term"
)

func MakeRaw(f *os.File)(func()(error), error){
	fd := (int)(f.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	return func()(error){
		return term.Restore(fd, oldState)
	}, nil
}

type Console struct{
	Reader
	io.Writer

	writeLock sync.Mutex

	prompt string

	current []byte
	cursor int
	histories []string
	hist_i int
}

func NewConsole(r io.Reader, w io.Writer)(*Console){
	return &Console{
		Reader: *NewReader(r),
		Writer: w,
		current: make([]byte, 0),
	}
}

func (c *Console)WriteByte(b byte)(err error){
	if bw, ok := c.Writer.(io.ByteWriter); ok {
		return bw.WriteByte(b)
	}
	_, err = c.Write([]byte{b})
	return
}

func (c *Console)WriteRune(b rune)(n int, err error){
	if (uint32)(b) < utf8.RuneSelf {
		err = c.WriteByte((byte)(b))
		if err != nil {
			return 0, err
		}
		return 1, nil
	}
	return c.Write(utf8.AppendRune(nil, b))
}

func (c *Console)WriteString(str string)(n int, err error){
	return io.WriteString(c.Writer, str)
}

func (c *Console)Print(args ...interface{})(n int, err error){
	return fmt.Fprint(c.Writer, args...)
}

func (c *Console)Println(args ...interface{})(n int, err error){
	return fmt.Fprintln(c.Writer, append(args, "\r")...)
}

func (c *Console)Printf(format string, args ...interface{})(n int, err error){
	return fmt.Fprintf(c.Writer, format, args...)
}

func (c *Console)Bell()(err error){
	return c.WriteByte(BEL_CH)
}

func (c *Console)WriteCSI(ch byte, args ...int)(n int, err error){
	return c.Write(NewCSI(ch, args...).Bytes())
}

func (c *Console)AskCSI(ch byte)(csi *CSISeq, err error){
	var tk *Token
	tk, err = c.NextFunc(func(tk *Token)(bool){ return tk.T == TkCSI && tk.V.(*CSISeq).Ch == ch })
	if err != nil { return }
	return tk.V.(*CSISeq), nil
}

func (c *Console)Cursor()(y, x int, err error){
	_, err = c.WriteCSI(DSR_CH, DSR_CPR)
	if err != nil { return }
	var csi *CSISeq
	csi, err = c.AskCSI(DSR_CPR_RE)
	if err != nil { return }
	return csi.Args[0], csi.Args[1], nil
}

func (c *Console)SetCursor(y, x int)(err error){
	_, err = c.WriteCSI(HVP_CH, y, x)
	return
}

func (c *Console)Prompt()(string){
	return c.prompt
}

func (c *Console)SetPrompt(p string){
	c.prompt = p
}

func (c *Console)ReadLine()(line []byte, err error){
	_, err = c.WriteCSI(EL_CH, 2)
	if err != nil { return }
	_, err = c.WriteCSI(CHA_CH, 1)
	if err != nil { return }
	_, err = c.WriteString(c.prompt)
	if err != nil { return }
	for {
		var tk *Token
		tk, err = c.Next()
		if err != nil { return }
		switch tk.T {
		case TkBytes:
			v := tk.V.([]byte)
			if err = c._insert(v); err != nil { return }
		case TkCtl:
			v := tk.V.(byte)
			switch v {
			case DEL_CH:
				if err = c._remove(); err != nil { return }
			case HT_CH:
				if err = c._tab(); err != nil { return }
			case LF_CH, CR_CH:
				if _, err = c.Write(NEW_LINE_SEQ); err != nil { return }
				line = make([]byte, len(c.current))
				copy(line, c.current)
				c.current = c.current[:0]
				if sl := (string)(line); len(sl) != 0 && (len(c.histories) == 0 || c.histories[len(c.histories) - 1] != sl) {
					c.histories = append(c.histories, sl)
				}
				c.hist_i = len(c.histories)
				c.SetLine(nil)
				return
			default:
				return nil, (CBreakErr)(v)
			}
		case TkCSI:
			csi := tk.V.(*CSISeq)
			switch csi.Ch {
			case CUU_CH:
				if err = c._move_up(csi.ArgDef(0, 1)); err != nil { return }
			case CUD_CH:
				if err = c._move_down(csi.ArgDef(0, 1)); err != nil { return }
			case CUF_CH:
				if err = c._move_right(csi.ArgDef(0, 1)); err != nil { return }
			case CUB_CH:
				if err = c._move_left(csi.ArgDef(0, 1)); err != nil { return }
			case CNL_CH:
				if err = c._move_down(csi.ArgDef(0, 1)); err != nil { return }
				if err = c._move_to(0); err != nil { return }
			case CPL_CH:
				if err = c._move_up(csi.ArgDef(0, 1)); err != nil { return }
				if err = c._move_to(0); err != nil { return }
			case CHA_CH:
				if err = c._move_to(csi.Arg(0) - 1); err != nil { return }
			}
		}
	}
}

func (c *Console)SetLine(line []byte){
	if len(line) > 0 {
		copy(c.current, line)
		if len(line) > len(c.current) {
			c.current = append(c.current, line[len(c.current):]...)
		}else{
			c.current = c.current[:len(line)]
		}
	}else{
		c.current = c.current[:0]
	}
	c.cursor = len(c.current)
	return
}

func (c *Console)_set_line(line []byte)(err error){
	_, err = c.WriteCSI(EL_CH, 2)
	if err != nil { return }
	_, err = c.WriteCSI(CHA_CH, 1)
	if err != nil { return }
	_, err = c.WriteString(c.prompt)
	if err != nil { return }
	if len(line) > 0 {
		_, err = c.Write(line)
		if err != nil { return }
	}
	c.SetLine(line)
	return
}

func (c *Console)_move_to(n int)(err error){
	if n > len(c.current) {
		n = len(c.current)
	}
	_, err = c.WriteCSI(CHA_CH, len(c.prompt) + n + 1)
	if err != nil { return }
	c.cursor = n
	return
}

func (c *Console)_move_up(n int)(err error){
	if c.hist_i == 0 {
		return c.Bell()
	}
	if c.hist_i < n {
		n = c.hist_i
	}
	c.hist_i -= n
	c._set_line(([]byte)(c.histories[c.hist_i]))
	return
}

func (c *Console)_move_down(n int)(err error){
	f := len(c.histories) - c.hist_i
	if f == 0 {
		return c.Bell()
	}
	if f < n {
		n = f
	}
	c.hist_i += n
	if c.hist_i == len(c.histories) {
		c._set_line(nil)
	}else{
		c._set_line(([]byte)(c.histories[c.hist_i]))
	}
	return
}

func (c *Console)_move_right(n int)(err error){
	f := len(c.current) - c.cursor
	if f == 0 {
		return c.Bell()
	}
	if f < n {
		n = f
	}
	var l int
	for ; n > 0 && c.cursor < len(c.current); n -= l {
		_, l = utf8.DecodeRune(c.current[c.cursor:])
		_, err = c.WriteCSI(CUF_CH, l)
		if err != nil { return }
		c.cursor += l
	}
	return
}

func (c *Console)_move_left(n int)(err error){
	if c.cursor == 0 {
		return c.Bell()
	}
	if c.cursor < n {
		n = c.cursor
	}
	var l int
	for ; n > 0 && c.cursor > 0; n -= l {
		_, l = utf8.DecodeLastRune(c.current[:c.cursor])
		_, err = c.WriteCSI(CUB_CH, l)
		if err != nil { return }
		c.cursor -= l
	}
	return
}

func (c *Console)_insert(buf []byte)(err error){
	tail := c.current[c.cursor:]
	cr := c.cursor + len(buf)
	nb := make([]byte, len(c.current) + len(buf))
	copy(nb[:c.cursor], c.current[:c.cursor])
	copy(nb[c.cursor:cr], buf)
	copy(nb[cr:], tail)
	c.current = nb
	if _, err = c.Write(buf); err != nil { return }
	if _, err = c.Write(tail); err != nil { return }
	if err = c._move_to(cr); err != nil { return }
	return
}

func (c *Console)_remove()(err error){
	if c.cursor == 0 {
		return c.Bell()
	}
	_, l := utf8.DecodeLastRune(c.current[:c.cursor])
	cr := c.cursor - l
	buf := make([]byte, len(c.current) - l)
	copy(buf[:cr], c.current[:cr])
	copy(buf[cr:], c.current[c.cursor:])
	if err = c._set_line(buf); err != nil { return }
	if err = c._move_to(cr); err != nil { return }
	return
}

func (c *Console)_tab()(err error){
	return c.Bell()
}
