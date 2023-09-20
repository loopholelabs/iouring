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

package buffer

import (
	"crypto/rand"
	"testing"
)

func BenchmarkFixedAllocationsNoResize(b *testing.B) {
	randomBytes := make([]byte, 512)
	_, err := rand.Read(randomBytes)
	if err != nil {
		b.Fatalf("failed to read random bytes: %v", err)
	}

	buf, err := NewFixed(512)
	if err != nil {
		b.Fatalf("failed to create buffer: %v", err)
	}

	b.Cleanup(func() {
		err = buf.Close()
		if err != nil {
			b.Fatalf("failed to close buffer: %v", err)
		}
	})

	var num int

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		num, err = buf.Write(randomBytes)
		if err != nil {
			b.Fatalf("failed to write bytes: %v", err)
		}
		if num != 512 {
			b.Fatalf("number of bytes written is not correct: %d", num)
		}
		buf.Reset()
	}
}

func BenchmarkFixedAllocationsNoResizePool(b *testing.B) {
	randomBytes := make([]byte, 512)
	_, err := rand.Read(randomBytes)
	if err != nil {
		b.Fatalf("failed to read random bytes: %v", err)
	}

	var num int
	var buf *Fixed

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf, err = GetFixed()
		if err != nil {
			b.Fatalf("failed to write bytes: %v", err)
		}
		num, err = buf.Write(randomBytes)
		if err != nil {
			b.Fatalf("failed to write bytes: %v", err)
		}
		if num != len(randomBytes) {
			b.Fatalf("number of bytes written is not correct: %d", num)
		}
		PutFixed(buf)
	}
}
