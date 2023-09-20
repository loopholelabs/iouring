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
	"syscall"
	"unsafe"
)

// RegisterBuffers is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/register.c#L59
func (r *Ring) RegisterBuffers(iovecs []syscall.Iovec, NRIOVecs uint32) (uint, error) {
	return r.DoRegister(RegisterOpCodeRegisterBuffers, unsafe.Pointer(&iovecs[0]), uint32(len(iovecs)))
}

// DoRegister is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/register.c#L11
func (r *Ring) DoRegister(opCode RegisterOpCode, arg unsafe.Pointer, NRArgs uint32) (uint, error) {
	if r.IntFlags&uint8(IntFlagRegRegRing) != 0 {
		opCode |= RegisterOpCodeRegisterUseRegisteredRing
	}

	return r._Register(uint32(opCode), arg, NRArgs)
}
