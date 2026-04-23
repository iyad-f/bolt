package bolt

import (
	"bufio"
	"io"
	"sync"
)

var (
	readerPool sync.Pool
	writerPool sync.Pool
)

func getReader(r io.Reader) *bufio.Reader {
	reader := readerPool.Get()
	if reader == nil {
		return bufio.NewReader(r)
	}

	br := reader.(*bufio.Reader)
	br.Reset(r)
	return br
}

func putReader(br *bufio.Reader) {
	br.Reset(nil)
	readerPool.Put(br)
}

func getWriter(w io.Writer) *bufio.Writer {
	writer := writerPool.Get()
	if writer == nil {
		return bufio.NewWriter(w)
	}

	bw := writer.(*bufio.Writer)
	bw.Reset(w)
	return bw
}

func putWriter(bw *bufio.Writer) {
	bw.Reset(nil)
	writerPool.Put(bw)
}
