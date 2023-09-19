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

import "syscall"

var (
	_available = false
)

func init() {
	_, _, errno := syscall.RawSyscall(SYS_IO_URING_REGISTER, 0, 1, 0)
	_available = errno == syscall.ENOSYS
}

func IsAvailable() bool {
	return _available
}
