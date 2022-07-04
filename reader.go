
package unixcsl

import (
	"io"
	"strconv"
	"sync"
)

type Reader struct{
	r io.Reader

	buf []byte

	mux *sync.Cond
	reading bool

	tkbf []*Token
}

func NewReader(r io.Reader)(*Reader){
	return &Reader{
		r: r,
		mux: sync.NewCond(new(sync.Mutex)),
	}
}

func (r *Reader)Close()(error){
	if c, ok := r.r.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

func (r *Reader)read(l int)(err error){
	if l <= 0 {
		panic(l)
	}
	var (
		buf = make([]byte, l)
		n int
	)
	for len(buf) > 0 {
		n, err = r.r.Read(buf)
		r.buf = append(r.buf, buf[:n]...)
		if err != nil { return }
		buf = buf[n:]
	}
	return
}

func (r *Reader)getByte(i *int)(b byte, err error){
	c := *i
	for c >= len(r.buf) {
		if err = r.read(len(r.buf) - c + 1); err != nil { return }
	}
	b = r.buf[c]
	*i++
	return
}

func (r *Reader)next()(_ *Token, err error){
	var (
		i int = 0
		b byte
	)
	defer func(){
		r.buf = r.buf[i:]
	}()
	b, err = r.getByte(&i)
	if err != nil { return }
	switch {
	case b == ESC_CH:
		return r.parseESC(&i)
	case b == CSI_CH:
		return r.parseCSI(&i)
	case b <= 0x1f || b == DEL_CH:
		return &Token{
			T: TkCtl,
			V: b,
		}, nil
	default:
		for ; i < len(r.buf); i++ {
			if b = r.buf[i]; b == ESC_CH || b == CSI_CH || b <= 0x1f || b == DEL_CH {
				break
			}
		}
		return &Token{
			T: TkBytes,
			V: r.buf[:i],
		}, nil
	}
}

func (r *Reader)parseESC(i *int)(_ *Token, err error){
	var b byte
	if b, err = r.getByte(i); err != nil { return }
	switch b {
	case CSI_SEQ_L:
		return r.parseCSI(i)
	case '%', '#', '(', ')':
		c := b
		b, err = r.getByte(i)
		if err != nil { return }
		return &Token{
			T: TkESC,
			V: &ESCSeq{
				Ch: c,
				Payload: b,
			},
		}, nil
	default:
		return &Token{
			T: TkESC,
			V: &ESCSeq{
				Ch: b,
			},
		}, nil
	}
}

func (r *Reader)parseCSI(i *int)(_ *Token, err error){
	var (
		b byte
		n int
		args []int
	)
	x := *i
	for {
		b, err = r.getByte(i)
		if err != nil { return }
		if '0' > b || b > '9' {
			if b != ';' {
				return &Token{
					T: TkCSI,
					V: &CSISeq{
						Ch: b,
						Args: args,
					},
				}, nil
			}
			if x + 1 == *i {
				n = 0
			}else{
				n, err = strconv.Atoi((string)(r.buf[x:*i]))
				if err != nil { // Unexcept token
					return &Token{
						T: TkBytes,
						V: r.buf[:*i],
					}, nil
				}
			}
			args = append(args, n)
			x = *i
		}
	}
}

func (r *Reader)nextFromBuf(c func(*Token)(bool))(tk *Token){
	if len(r.tkbf) == 0 {
		return nil
	}
	if c == nil {
		tk = r.tkbf[0]
		r.tkbf = r.tkbf[:copy(r.tkbf, r.tkbf[1:])]
		return
	}
	var i int
	for i, tk = range r.tkbf {
		if c(tk) {
			copy(r.tkbf[i:], r.tkbf[i + 1:])
			r.tkbf = r.tkbf[:len(r.tkbf) - 1]
			return
		}
	}
	return nil
}

func (r *Reader)Next()(tk *Token, err error){
	r.mux.L.Lock()
	defer r.mux.L.Unlock()
	for {
		if tk = r.nextFromBuf(nil); tk != nil {
			return
		}
		if !r.reading {
			break
		}
		r.mux.Wait()
	}
	r.reading = true
	r.mux.L.Unlock()
	defer func(){
		r.mux.L.Lock()
		r.reading = false
		r.mux.Broadcast()
	}()
	return r.next()
}

func (r *Reader)NextFunc(c func(*Token)(bool))(tk *Token, err error){
	r.mux.L.Lock()
	defer r.mux.L.Unlock()
	for {
		if tk = r.nextFromBuf(c); tk != nil {
			return
		}
		if !r.reading {
			break
		}
		r.mux.Wait()
	}
	r.reading = true
	defer func(){
		r.reading = false
		r.mux.Broadcast()
	}()
	for {
		r.mux.L.Unlock()
		tk, err = r.next()
		r.mux.L.Lock()
		if err != nil { return }
		if c(tk) { return }
		if l := len(r.tkbf) - 1; l >= 0 && tk.T == TkBytes && r.tkbf[l].T == TkBytes {
			r.tkbf[l].V = append(r.tkbf[l].V.([]byte), tk.V.([]byte)...)
		}else{
			r.tkbf = append(r.tkbf, tk)
		}
		r.mux.Broadcast()
	}
}

func (r *Reader)NextType(t TkType)(*Token, error){
	return r.NextFunc(func(tk *Token)(bool){ return tk.T == t })
}
