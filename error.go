
package unixcsl

import (
	fmt "fmt"
)

type UnexpectByteError struct{
	char byte
	expects []byte
	extra string
}

func NewUnexpectByteError(char byte, expects ...byte)(*UnexpectByteError){
	return &UnexpectByteError{
		char: char,
		expects: expects,
		extra: "",
	}
}

func (e *UnexpectByteError)setExtra(extra string)(*UnexpectByteError){
	e.extra = extra
	return e
}

func (e *UnexpectByteError)Error()(str string){
	str = fmt.Sprintf("Unexpected token: (%x)", e.char)
	if len(e.expects) > 0 {
		str += " expect ["
		for _, c := range e.expects {
			str += fmt.Sprintf("(%x), ", c)
		}
		str = str[:len(str) - 2] + "]"
	}
	str += e.extra
	return
}
