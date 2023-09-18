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

// PrepareRW is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing.h#L378
func (e *SQEntry) PrepareRW(opCode OpCode, fd int, addressPointer uintptr, length uint32, offset uint64) {
	e.OpCode = uint8(opCode)
	e.Flags = 0
	e.IOPriority = 0
	e.FD = int32(fd)
	e.UnionOffset = offset
	e.UnionAddress = uint64(addressPointer)
	e.Length = length
	e.UnionRWFlags = 0
	e.UnionBufferIndexPacked = 0
	e.Personality = 0
	e.UnionSplicedFDIn = 0
	e.UnionAddress3.Address3 = 0
	e.UnionAddress3._Pad2[0] = 0
}

// PrepareAccept is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/include/liburing.h#L591
func (e *SQEntry) PrepareAccept(fd int, addressPointer uintptr, addressLength uint64, flags uint32) {
	e.PrepareRW(OpCodeAccept, fd, addressPointer, 0, addressLength)
	e.UnionRWFlags = flags
}
