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

// MMap is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/setup.c#L18
func MMap(fd int, params *Params, sq *SubmissionQueue, cq *CompletionQueue) error {
	cqEventSize := cqEventSize
	if params.Flags&uint32(SetupCQE32) != 0 {
		cqEventSize += cqEventSize
	}

	sq.RingSize = uint(uintptr(params.SQOffsets.Array) + uintptr(params.SQEntries)*uint32Size)
	cq.RingSize = uint(uintptr(params.CQOffsets.CQEs) + uintptr(params.CQEntries)*cqEventSize)

	if params.Features&uint32(FeatureSingleMMap) != 0 {
		if cq.RingSize > sq.RingSize {
			sq.RingSize = cq.RingSize
		}
		cq.RingSize = sq.RingSize
	}

	ringPtr, err := mmap(0, uintptr(sq.RingSize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED|syscall.MAP_POPULATE, fd, int64(SQRingOffset))
	if err != nil {
		return fmt.Errorf("error while MMAPing SQ Ring: %w", err)
	}
	sq.RingPointer = unsafe.Pointer(ringPtr)

	ringPtr, err = mmap(0, uintptr(cq.RingSize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED|syscall.MAP_POPULATE, fd, int64(CQRingOffset))
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

	ringPtr, err = mmap(0, sqEntrySize*uintptr(params.SQEntries), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED|syscall.MAP_POPULATE, fd, int64(SQEntriesOffset))
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
		_ = munmap(uintptr(sq.RingPointer), uintptr(sq.RingSize))
	}

	if cq.RingSize > 0 {
		_ = munmap(uintptr(cq.RingPointer), uintptr(cq.RingSize))
	}
}
