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

package iouring

import (
	"unsafe"
)

type Sigset_t struct {
	Val [16]uint64
}

type Ring struct{}

func NewRing() (*Ring, error) {
	return nil, ErrNotAvailable
}

func (r *Ring) GetSQEntry() *SQEntry {
	return nil
}

func (r *Ring) WaitCQEvent() (*CQEvent, error) {
	return nil, ErrNotAvailable
}

func (r *Ring) WaitCQEventNR(uint32) (*CQEvent, error) {
	return nil, ErrNotAvailable
}

func (r *Ring) _PeekCQEvent(*uint32) (*CQEvent, error) {
	return nil, ErrNotAvailable
}

func (r *Ring) PeekCQEvent() (*CQEvent, error) {
	return nil, ErrNotAvailable
}

func (r *Ring) GetCQEvent(uint32, uint32, *Sigset_t) (*CQEvent, error) {
	return nil, ErrNotAvailable
}

func (r *Ring) _GetCQEvent(*GetData) (*CQEvent, error) {
	return nil, ErrNotAvailable
}

func (r *Ring) CQESeen(*CQEvent) {}

func (r *Ring) CQAdvance(uint32) {}

func (r *Ring) CQNeedsFlush() bool {
	return false
}

func (r *Ring) CQNeedsEnter() bool {
	return false
}

func (r *Ring) SQNeedsEnter(uint32, *uint32) bool {
	return false
}

func (r *Ring) FlushSQ() uint32 {
	return 0
}

func (r *Ring) _Submit(uint32, uint32, bool) (uint, error) {
	return 0, ErrNotAvailable
}

func (r *Ring) Submit() (uint, error) {
	return 0, ErrNotAvailable
}

func (r *Ring) Enter(uint32, uint32, uint32, unsafe.Pointer) (uint, error) {
	return 0, ErrNotAvailable
}

func (r *Ring) Enter2(uint32, uint32, uint32, unsafe.Pointer, int) (uint, error) {
	return 0, ErrNotAvailable
}

func (r *Ring) QueueMMap(int, *Params) error {
	return ErrNotAvailable
}

func (r *Ring) QueueInitParams(uint32, *Params) error {
	return ErrNotAvailable
}

func (r *Ring) QueueInit(uint32, uint32) error {
	return ErrNotAvailable
}

func (r *Ring) Close() error {
	return ErrNotAvailable
}
