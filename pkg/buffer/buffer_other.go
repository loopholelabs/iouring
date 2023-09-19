//go:build !linux

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
)

var (
	ErrNotAvailable = errors.New("buffer is not available on this platform")
)

type Buffer struct{}

func New(int64) (*Buffer, error) {
	return nil, ErrNotAvailable
}

func (buf *Buffer) Reset() {}

func (buf *Buffer) Write([]byte) bool {
	return false
}

func (buf *Buffer) Bytes() []byte {
	return nil
}

func (buf *Buffer) Len() int {
	return 0
}

func (buf *Buffer) Cap() int {
	return 0
}

func (buf *Buffer) Close() error {
	return ErrNotAvailable
}
