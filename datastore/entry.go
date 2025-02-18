package datastore

import (
	"bufio"
	"encoding/binary"
	"fmt"
)

type entry struct {
	key, value string
	mark       int
}

func (e *entry) Encode() []byte {
	kl := len(e.key)
	vl := len(e.value)
	size := kl + vl + 16
	res := make([]byte, size)
	binary.LittleEndian.PutUint32(res, uint32(size))
	binary.LittleEndian.PutUint32(res[4:], uint32(kl))
	copy(res[8:], e.key)
	binary.LittleEndian.PutUint32(res[kl+8:], uint32(vl))
	copy(res[kl+12:], e.value)
	binary.LittleEndian.PutUint32(res[kl+12+vl:], uint32(e.mark))
	return res
}

func (e *entry) Decode(input []byte) {
	kl := binary.LittleEndian.Uint32(input[4:])
	keyBuf := make([]byte, kl)
	copy(keyBuf, input[8:kl+8])
	e.key = string(keyBuf)

	vl := binary.LittleEndian.Uint32(input[kl+8:])
	valBuf := make([]byte, vl)
	copy(valBuf, input[kl+12:kl+12+vl])
	e.value = string(valBuf)

	e.mark = int(binary.LittleEndian.Uint32(input[kl+12+vl:]))
}

func readValue(in *bufio.Reader) (string, error) {
	header, err := in.Peek(8)
	if err != nil {
		return "", err
	}
	keySize := int(binary.LittleEndian.Uint32(header[4:]))
	_, err = in.Discard(keySize + 8)
	if err != nil {
		return "", err
	}

	header, err = in.Peek(4)
	if err != nil {
		return "", err
	}
	valSize := int(binary.LittleEndian.Uint32(header))
	_, err = in.Discard(4)
	if err != nil {
		return "", err
	}

	data := make([]byte, valSize)
	n, err := in.Read(data)
	if err != nil {
		return "", err
	}
	if n != valSize {
		return "", fmt.Errorf("can't read value bytes (read %d, expected %d)", n, valSize)
	}

	return string(data), nil
}

func readMark(in *bufio.Reader) (int, error) {
	header, err := in.Peek(8)
	if err != nil {
		return -1, err
	}
	keySize := int(binary.LittleEndian.Uint32(header[4:]))
	_, err = in.Discard(keySize + 8)
	if err != nil {
		return -1, err
	}
	header, err = in.Peek(4)
	if err != nil {
		return -1, err
	}
	valSize := int(binary.LittleEndian.Uint32(header))
	_, err = in.Discard(valSize + 4)
	if err != nil {
		return -1, err
	}
	header, err = in.Peek(4)
	if err != nil {
		return -1, err
	}
	mark := int(binary.LittleEndian.Uint32(header))
	return mark, nil
}

func (e *entry) Length() int64 {
	return int64(len(e.key) + len(e.value) + 12)
}
