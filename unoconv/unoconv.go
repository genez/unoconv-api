package unoconv

import (
	"io"
)

type Request struct {
	Filename string
	Filetype string
	W        io.Writer
	ErrChan  chan error
}

type UnoConv struct {
	RequestChan chan Request
}

func (u *UnoConv) Convert(filename, filetype string, w io.Writer) error {
	err := make(chan error)
	req := Request{
		filename,
		filetype,
		w,
		err,
	}

	u.RequestChan <- req
	return <-err
}
