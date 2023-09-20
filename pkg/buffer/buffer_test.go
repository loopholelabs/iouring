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
	"github.com/loopholelabs/polyglot"
	"github.com/stretchr/testify/require"
	"testing"
)

func BenchmarkBufferAllocationsNoResize(b *testing.B) {
	randomBytes := make([]byte, 512)
	_, err := rand.Read(randomBytes)
	if err != nil {
		b.Fatalf("failed to read random bytes: %v", err)
	}

	buf, err := New(512)
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

func BenchmarkPolyglotAllocationsNoResize(b *testing.B) {
	randomBytes := make([]byte, 512)
	_, err := rand.Read(randomBytes)
	if err != nil {
		b.Fatalf("failed to read random bytes: %v", err)
	}

	buf := polyglot.GetBuffer()
	b.Cleanup(func() {
		polyglot.PutBuffer(buf)
	})

	var num int

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		num = buf.Write(randomBytes)
		if num != 512 {
			b.Fatalf("number of bytes written is not correct: %d", num)
		}
		buf.Reset()
	}
}

func BenchmarkBufferSizeCheck(b *testing.B) {
	buf, err := New(512)
	if err != nil {
		b.Fatalf("failed to create buffer: %v", err)
	}

	b.Cleanup(func() {
		err = buf.Close()
		if err != nil {
			b.Fatalf("failed to close buffer: %v", err)
		}
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if cap(*buf)-len(*buf) < 512 {
			b.Fatalf("buffer size is not correct: %d", cap(*buf)-len(*buf))
		}
	}
}

func BenchmarkPolyglotSizeCheck(b *testing.B) {
	randomBytes := make([]byte, 512)
	_, err := rand.Read(randomBytes)
	if err != nil {
		b.Fatalf("failed to read random bytes: %v", err)
	}

	buf := polyglot.GetBuffer()
	b.Cleanup(func() {
		polyglot.PutBuffer(buf)
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if cap(*buf)-len(*buf) < 512 {
			b.Fatalf("buffer size is not correct: %d", cap(*buf)-len(*buf))
		}
	}
}

func TestBuffer(t *testing.T) {
	buf, err := New(512)
	require.NoError(t, err)
	err = buf.Close()
	require.NoError(t, err)
}
