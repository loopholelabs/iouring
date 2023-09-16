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

import "unsafe"

// SQRingOffsets is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing/io_uring.h#L400
type SQRingOffsets struct {
	Head        uint32
	Tail        uint32
	RingMask    uint32
	RingEntries uint32
	Flags       uint32
	Dropped     uint32
	Array       uint32
	ResV1       uint32
	ResV2       uint64
}

// CQRingOffsets is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing/io_uring.h#L419
type CQRingOffsets struct {
	Head        uint32
	Tail        uint32
	RingMask    uint32
	RingEntries uint32
	Overflow    uint32
	CQEs        uint32
	Flags       uint32
	ResV1       uint32
	ResV2       uint64
}

// Params is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing/io_uring.h#L450
type Params struct {
	SQEntries    uint32
	CQEntries    uint32
	Flags        uint32
	SQThreadCPU  uint32
	SQThreadIdle uint32
	Features     uint32
	WQFD         uint32
	ResV         [3]uint32
	SQOffsets    SQRingOffsets
	CQOffsets    CQRingOffsets
}

// CQEvent is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing/io_uring.h#L357
type CQEvent struct {
	UserData uint64
	Res      int32
	Flags    uint32
	BigCQE   []uint64
}

// UnionAddress3 is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing/io_uring.h#L88
type UnionAddress3 struct {
	Address3 uint64
	_pad2    [1]uint64
}

// SQEntry is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing/io_uring.h#L30
type SQEntry struct {
	OpCode                 uint8
	Flags                  uint8
	IOPriority             uint16
	FD                     int32
	UnionOffset            uint64
	UnionAddress           uint64
	Length                 uint32
	UnionOpCodeFlags       uint32
	UserData               uint64
	UnionBufferIndexPacked uint16
	Personality            uint16
	UnionSplicedFDIn       int32
	UnionAddress3          UnionAddress3
}

// SubmissionQueue is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing.h#L84
type SubmissionQueue struct {
	KHead         *uint32
	KTail         *uint32
	_KRingMask    *uint32 // Deprecated: use `RingMask` instead of `*_KRingMask`
	_KRingEntries *uint32 // Deprecated: use `ring_entries` instead of `*_KRingEntries`
	KFlags        *uint32
	KDropped      *uint32
	Array         *uint32
	SQEs          *SQEntry
	SQEHead       uint32
	SQETail       uint32
	RingSize      uint
	RingPointer   unsafe.Pointer
	RingMask      uint32
	RingEntries   uint32
	_pad          [2]uint32
}

// CompletionQueue is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing.h#L108
type CompletionQueue struct {
	KHead         *uint32
	KTail         *uint32
	_KRingMask    *uint32 // Deprecated: use `RingMask` instead of `*_KRingMask`
	_KRingEntries *uint32 // Deprecated: use `ring_entries` instead of `*_KRingEntries`
	KFlags        *uint32
	KOverflow     *uint32
	CQEs          *CQEvent
	RingSize      uint
	RingPointer   unsafe.Pointer
	RingMask      uint32
	RingEntries   uint32
	_pad          [2]uint32
}
