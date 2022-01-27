
package unixcsl

import (
	io "io"
	strconv "strconv"
	sync "sync"
)

type Scanner struct{
	r io.Reader

	_rbuf []byte
	buf []byte
	mux *sync.Cond
	reading bool
	tkbf []*Token
}

func NewScanner(r io.Reader)(s *Scanner){
	s = &Scanner{
		r: r,
		_rbuf: make([]byte, 512),
		mux: sync.NewCond(new(sync.Mutex)),
		tkbf: make([]*Token, 0, 8),
	}
	return
}

func (s *Scanner)Close()(error){
	if c, ok := s.r.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

func (s *Scanner)read()(err error){
	var n int
	n, err = s.r.Read(s._rbuf)
	if err != nil { return }
	s.buf = catbytes(s.buf, s._rbuf[:n])
	return
}

func (s *Scanner)peek(i int)(b byte, ok bool){
	if i >= len(s.buf) {
		return 0, false
	}
	return s.buf[i], true
}

func (s *Scanner)nextByte(i *int)(b byte, err error){
	for *i >= len(s.buf) {
		err = s.read()
		if err != nil { return }
	}
	b = s.buf[*i]
	*i++
	return
}

func (s *Scanner)next()(tk *Token, err error){
	var (
		i int = 0
		b byte
		n int
	)
	defer func(){
		s.buf = s.buf[i:]
	}()
	b, err = s.nextByte(&i)
	if err != nil { return }
	switch {
	case b == ESC_CH:
		b, err = s.nextByte(&i)
		if err != nil { return }
		if b != CSI_SEQ_L {
			switch b {
			case '%', '#', '(', ')':
				c := b
				b, err = s.nextByte(&i)
				if err != nil { return }
				return &Token{
					t: TkESC,
					v: &ESCSeq{
						Ch: c,
						Payload: b,
					},
				}, nil
			default:
				return &Token{
					t: TkESC,
					v: &ESCSeq{
						Ch: b,
					},
				}, nil
			}
			break
		}
		fallthrough
	case b == CSI_CH:
		x := i
		args := make([]int, 0)
		for {
			b, err = s.nextByte(&i)
			if err != nil { return }
			if '0' > b || b > '9' {
				if b != ';' {
					return &Token{
						t: TkCSI,
						v: &CSISeq{
							Ch: b,
							Args: args,
						},
					}, nil
				}
				if x + 1 == i {
					n = 0
				}else{
					n, err = strconv.Atoi((string)(s.buf[x:i]))
					if err != nil { return }
				}
				args = append(args, n)
				x = i
			}
		}
	case (0 <= b && b <= 0x1b) || b == DEL_CH:
		return &Token{
			t: TkCtl,
			v: b,
		}, nil
	default:
		for {
			b, k := s.peek(i)
			if !k || (b == ESC_CH || b == CSI_CH || (0 <= b && b <= 0x1b) || b == DEL_CH) {
				break
			}
			i++
		}
		return &Token{
			t: TkBytes,
			v: s.buf[:i],
		}, nil
	}
	return nil, NewUnexpectByteError(b)
}

func (s *Scanner)nextFromBuf(c func(*Token)(bool))(tk *Token){
	if len(s.tkbf) == 0 {
		return nil
	}
	if c == nil {
		tk = s.tkbf[0]
		s.tkbf = s.tkbf[:copy(s.tkbf, s.tkbf[1:])]
		return
	}
	var i int
	for i, tk = range s.tkbf {
		if c(tk) {
			copy(s.tkbf[i:], s.tkbf[i + 1:])
			s.tkbf = s.tkbf[:len(s.tkbf) - 1]
			return
		}
	}
	return nil
}

func (s *Scanner)Next()(tk *Token, err error){
	s.mux.L.Lock()
	defer s.mux.L.Unlock()
	for {
		if tk = s.nextFromBuf(nil); tk != nil {
			return
		}
		if !s.reading {
			break
		}
		s.mux.Wait()
	}
	s.reading = true
	s.mux.L.Unlock()
	defer func(){
		s.mux.L.Lock()
		s.reading = false
		s.mux.Broadcast()
	}()
	return s.next()
}

func (s *Scanner)NextFunc(c func(*Token)(bool))(tk *Token, err error){
	s.mux.L.Lock()
	defer s.mux.L.Unlock()
	for {
		if tk = s.nextFromBuf(c); tk != nil {
			return
		}
		if !s.reading {
			break
		}
		s.mux.Wait()
	}
	s.reading = true
	defer func(){
		s.reading = false
		s.mux.Broadcast()
	}()
	for {
		s.mux.L.Unlock()
		tk, err = s.next()
		s.mux.L.Lock()
		if err != nil { return }
		if c(tk) { return }
		if l := len(s.tkbf) - 1; l >= 0 && tk.t == TkBytes && s.tkbf[l].t == TkBytes {
			s.tkbf[l].v = catbytes(s.tkbf[l].v.([]byte), tk.v.([]byte))
		}else{
			s.tkbf = append(s.tkbf, tk)
		}
		s.mux.Broadcast()
	}
}

func (s *Scanner)NextType(t TkType)(*Token, error){
	return s.NextFunc(func(tk *Token)(bool){ return tk.t == t })
}

func catbytes(a []byte, b []byte)([]byte){
	l := len(a)
	n := l + len(b)
	if cap(a) >= n {
		a = a[:n]
		copy(a[l:], b)
		return a
	}
	c := make([]byte, n)
	copy(c[copy(c, a):], b)
	return c
}
