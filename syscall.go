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

// Enter is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/arch/generic/syscall.h#L35
func (r *Ring) Enter(submitted uint32, waitNR uint32, flags uint32, sig unsafe.Pointer) (uint, error) {
	return r.Enter2(submitted, waitNR, flags, sig, _NSIG/8)
}

// Enter2 is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/arch/generic/syscall.h#L24
func (r *Ring) Enter2(submitted uint32, waitNR uint32, flags uint32, sig unsafe.Pointer, size int) (uint, error) {
	res, _, err := syscall.Syscall6(
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

	return uint(res), nil
}

// _Register is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/arch/syscall-defs.h#L64
func (r *Ring) _Register(opCode uint32, arg unsafe.Pointer, NRArgs uint32) (uint, error) {
	res, _, err := syscall.Syscall6(
		SYS_IO_URING_REGISTER,
		uintptr(r.EnterRingFd),
		uintptr(opCode),
		uintptr(arg),
		uintptr(NRArgs),
		0,
		0,
	)

	if err > 0 {
		return 0, err
	}

	return uint(res), nil
}
