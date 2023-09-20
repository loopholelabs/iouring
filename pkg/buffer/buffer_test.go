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
		if num != len(randomBytes) {
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
		if num != len(randomBytes) {
			b.Fatalf("number of bytes written is not correct: %d", num)
		}
		buf.Reset()
	}
}

func BenchmarkBufferAllocationsNoResizePool(b *testing.B) {
	randomBytes := make([]byte, 512)
	_, err := rand.Read(randomBytes)
	if err != nil {
		b.Fatalf("failed to read random bytes: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var num int
		var buf *Buffer
		var err error
		for pb.Next() {
			buf, err = GetBuffer()
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
			PutBuffer(buf)
		}
	})
}

func BenchmarkPolyglotAllocationsNoResizePool(b *testing.B) {
	randomBytes := make([]byte, 512)
	_, err := rand.Read(randomBytes)
	if err != nil {
		b.Fatalf("failed to read random bytes: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var num int
		var buf *polyglot.Buffer
		for pb.Next() {
			buf = polyglot.GetBuffer()
			num = buf.Write(randomBytes)
			if num != len(randomBytes) {
				b.Fatalf("number of bytes written is not correct: %d", num)
			}
			polyglot.PutBuffer(buf)
		}
	})
}

func BenchmarkBufferAllocationsResize(b *testing.B) {
	randomBytes := make([]byte, 2048)
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
		if num != len(randomBytes) {
			b.Fatalf("number of bytes written is not correct: %d", num)
		}
		num, err = buf.Write(randomBytes)
		if err != nil {
			b.Fatalf("failed to write bytes: %v", err)
		}
		if num != len(randomBytes) {
			b.Fatalf("number of bytes written is not correct: %d", num)
		}
		buf.Reset()
	}
}

func BenchmarkPolyglotAllocationsResize(b *testing.B) {
	randomBytes := make([]byte, 2048)
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
		if num != len(randomBytes) {
			b.Fatalf("number of bytes written is not correct: %d", num)
		}
		num = buf.Write(randomBytes)
		if num != len(randomBytes) {
			b.Fatalf("number of bytes written is not correct: %d", num)
		}
		buf.Reset()
	}
}

func BenchmarkBufferAllocationsResizePool(b *testing.B) {
	randomBytes := make([]byte, 2048)
	_, err := rand.Read(randomBytes)
	if err != nil {
		b.Fatalf("failed to read random bytes: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var num int
		var buf *Buffer
		var err error
		for pb.Next() {
			buf, err = GetBuffer()
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
			num, err = buf.Write(randomBytes)
			if err != nil {
				b.Fatalf("failed to write bytes: %v", err)
			}
			if num != len(randomBytes) {
				b.Fatalf("number of bytes written is not correct: %d", num)
			}
			PutBuffer(buf)
		}
	})
}

func BenchmarkPolyglotAllocationsResizePool(b *testing.B) {
	randomBytes := make([]byte, 2048)
	_, err := rand.Read(randomBytes)
	if err != nil {
		b.Fatalf("failed to read random bytes: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var num int
		var buf *polyglot.Buffer
		for pb.Next() {
			buf = polyglot.GetBuffer()
			num = buf.Write(randomBytes)
			if num != len(randomBytes) {
				b.Fatalf("number of bytes written is not correct: %d", num)
			}
			num = buf.Write(randomBytes)
			if num != len(randomBytes) {
				b.Fatalf("number of bytes written is not correct: %d", num)
			}
			polyglot.PutBuffer(buf)
		}
	})
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
