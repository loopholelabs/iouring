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

const (
	SQRingOffset    uint64 = 0
	CQRingOffset    uint64 = 0x8000000
	SQEntriesOffset uint64 = 0x10000000
	PBUFRingOffset  uint64 = 0x80000000
	PBUFShiftOffset uint64 = 16
	MMAPMaskOffset  uint64 = 0xf8000000
)
