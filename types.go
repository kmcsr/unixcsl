
package unixcsl

import (
	"bytes"
	"strconv"
)

type TkType int

const (
	TkBytes TkType = iota
	TkCtl
	TkESC
	TkCSI
)

func (t TkType)String()(string){
	switch t {
	case TkBytes: return "BTS"
	case TkCtl: return "CTL"
	case TkESC: return "ESC"
	case TkCSI: return "CSI"
	}
	return "UKN"
}

type Token struct{
	T TkType
	V interface{}
}

func (tk *Token)Bytes()([]byte){
	switch tk.T {
	case TkBytes:
		return tk.V.([]byte)
	case TkCtl:
		return []byte{tk.V.(byte)}
	case TkESC:
		return tk.V.(*ESCSeq).Bytes()
	case TkCSI:
		return tk.V.(*CSISeq).Bytes()
	}
	return nil
}

type ESCSeq struct{
	Ch byte
	Payload byte
}

func (s *ESCSeq)Bytes()([]byte){
	if s.Payload == 0 {
		return []byte{s.Ch}
	}
	return []byte{s.Ch, s.Payload}
}

type CSISeq struct{
	Ch byte
	Args []int
}

func NewCSI(ch byte, args ...int)(*CSISeq){
	return &CSISeq{
		Ch: ch,
		Args: args,
	}
}

func (s *CSISeq)Bytes()([]byte){
	buf := bytes.NewBuffer(make([]byte, 0, 2 + len(s.Args) * 3))
	buf.Write(CSI_SEQ)
	if len(s.Args) > 0 {
		buf.WriteString(strconv.Itoa(s.Args[0]))
		for _, a := range s.Args[1:] {
			buf.WriteByte(';')
			buf.WriteString(strconv.Itoa(a))
		}
	}
	buf.WriteByte(s.Ch)
	return buf.Bytes()
}

func (s *CSISeq)ArgDef(i int, def int)(int){
	if i >= len(s.Args) {
		return def
	}
	return s.Args[i]
}

func (s *CSISeq)Arg(i int)(int){
	return s.ArgDef(i, 0)
}
