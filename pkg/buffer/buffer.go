//go:build linux

/*
	Copyright 2023 Loophole Labs

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

		   http://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package buffer

import (
	"fmt"
	"github.com/loopholelabs/iouring/pkg/linked"
	"golang.org/x/sys/unix"
	"syscall"
	"unsafe"
)

const (
	emptyFD = int(^uintptr(0))
)

// Buffer is a special buffer that has its memory allocated outside of the Go heap
// via mmap.
type Buffer []byte

func New(size int64) (*Buffer, error) {
	if size < 0 {
		return nil, fmt.Errorf("size cannot be negative")
	}

	bufferAddress, err := allocateBuffer(size)
	if err != nil {
		return nil, fmt.Errorf("error while allocating buffer: %w", err)
	}

	buffer := (Buffer)(unsafe.Slice((*byte)(unsafe.Pointer(bufferAddress)), size))
	return &buffer, nil
}

func (buf *Buffer) Write(b []byte) (int, error) {
	if cap(*buf)-len(*buf) < len(b) {
		newSize := int64(cap(*buf) + len(b))
		bufferAddress, err := allocateBuffer(newSize)
		if err != nil {
			return 0, fmt.Errorf("error while allocating resized buffer: %w", err)
		}
		buffer := (Buffer)(unsafe.Slice((*byte)(unsafe.Pointer(bufferAddress)), newSize))[:len(*buf)]
		copy(buffer, *buf)

		err = linked.MUnmap(uintptr(unsafe.Pointer(&(*buf)[0])), uintptr(cap(*buf)))
		if err != nil {
			return 0, fmt.Errorf("error while unmapping existing buffer: %w", err)
		}

		*buf = append(buffer, b...)

		//*buf = append((*buf)[:len(*buf)], b...)
	} else {
		*buf = (*buf)[:len(*buf)+copy((*buf)[len(*buf):cap(*buf)], b)]
	}
	return len(*buf), nil
}

func (buf *Buffer) Reset() {
	*buf = (*buf)[:0]
}

func (buf *Buffer) Bytes() []byte {
	return *buf
}

func (buf *Buffer) Len() int {
	return len(*buf)
}

func (buf *Buffer) Cap() int {
	return cap(*buf)
}

func (buf *Buffer) Close() error {
	return linked.MUnmap(uintptr(unsafe.Pointer(&(*buf)[0])), uintptr(cap(*buf)))
}

func allocateBuffer(size int64) (uintptr, error) {
	sizePointer := uintptr(size)
	bufferAddress, err := linked.MMap(0, sizePointer, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED|syscall.MAP_ANONYMOUS, emptyFD, 0)
	if err != nil {
		return 0, fmt.Errorf("error while mmaping buffer memory space: %w", err)
	}

	fd, err := unix.MemfdCreate("buffer", 0)
	if err != nil {
		return 0, fmt.Errorf("error while creating memfd: %w", err)
	}

	err = unix.Ftruncate(fd, size)
	if err != nil {
		return 0, fmt.Errorf("error while truncating memfd: %w", err)
	}

	_, err = linked.MMap(bufferAddress, sizePointer, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED|syscall.MAP_FIXED, fd, 0)
	if err != nil {
		return 0, fmt.Errorf("error while mmaping buffer: %w", err)
	}

	err = syscall.Close(fd)
	if err != nil {
		return 0, fmt.Errorf("error while closing memfd: %w", err)
	}

	return bufferAddress, nil
}
