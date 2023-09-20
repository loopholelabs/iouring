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
	"errors"
	"fmt"
	"github.com/loopholelabs/iouring/pkg/linked"
	"math"
	"unsafe"
)

var (
	ErrTooLarge = errors.New("invalid data size")
)

// Fixed is a type of Buffer that has a constant, fixed size and is not dynamically resizable.
type Fixed []byte

func NewFixed(size int64) (*Fixed, error) {
	size = int64(math.Ceil(float64(size)/float64(pageSize)) * float64(pageSize))

	if size < 0 {
		return nil, fmt.Errorf("size cannot be negative")
	}

	bufferAddress, err := allocateBuffer(size)
	if err != nil {
		return nil, fmt.Errorf("error while allocating buffer: %w", err)
	}

	buffer := (Fixed)(unsafe.Slice((*byte)(bufferAddress), size))[:0]
	return &buffer, nil
}

func (buf *Fixed) Write(b []byte) (int, error) {
	if cap(*buf)-len(*buf) < len(b) {
		return 0, ErrTooLarge
	} else {
		*buf = (*buf)[:len(*buf)+copy((*buf)[len(*buf):cap(*buf)], b)]
	}
	return len(b), nil
}

func (buf *Fixed) Reset() {
	*buf = (*buf)[:0]
}

func (buf *Fixed) Bytes() []byte {
	return *buf
}

func (buf *Fixed) Len() int {
	return len(*buf)
}

func (buf *Fixed) Cap() int {
	return cap(*buf)
}

func (buf *Fixed) Close() error {
	return linked.MUnmap(uintptr(unsafe.Pointer(&((*buf)[:cap(*buf)])[0])), uintptr(cap(*buf)))
}
