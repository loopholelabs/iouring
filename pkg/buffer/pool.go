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

import "sync"

const (
	defaultSize = 512
)

var (
	pool      = NewPool(defaultSize)
	fixedPool = NewFixedPool(defaultSize)
)

type Pool struct {
	pool sync.Pool
	size int64
}

func NewPool(size int64) *Pool {
	return &Pool{
		size: size,
	}
}

func (p *Pool) Get() (b *Buffer, err error) {
	v := p.pool.Get()
	if v == nil {
		b, err = New(p.size)
		return
	}
	return v.(*Buffer), nil
}

func (p *Pool) Put(b *Buffer) {
	if b != nil {
		b.Reset()
		p.pool.Put(b)
	}
}

func GetBuffer() (*Buffer, error) {
	return pool.Get()
}

func PutBuffer(b *Buffer) {
	pool.Put(b)
}

type FixedPool struct {
	pool sync.Pool
	size int64
}

func NewFixedPool(size int64) *FixedPool {
	return &FixedPool{
		size: size,
	}
}

func (p *FixedPool) Get() (b *Fixed, err error) {
	v := p.pool.Get()
	if v == nil {
		b, err = NewFixed(p.size)
		return
	}
	return v.(*Fixed), nil
}

func (p *FixedPool) Put(b *Fixed) {
	if b != nil {
		b.Reset()
		p.pool.Put(b)
	}
}

func GetFixed() (*Fixed, error) {
	return fixedPool.Get()
}

func PutFixed(b *Fixed) {
	fixedPool.Put(b)
}
