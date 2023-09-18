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
	"golang.org/x/sys/unix"
	"unsafe"
)

type ClientAddress struct {
	Length         *uint32
	LengthPointer  uint64
	Address        *unix.RawSockaddrAny
	AddressPointer uintptr
}

func NewClientAddress() *ClientAddress {
	addressLength := new(uint32)
	*addressLength = unix.SizeofSockaddrAny
	addressLengthPointer := uint64(uintptr(unsafe.Pointer(addressLength)))

	address := new(unix.RawSockaddrAny)
	addressPointer := uintptr(unsafe.Pointer(address))

	return &ClientAddress{
		Length:         addressLength,
		LengthPointer:  addressLengthPointer,
		Address:        address,
		AddressPointer: addressPointer,
	}
}
