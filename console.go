
package unixcsl

import (
	io "io"
	bufio "bufio"
)

type Console struct{
	Scanner
	Writer *bufio.Writer
}

func NewConsole(r io.Reader, w io.Writer)(*Console){
	return &Console{
		Scanner: *NewScanner(r),
		Writer: bufio.NewWriterSize(w, 2048),
	}
}

func (c *Console)Write(buf []byte)(n int, err error){
	return c.Writer.Write(buf)
}

func (c *Console)Flush()(error){
	return c.Writer.Flush()
}

func (c *Console)WriteByte(b byte)(err error){
	return c.Writer.WriteByte(b)
}

func (c *Console)WriteRune(b rune)(n int, err error){
	return c.Writer.WriteRune(b)
}

func (c *Console)WriteString(str string)(n int, err error){
	return c.Writer.WriteString(str)
}

type immediatelyWriter struct{
	c *Console
}

func (c *Console)ImmediatelyWriter()(immediatelyWriter){
	return immediatelyWriter{c}
}

func (c immediatelyWriter)Write(buf []byte)(n int, err error){
	defer c.c.Flush()
	return c.c.Write(buf)
}

func (c immediatelyWriter)WriteByte(b byte)(err error){
	defer c.c.Flush()
	return c.c.WriteByte(b)
}

func (c immediatelyWriter)WriteRune(b rune)(n int, err error){
	defer c.c.Flush()
	return c.c.WriteRune(b)
}

func (c immediatelyWriter)WriteString(str string)(n int, err error){
	defer c.c.Flush()
	return c.c.WriteString(str)
}

func (c *Console)WriteImmedia(buf []byte)(n int, err error){
	return immediatelyWriter{c}.Write(buf)
}

func (c *Console)WriteCSI(ch byte, args ...int)(n int, err error){
	return c.Write(NewCSI(ch, args...).Bytes())
}

func (c *Console)AskCSI(ch byte)(csi *CSISeq, err error){
	var tk *Token
	tk, err = c.NextFunc(func(tk *Token)(bool){ return tk.t == TkCSI && tk.v.(*CSISeq).Ch == ch })
	if err != nil { return }
	return tk.v.(*CSISeq), nil
}

func (c *Console)Cursor()(p Pos, err error){
	var csi *CSISeq
	_, err = c.WriteCSI(DSR_CH, DSR_CPR)
	if err != nil { return }
	csi, err = c.AskCSI(DSR_CPR_RE)
	if err != nil { return }
	return Pos{
		Y: csi.Args[0],
		X: csi.Args[1],
	}, nil
}

func (c *Console)SetCursor(p Pos)(err error){
	_, err = c.WriteCSI(HVP_CH, p.Y, p.X)
	return
}
