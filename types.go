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

	// BigCQE is only required when the ring is initialized with IORING_SETUP_CQE32.
	// Since we don't support IORING_SETUP_CQE32, we don't need to define BigCQE.
	//BigCQE   []uint64
}

// UnionAddress3 is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing/io_uring.h#L88
type UnionAddress3 struct {
	Address3 uint64
	_Pad2    [1]uint64
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
	UnionRWFlags           uint32
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
	_Pad          [2]uint32
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
	_Pad          [2]uint32
}

// OpCode is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing/io_uring.h#L176
type OpCode uint8

const (
	OpCodeNOP OpCode = iota
	OpCodeReadV
	OpCodeWriteV
	OpCodeFsync
	OpCodeReadFixed
	OpCodeWriteFixed
	OpCodePollAdd
	OpCodePollRemove
	OpCodeSyncFileRange
	OpCodeSendMsg
	OpCodeRecvMsg
	OpCodeTimeout
	OpCodeTimeoutRemove
	OpCodeAccept
	OpCodeAsyncCancel
	OpCodeLinkTimeout
	OpCodeConnect
	OpCodeFallocate
	OpCodeOpenat
	OpCodeClose
	OpCodeFilesUpdate
	OpCodeStatx
	OpCodeRead
	OpCodeWrite
	OpCodeFadvise
	OpCodeMadvise
	OpCodeSend
	OpCodeRecv
	OpOpenat2
	OpCodeEpollCtl
	OpCodeSplice
	OpCodeProvideBuffers
	OpCodeRemoveBuffers
	OpCodeTee
	OpCodeShutdown
	OpCodeRenameat
	OpCodeUnlinkat
	OpCodeMkdirat
	OpCodeSymlinkat
	OpCodeLinkat
	OpCodeMsgRing
	OpCodeFsetxattr
	OpCodeSetxattr
	OpCodeFgetxattr
	OpCodeGetxattr
	OpCodeSocket
	OpCodeUringCmd
	OpCodeSendZC
	OpCodeSendMsgZC

	OpCodeLast
)

// Setup is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing/io_uring.h#L140
type Setup uint32

const (
	SetupIOPoll Setup = 1 << iota
	SetupSQPoll
	SetupSQAff
	SetupCQSize
	SetupClamp
	SetupAttachWQ
	SetupRDisabled
	SetupSubmitAll
	SetupCoopTaskRun
	SetupTaskRunFlag
	SetupSQE128
	SetupCQE32
	SetupSingleIssuer
	SetupDeferTaskRun
	SetupNoMMap
	SetupRegisteredFDOnly
)

// SQStatus is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing/io_uring.h#L415
type SQStatus uint32

const (
	SQStatusNeedWakeup SQStatus = 1 << iota
	SQStatusCQOverflow
	SQStatusTaskRun
)

// Enter is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing/io_uring.h#L441
type Enter uint32

const (
	EnterGetEvents Enter = 1 << iota
	EnterSQWakeup
	EnterSQWait
	EnterExtArg
	EnterRegisteredRing
)

// IntFlag is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/int_flags.h#L5
type IntFlag uint8

const (
	IntFlagRegRing    IntFlag = 1
	IntFlagRegRegRing IntFlag = 2
)

// Feature is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing/io_uring.h#L466
type Feature uint32

const (
	FeatureSingleMMap Feature = 1 << iota
	FeatureNoDrop
	FeatureSubmitStable
	FeatureRWCurPos
	FeatureCurPersonality
	FeatureFastPoll
	FeaturePoll32Bits
	FeatureSQPollNonfixed
	FeatureExtArg
	FeatureNativeWorkers
	FeatureRcrcTags
	FeatureCQESkip
	FeatureLinkedFile
	FeatureRegRegRing
)

const (
	// _NSIG is defined here: https://github.com/torvalds/linux/blob/v6.5/include/uapi/asm-generic/signal.h#L7
	_NSIG = 64
)
