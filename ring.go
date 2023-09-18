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
	"sync/atomic"
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
	_Pad        [3]uint8
	_Pad2       uint32
}

var (
	emptyCQEvent CQEvent
	emptySQEntry SQEntry

	cqEventSize = unsafe.Sizeof(emptyCQEvent)
	sqEntrySize = unsafe.Sizeof(emptySQEntry)
	uint32Size  = unsafe.Sizeof(uint32(0))
)

func NewRing() *Ring {
	return new(Ring)
}

// GetSQEntry is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing.h#L1320
//
// TODO: Handle IORING_SETUP_SQPOLL (https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing.h#L1329)
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

// CQRingNeedsFlush is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/queue.c#L42
func (r *Ring) CQRingNeedsFlush() bool {
	return atomic.LoadUint32(r.SQ.KFlags)&uint32(SQStatusCQOverflow|SQStatusTaskRun) != 0
}

// CQNeedsEnter is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/queue.c#L48
func (r *Ring) CQNeedsEnter() bool {
	return (r.Flags&uint32(SetupIOPoll)) != 0 || r.CQRingNeedsFlush()
}

// SQNeedsEnter is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/queue.c#L17
func (r *Ring) SQNeedsEnter(submit uint32, flags *uint32) bool {
	if submit == 0 {
		return false
	}
	if (r.Flags & uint32(SetupSQPoll)) == 0 {
		return true
	}

	if atomic.LoadUint32(r.SQ.KFlags)&uint32(SQStatusNeedWakeup) != 0 {
		*flags |= uint32(EnterSQWakeup)
		return true
	}

	return false
}

// FlushSQ is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/queue.c#L204
func (r *Ring) FlushSQ() uint32 {
	tail := r.SQ.SQETail
	if r.SQ.SQEHead != tail {
		r.SQ.SQEHead = tail
		atomic.StoreUint32(r.SQ.KTail, tail)
	}

	// There is a potential race condition here, left
	// intentionally because it will not cause any issues
	// https://github.com/axboe/liburing/blob/liburing-2.4/src/queue.c#L219
	return tail - *r.SQ.KHead
}

// _Submit is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/queue.c#L368
func (r *Ring) _Submit(submitted uint32, waitNR uint32, getEvents bool) (ret uint, err error) {
	cqNeedsEnter := getEvents || waitNR != 0 || r.CQNeedsEnter()

	flags := uint32(0)
	if r.SQNeedsEnter(submitted, &flags) || cqNeedsEnter {
		if cqNeedsEnter {
			flags |= uint32(EnterGetEvents)
		}

		if r.IntFlags&uint8(IntFlagRegRing) != 0 {
			flags |= uint32(EnterRegisteredRing)
		}

		ret, err = r.Enter(submitted, waitNR, flags, nil)
	} else {
		ret = uint(submitted)
	}
	return
}

func (r *Ring) Submit() (uint, error) {
	return r._Submit(r.FlushSQ(), 0, false)
}

// Enter is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/arch/generic/syscall.h#L35
func (r *Ring) Enter(submitted uint32, waitNR uint32, flags uint32, sig unsafe.Pointer) (uint, error) {
	return r.Enter2(submitted, waitNR, flags, sig, _NSIG/8)
}

// Enter2 is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/arch/generic/syscall.h#L24
func (r *Ring) Enter2(submitted uint32, waitNR uint32, flags uint32, sig unsafe.Pointer, size int) (uint, error) {
	consumed, _, err := syscall.Syscall6(
		SYS_IO_URING_ENTER,
		uintptr(r.EnterRingFd),
		uintptr(submitted),
		uintptr(waitNR),
		uintptr(flags),
		uintptr(sig),
		uintptr(size),
	)

	if err > 0 {
		return 0, err
	}

	return uint(consumed), nil
}

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

func (r *Ring) Close() error {
	MUnmap(&r.SQ, &r.CQ)
	return syscall.Close(r.FD)
}
