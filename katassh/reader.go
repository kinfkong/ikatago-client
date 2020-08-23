package katassh

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
)

const (
	// CompressStarterSymbol the symbol starts the compresssion
	CompressStarterSymbol = 0xff
)

// GTPReader the gtp reader
type GTPReader struct {
	reader    io.Reader
	buffer    *bytes.Buffer
	lastError error
}

// NewGTPReader the new gtp reader
func NewGTPReader(reader io.Reader) *GTPReader {
	return &GTPReader{
		reader:    reader,
		buffer:    bytes.NewBuffer(make([]byte, 0)),
		lastError: nil,
	}
}

func (r *GTPReader) Read(p []byte) (int, error) {
	bufferBytes := r.buffer.Bytes()
	if len(bufferBytes) == 0 {
		if r.lastError != nil {
			return 0, r.lastError
		}
		r.lastError = r.readAndProcess()
		bufferBytes = r.buffer.Bytes()
	}
	// copy the bytes to p buffer
	var n int = 0
	for i := 0; i < len(p) && i < len(bufferBytes); i++ {
		n++
		p[i] = bufferBytes[i]
	}
	if n < len(bufferBytes) {
		r.buffer = bytes.NewBuffer(bufferBytes[n:])
	} else {
		r.buffer = bytes.NewBuffer(make([]byte, 0))
	}
	return n, nil
}

func (r *GTPReader) readAndProcess() error {
	var buf []byte = make([]byte, 4096)
	n, err := r.reader.Read(buf)
	if n == 0 {
		return err
	}
	idx := 0
	s := -1
	e := -1
	for idx < n {
		if buf[idx] == CompressStarterSymbol {
			if s >= 0 && e >= 0 {
				// write the normal buf
				r.buffer.Write(buf[s : e+1])
				s = -1
				e = -1
			}
			n, err := r.processCompress(buf[idx:])
			idx += n
			if err != nil {
				return err
			}
		} else {
			if s < 0 {
				s = idx
			}
			e = idx
			idx++
		}
	}
	if s >= 0 && e >= 0 {
		// write the normal buf
		r.buffer.Write(buf[s : e+1])
		s = -1
		e = -1
	}
	return err
}

func (r *GTPReader) processCompress(p []byte) (int, error) {
	pIndex := 0
	if p[0] != CompressStarterSymbol {
		// safe guard
		return pIndex, errors.New("not_compress_buffer")
	}
	pIndex++
	// read compress length
	lenBuf := make([]byte, 4)
	lenLen := 0
	for i := 0; i < 4 && pIndex < len(p); i++ {
		lenBuf[i] = p[pIndex]
		lenLen++
		pIndex++
	}

	for lenLen < 4 {
		n, err := r.reader.Read(lenBuf[lenLen:])
		lenLen += n
		if err != nil {
			return pIndex, err
		}
	}
	contentLength := int(binary.LittleEndian.Uint32(lenBuf))
	currentContentLength := 0
	contentBuf := make([]byte, contentLength)
	for i := 0; i < contentLength && pIndex < len(p); i++ {
		contentBuf[i] = p[pIndex]
		currentContentLength++
		pIndex++
	}
	for currentContentLength < contentLength {
		n, err := r.reader.Read(contentBuf[currentContentLength:])
		currentContentLength += n
		if err != nil {
			return pIndex, err
		}
	}
	unzipped := toUnzippedBuffer(contentBuf)
	r.buffer.Write(unzipped)
	return pIndex, nil
}

func toUnzippedBuffer(buf []byte) []byte {
	// Write gzipped data to the client
	gr, err := gzip.NewReader(bytes.NewBuffer(buf))
	defer gr.Close()
	data, err := ioutil.ReadAll(gr)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil
	}
	return data
}
