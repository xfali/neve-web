/*
 * Copyright (C) 2019-2024, Xiongfa Li.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package buffer

import (
	"bytes"
	"github.com/xfali/xlog"
	"io"
	"sync"
	"sync/atomic"
)

const (
	InitialBufferSize = 1024
	MaxBufferSize     = 4096
)

type Pool interface {
	Get() *bytes.Buffer
	Put(*bytes.Buffer)
}

type defaultPool struct {
	initialSize int
	maxSize     int
	pool        sync.Pool

	count int32
	Debug bool
}

type Opt func(*defaultPool)

func NewPool(opts ...Opt) *defaultPool {
	ret := &defaultPool{
		initialSize: InitialBufferSize,
		maxSize:     MaxBufferSize,
	}
	for _, opt := range opts {
		opt(ret)
	}
	ret.pool = sync.Pool{New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, ret.initialSize))
	}}
	return ret
}

func OptSetInitialBufferSize(size int) Opt {
	return func(pool *defaultPool) {
		pool.initialSize = size
	}
}

func OptSetMaxBufferSize(size int) Opt {
	return func(pool *defaultPool) {
		pool.maxSize = size
	}
}

func (p *defaultPool) Get() *bytes.Buffer {
	buf := p.pool.Get().(*bytes.Buffer)
	buf.Reset()
	if p.Debug {
		xlog.Errorf("pool %p get : %d %p", p, atomic.AddInt32(&p.count, 1), buf)
	}
	return buf
}

func (p *defaultPool) Put(buf *bytes.Buffer) {
	if buf == nil {
		return
	}
	if buf.Len() > p.maxSize {
		if p.Debug {
			xlog.Errorf("pool %p return and wait for GC : %d %p", p, atomic.AddInt32(&p.count, 1), buf)
		}
		return
	}
	p.pool.Put(buf)
	if p.Debug {
		xlog.Errorf("pool %p Put : %d %p", p, atomic.AddInt32(&p.count, 1), buf)
	}
}

type ContentLength interface {
	ContentLength() int64
}

type NopReadCloser struct {
	r   io.Reader
	len int64
}

func (rc *NopReadCloser) Read(p []byte) (n int, err error) {
	return rc.r.Read(p)
}

func (rc *NopReadCloser) Close() error {
	return nil
}

func (rc *NopReadCloser) ContentLength() int64 {
	return rc.len
}

func NewReadCloser(d []byte) *NopReadCloser {
	if d == nil {
		return nil
	}
	return &NopReadCloser{
		r:   bytes.NewReader(d),
		len: int64(len(d)),
	}
}

type ReadWriteCloser struct {
	pool Pool
	buf  *bytes.Buffer
	once sync.Once
}

func (rc *ReadWriteCloser) Bytes() []byte {
	return rc.buf.Bytes()
}

func (rc *ReadWriteCloser) Read(p []byte) (n int, err error) {
	return rc.buf.Read(p)
}

func (rc *ReadWriteCloser) Write(p []byte) (n int, err error) {
	return rc.buf.Write(p)
}

func (rc *ReadWriteCloser) Close() error {
	// just return once
	rc.once.Do(func() {
		rc.pool.Put(rc.buf)
	})
	return nil
}

func (rc *ReadWriteCloser) ContentLength() int64 {
	return int64(rc.buf.Len())
}

func NewReadWriteCloser(pool Pool) *ReadWriteCloser {
	buf := pool.Get()
	return &ReadWriteCloser{
		pool: pool,
		buf:  buf,
	}
}
