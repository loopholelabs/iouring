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

package iouring

import (
	"fmt"
	"syscall"
	"unsafe"
)

// Ring is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing.h#L128
type Ring struct {
	SQ          SubmissionQueue
	CQ          CompletionQueue
	Flags       uint32
	FD          int
	Features    uint32
	EnterRingFd int
	IntFlags    uint8
	_pad        [3]uint8
	_pad2       uint32
}

var (
	emptyCQEvent CQEvent
	emptySQEntry SQEntry

	cqEventSize = unsafe.Sizeof(emptyCQEvent)
	sqEntrySize = unsafe.Sizeof(emptySQEntry)
	uint32Size  = unsafe.Sizeof(uint32(0))
)

func NewRing(size uint, params *Params) (*Ring, error) {
	ringFDPTR, _, errno := syscall.Syscall(SYS_IO_URING_SETUP, uintptr(size), uintptr(unsafe.Pointer(params)), 0)
	if errno != 0 {
	}

	ring := &Ring{
		FD: int(ringFDPTR),
		SQ: SubmissionQueue{
			RingSize: uint(uintptr(params.SQOffsets.Array) + uintptr(params.SQEntries)*uint32Size),
		},
		CQ: CompletionQueue{
			RingSize: uint(uintptr(params.CQOffsets.CQEs) + uintptr(params.CQEntries)*cqEventSize),
		},
	}

	err := ring.mmap(params)
	if err != nil {
		return nil, fmt.Errorf("error while MMAPing ring: %w", err)
	}

	for index := uint32(0); index < ring.SQ.RingEntries; index++ {
		*(*uint32)(unsafe.Add(unsafe.Pointer(ring.SQ.Array), index*uint32(unsafe.Sizeof(uint32(0))))) = index
	}

	ring.Features = params.Features
	ring.Flags = params.Flags
	ring.EnterRingFd = ring.FD
	return ring, nil
}

func (r *Ring) mmap(params *Params) error {
	ringPtr, err := mmap(0, uintptr(r.SQ.RingSize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED|syscall.MAP_POPULATE, r.FD, int64(SQRingOffset))
	if err != nil {
		return fmt.Errorf("error while MMAPing SQ Ring: %w", err)
	}
	r.SQ.RingPointer = unsafe.Pointer(ringPtr)

	ringPtr, err = mmap(0, uintptr(r.CQ.RingSize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED|syscall.MAP_POPULATE, r.FD, int64(CQRingOffset))
	if err != nil {
		r.munmap()
		return fmt.Errorf("error while MMAPing CQ Ring: %w", err)
	}
	r.CQ.RingPointer = unsafe.Pointer(ringPtr)

	r.SQ.KHead = (*uint32)(unsafe.Pointer(uintptr(r.SQ.RingPointer) + uintptr(params.SQOffsets.Head)))
	r.SQ.KTail = (*uint32)(unsafe.Pointer(uintptr(r.SQ.RingPointer) + uintptr(params.SQOffsets.Tail)))
	r.SQ._KRingMask = (*uint32)(unsafe.Pointer(uintptr(r.SQ.RingPointer) + uintptr(params.SQOffsets.RingMask)))
	r.SQ._KRingEntries = (*uint32)(unsafe.Pointer(uintptr(r.SQ.RingPointer) + uintptr(params.SQOffsets.RingEntries)))
	r.SQ.KFlags = (*uint32)(unsafe.Pointer(uintptr(r.SQ.RingPointer) + uintptr(params.SQOffsets.Flags)))
	r.SQ.KDropped = (*uint32)(unsafe.Pointer(uintptr(r.SQ.RingPointer) + uintptr(params.SQOffsets.Dropped)))
	r.SQ.Array = (*uint32)(unsafe.Pointer(uintptr(r.SQ.RingPointer) + uintptr(params.SQOffsets.Array)))

	ringPtr, err = mmap(0, sqEntrySize*uintptr(params.SQEntries), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED|syscall.MAP_POPULATE, r.FD, int64(SQEntriesOffset))
	if err != nil {
		r.munmap()
		return fmt.Errorf("error while MMAPing SQ Ring's SQ Entry: %w", err)
	}

	r.SQ.SQEs = (*SQEntry)(unsafe.Pointer(ringPtr))

	r.SQ.RingMask = *r.SQ._KRingMask
	r.SQ.RingEntries = *r.SQ._KRingEntries

	r.CQ.KHead = (*uint32)(unsafe.Pointer(uintptr(r.CQ.RingPointer) + uintptr(params.CQOffsets.Head)))
	r.CQ.KTail = (*uint32)(unsafe.Pointer(uintptr(r.CQ.RingPointer) + uintptr(params.CQOffsets.Tail)))
	r.CQ._KRingMask = (*uint32)(unsafe.Pointer(uintptr(r.CQ.RingPointer) + uintptr(params.CQOffsets.RingMask)))
	r.CQ._KRingEntries = (*uint32)(unsafe.Pointer(uintptr(r.CQ.RingPointer) + uintptr(params.CQOffsets.RingEntries)))
	r.CQ.KOverflow = (*uint32)(unsafe.Pointer(uintptr(r.CQ.RingPointer) + uintptr(params.CQOffsets.Overflow)))
	r.CQ.CQEs = (*CQEvent)(unsafe.Pointer(uintptr(r.CQ.RingPointer) + uintptr(params.CQOffsets.CQEs)))
	if params.CQOffsets.Flags != 0 {
		r.CQ.KFlags = (*uint32)(unsafe.Pointer(uintptr(r.CQ.RingPointer) + uintptr(params.CQOffsets.Flags)))
	}

	r.CQ.RingMask = *r.CQ._KRingMask
	r.CQ.RingEntries = *r.CQ._KRingEntries

	return nil
}

func (r *Ring) munmap() {
	if r.SQ.RingSize > 0 {
		_ = munmap(uintptr(r.SQ.RingPointer), uintptr(r.SQ.RingSize))
	}

	if r.CQ.RingSize > 0 {
		_ = munmap(uintptr(r.CQ.RingPointer), uintptr(r.CQ.RingSize))
	}
}
