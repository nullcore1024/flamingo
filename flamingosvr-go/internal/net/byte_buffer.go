package net

import (
	"bytes"
	"encoding/binary"
)

type ByteBuffer struct {
	buf *bytes.Buffer
}

func NewByteBuffer() *ByteBuffer {
	return &ByteBuffer{
		buf: &bytes.Buffer{},
	}
}

func NewByteBufferWithSize(size int) *ByteBuffer {
	buf := make([]byte, 0, size)
	return &ByteBuffer{
		buf: bytes.NewBuffer(buf),
	}
}

func (bb *ByteBuffer) Read(p []byte) (n int, err error) {
	return bb.buf.Read(p)
}

func (bb *ByteBuffer) Write(p []byte) (n int, err error) {
	return bb.buf.Write(p)
}

func (bb *ByteBuffer) WriteByte(c byte) error {
	return bb.buf.WriteByte(c)
}

func (bb *ByteBuffer) WriteString(s string) (n int, err error) {
	return bb.buf.WriteString(s)
}

func (bb *ByteBuffer) ReadByte() (byte, error) {
	return bb.buf.ReadByte()
}

func (bb *ByteBuffer) ReadUint16() (uint16, error) {
	var v uint16
	err := binary.Read(bb.buf, binary.BigEndian, &v)
	return v, err
}

func (bb *ByteBuffer) ReadUint32() (uint32, error) {
	var v uint32
	err := binary.Read(bb.buf, binary.BigEndian, &v)
	return v, err
}

func (bb *ByteBuffer) ReadUint64() (uint64, error) {
	var v uint64
	err := binary.Read(bb.buf, binary.BigEndian, &v)
	return v, err
}

func (bb *ByteBuffer) WriteUint16(v uint16) error {
	return binary.Write(bb.buf, binary.BigEndian, v)
}

func (bb *ByteBuffer) WriteUint32(v uint32) error {
	return binary.Write(bb.buf, binary.BigEndian, v)
}

func (bb *ByteBuffer) WriteUint64(v uint64) error {
	return binary.Write(bb.buf, binary.BigEndian, v)
}

func (bb *ByteBuffer) Len() int {
	return bb.buf.Len()
}

func (bb *ByteBuffer) Cap() int {
	return bb.buf.Cap()
}

func (bb *ByteBuffer) Reset() {
	bb.buf.Reset()
}

func (bb *ByteBuffer) Bytes() []byte {
	return bb.buf.Bytes()
}

func (bb *ByteBuffer) String() string {
	return bb.buf.String()
}
