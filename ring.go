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
	"sync/atomic"
	"syscall"
	"unsafe"
)

var (
	emptyCQEvent CQEvent
	emptySQEntry SQEntry

	cqEventSize = unsafe.Sizeof(emptyCQEvent)
	sqEntrySize = unsafe.Sizeof(emptySQEntry)
	uint32Size  = unsafe.Sizeof(uint32(0))
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
	_Pad        [3]uint8
	_Pad2       uint32
}

func NewRing() (*Ring, error) {
	return new(Ring), nil
}

// GetSQEntry is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing.h#L1320
func (r *Ring) GetSQEntry() *SQEntry {
	head := atomic.LoadUint32(r.SQ.KHead)
	next := r.SQ.SQETail + 1
	if next-head <= r.SQ.RingEntries {
		sqe := (*SQEntry)(unsafe.Add(unsafe.Pointer(r.SQ.SQEs), uintptr(r.SQ.SQETail&r.SQ.RingMask)*sqEntrySize))
		r.SQ.SQETail = next
		return sqe
	}
	return nil
}

// WaitCQEvent is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing.h#L1304
func (r *Ring) WaitCQEvent() (*CQEvent, error) {
	cqe, err := r._PeekCQEvent(nil)
	if err == nil && cqe != nil {
		return cqe, nil
	}

	return r.WaitCQEventNR(1)
}

// WaitCQEventNR is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing.h#L1233
func (r *Ring) WaitCQEventNR(waitNR uint32) (*CQEvent, error) {
	return r.GetCQEvent(0, waitNR, nil)
}

// _PeekCQEvent is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing.h#L1245
func (r *Ring) _PeekCQEvent(nrAvailable *uint32) (cqe *CQEvent, err error) {
	mask := r.CQ.RingMask
	available := uint32(0)
	for {
		tail := atomic.LoadUint32(r.CQ.KTail)
		head := *r.CQ.KHead

		cqe = nil
		available = tail - head
		if available == 0 {
			break
		}

		cqe = (*CQEvent)(
			unsafe.Add(unsafe.Pointer(r.CQ.CQEs), uintptr(head&mask)*cqEventSize),
		)

		if r.Features&uint32(FeatureExtArg) == 0 && cqe.UserData == LIBURING_UDATA_TIMEOUT {
			if cqe.Res < 0 {
				err = syscall.Errno(uintptr(-cqe.Res))
			}
			r.CQAdvance(1)
			if err == nil {
				continue
			}
			cqe = nil
		}

		break
	}

	if nrAvailable != nil {
		*nrAvailable = available
	}

	return
}

// PeekCQEvent is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing.h#L1291
func (r *Ring) PeekCQEvent() (cqe *CQEvent, err error) {
	cqe, err = r._PeekCQEvent(nil)
	if err == nil && cqe != nil {
		return cqe, nil
	}

	return r.WaitCQEventNR(0)
}

// CQESeen is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing.h#L319
func (r *Ring) CQESeen(cqe *CQEvent) {
	if cqe != nil {
		r.CQAdvance(1)
	}
}

// CQAdvance is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing.h#L302
func (r *Ring) CQAdvance(numCQEs uint32) {
	if numCQEs > 0 {
		atomic.StoreUint32(r.CQ.KHead, *r.CQ.KHead+numCQEs)
	}
}

func (r *Ring) Close() error {
	MUnmap(&r.SQ, &r.CQ)
	return syscall.Close(r.FD)
}
