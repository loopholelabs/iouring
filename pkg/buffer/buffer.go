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
	"math"
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

var (
	emptyFD  = ^uintptr(0)
	pageSize = os.Getpagesize()
)

// Buffer is a special buffer that has its memory allocated outside of the Go heap
// via mmap.
type Buffer []byte

func New(size int64, finalizer ...bool) (*Buffer, error) {
	size = int64(math.Ceil(float64(size)/float64(pageSize)) * float64(pageSize))

	if size < 0 {
		return nil, fmt.Errorf("size cannot be negative")
	}

	bufferAddress, err := allocateBuffer(size)
	if err != nil {
		return nil, fmt.Errorf("error while allocating buffer: %w", err)
	}

	buffer := (Buffer)(unsafe.Slice((*byte)(bufferAddress), size))[:0]
	if len(finalizer) > 0 && finalizer[0] {
		runtime.SetFinalizer(&buffer, func(buf *Buffer) {
			err := buf.close()
			if err != nil {
				panic(fmt.Errorf("error while closing buffer: %w", err))
			}
		})
	}
	return &buffer, nil
}

func (buf *Buffer) Write(b []byte) (int, error) {
	if cap(*buf)-len(*buf) < len(b) {
		newSize := int64(math.Ceil(float64(cap(*buf)+len(b))/float64(pageSize)) * float64(pageSize))
		bufferAddress, err := allocateBuffer(newSize)
		if err != nil {
			return 0, fmt.Errorf("error while allocating resized buffer: %w", err)
		}
		buffer := (Buffer)(unsafe.Slice((*byte)(bufferAddress), newSize))[:len(*buf)]
		copy(buffer, *buf)
		err = linked.MUnmap(uintptr(unsafe.Pointer(&(*buf)[0])), uintptr(cap(*buf)))
		if err != nil {
			return 0, fmt.Errorf("error while unmapping existing buffer: %w", err)
		}
		*buf = buffer[:len(buffer)+copy((buffer)[len(buffer):cap(buffer)], b)]
	} else {
		*buf = (*buf)[:len(*buf)+copy((*buf)[len(*buf):cap(*buf)], b)]
	}
	return len(b), nil
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

// Close closes the buffer and frees the memory allocated for it.
//
// Note: This method should only be used if the buffer was allocated with the finalizer
func (buf *Buffer) Close() error {
	return buf.close()
}

func (buf *Buffer) close() error {
	return linked.MUnmap(uintptr(unsafe.Pointer(&((*buf)[:cap(*buf)])[0])), uintptr(cap(*buf)))
}

func allocateBuffer(size int64) (unsafe.Pointer, error) {
	sizePointer := uintptr(size)
	bufferAddress, err := linked.MMap(0, sizePointer, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED|syscall.MAP_ANONYMOUS, int(emptyFD), 0)
	if err != nil {
		return nil, fmt.Errorf("error while mmaping buffer memory space: %w", err)
	}

	fd, err := unix.MemfdCreate("buffer", 0)
	if err != nil {
		return nil, fmt.Errorf("error while creating memfd: %w", err)
	}

	err = unix.Ftruncate(fd, size)
	if err != nil {
		return nil, fmt.Errorf("error while truncating memfd: %w", err)
	}

	_, err = linked.MMap(bufferAddress, sizePointer, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED|syscall.MAP_FIXED, fd, 0)
	if err != nil {
		return nil, fmt.Errorf("error while mmaping buffer: %w", err)
	}

	err = syscall.Close(fd)
	if err != nil {
		return nil, fmt.Errorf("error while closing memfd: %w", err)
	}

	return unsafe.Pointer(bufferAddress), nil
}
