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
	"github.com/loopholelabs/iouring/pkg/linked"
	"syscall"
	"unsafe"
)

// QueueMMap is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/setup.c#L96
func (r *Ring) QueueMMap(fd int, params *Params) error {
	err := MMap(fd, params, &r.SQ, &r.CQ)
	if err != nil {
		return fmt.Errorf("error while MMAPing ring: %w", err)
	}
	r.Flags = params.Flags
	r.FD = fd
	r.EnterRingFd = fd
	r.IntFlags = 0

	return nil
}

// QueueInitParams is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/setup.c#L148
func (r *Ring) QueueInitParams(entries uint32, params *Params) error {
	ringFDPTR, _, errno := syscall.Syscall(SYS_IO_URING_SETUP, uintptr(entries), uintptr(unsafe.Pointer(params)), 0)
	if errno != 0 {
		return fmt.Errorf("error while creating ring: %w", errno)
	}

	ringFD := int(ringFDPTR)
	err := r.QueueMMap(ringFD, params)
	if err != nil {
		_ = syscall.Close(ringFD)
		return fmt.Errorf("error while MMAPing ring: %w", err)
	}

	for index := uint32(0); index < r.SQ.RingEntries; index++ {
		*(*uint32)(unsafe.Add(unsafe.Pointer(r.SQ.Array), index*uint32(unsafe.Sizeof(uint32(0))))) = index
	}
	r.Features = params.Features

	return nil
}

// QueueInit is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/setup.c#L181
func (r *Ring) QueueInit(entries uint32, flags uint32) error {
	var params Params
	params.Flags = flags
	return r.QueueInitParams(entries, &params)
}

// MMap is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/setup.c#L18
func MMap(fd int, params *Params, sq *SubmissionQueue, cq *CompletionQueue) error {
	sq.RingSize = uint(uintptr(params.SQOffsets.Array) + uintptr(params.SQEntries)*uint32Size)
	cq.RingSize = uint(uintptr(params.CQOffsets.CQEs) + uintptr(params.CQEntries)*cqEventSize)

	if params.Features&uint32(FeatureSingleMMap) != 0 {
		if cq.RingSize > sq.RingSize {
			sq.RingSize = cq.RingSize
		}
		cq.RingSize = sq.RingSize
	}

	ringPtr, err := linked.MMap(0, uintptr(sq.RingSize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED|syscall.MAP_POPULATE, fd, int64(SQRingOffset))
	if err != nil {
		return fmt.Errorf("error while MMAPing SQ Ring: %w", err)
	}
	sq.RingPointer = unsafe.Pointer(ringPtr)

	ringPtr, err = linked.MMap(0, uintptr(cq.RingSize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED|syscall.MAP_POPULATE, fd, int64(CQRingOffset))
	if err != nil {
		MUnmap(sq, cq)
		return fmt.Errorf("error while MMAPing CQ Ring: %w", err)
	}
	cq.RingPointer = unsafe.Pointer(ringPtr)

	sq.KHead = (*uint32)(unsafe.Pointer(uintptr(sq.RingPointer) + uintptr(params.SQOffsets.Head)))
	sq.KTail = (*uint32)(unsafe.Pointer(uintptr(sq.RingPointer) + uintptr(params.SQOffsets.Tail)))
	sq._KRingMask = (*uint32)(unsafe.Pointer(uintptr(sq.RingPointer) + uintptr(params.SQOffsets.RingMask)))
	sq._KRingEntries = (*uint32)(unsafe.Pointer(uintptr(sq.RingPointer) + uintptr(params.SQOffsets.RingEntries)))
	sq.KFlags = (*uint32)(unsafe.Pointer(uintptr(sq.RingPointer) + uintptr(params.SQOffsets.Flags)))
	sq.KDropped = (*uint32)(unsafe.Pointer(uintptr(sq.RingPointer) + uintptr(params.SQOffsets.Dropped)))
	sq.Array = (*uint32)(unsafe.Pointer(uintptr(sq.RingPointer) + uintptr(params.SQOffsets.Array)))

	ringPtr, err = linked.MMap(0, sqEntrySize*uintptr(params.SQEntries), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED|syscall.MAP_POPULATE, fd, int64(SQEntriesOffset))
	if err != nil {
		MUnmap(sq, cq)
		return fmt.Errorf("error while MMAPing SQ Ring's SQ Entry: %w", err)
	}

	sq.SQEs = (*SQEntry)(unsafe.Pointer(ringPtr))

	sq.RingMask = *sq._KRingMask
	sq.RingEntries = *sq._KRingEntries

	cq.KHead = (*uint32)(unsafe.Pointer(uintptr(cq.RingPointer) + uintptr(params.CQOffsets.Head)))
	cq.KTail = (*uint32)(unsafe.Pointer(uintptr(cq.RingPointer) + uintptr(params.CQOffsets.Tail)))
	cq._KRingMask = (*uint32)(unsafe.Pointer(uintptr(cq.RingPointer) + uintptr(params.CQOffsets.RingMask)))
	cq._KRingEntries = (*uint32)(unsafe.Pointer(uintptr(cq.RingPointer) + uintptr(params.CQOffsets.RingEntries)))
	cq.KOverflow = (*uint32)(unsafe.Pointer(uintptr(cq.RingPointer) + uintptr(params.CQOffsets.Overflow)))
	cq.CQEs = (*CQEvent)(unsafe.Pointer(uintptr(cq.RingPointer) + uintptr(params.CQOffsets.CQEs)))
	if params.CQOffsets.Flags != 0 {
		cq.KFlags = (*uint32)(unsafe.Pointer(uintptr(cq.RingPointer) + uintptr(params.CQOffsets.Flags)))
	}

	cq.RingMask = *cq._KRingMask
	cq.RingEntries = *cq._KRingEntries

	return nil
}

// MUnmap is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/setup.c#L11
func MUnmap(sq *SubmissionQueue, cq *CompletionQueue) {
	if sq.RingSize > 0 {
		_ = linked.MUnmap(uintptr(sq.RingPointer), uintptr(sq.RingSize))
	}

	if cq.RingSize > 0 {
		_ = linked.MUnmap(uintptr(cq.RingPointer), uintptr(cq.RingSize))
	}
}
