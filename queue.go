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
	"runtime"
	"sync/atomic"
	"unsafe"
)

// CQNeedsFlush is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/queue.c#L42
func (r *Ring) CQNeedsFlush() bool {
	return atomic.LoadUint32(r.SQ.KFlags)&uint32(SQStatusCQOverflow|SQStatusTaskRun) != 0
}

// CQNeedsEnter is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/queue.c#L48
func (r *Ring) CQNeedsEnter() bool {
	return (r.Flags&uint32(SetupIOPoll)) != 0 || r.CQNeedsFlush()
}

// SQNeedsEnter is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/queue.c#L17
func (r *Ring) SQNeedsEnter(submit uint32, flags *uint32) bool {
	if submit == 0 {
		return false
	}
	if (r.Flags & uint32(SetupSQPoll)) == 0 {
		return true
	}

	if atomic.LoadUint32(r.SQ.KFlags)&uint32(SQStatusNeedWakeup) != 0 {
		*flags |= uint32(EnterSQWakeup)
		return true
	}

	return false
}

// FlushSQ is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/queue.c#L204
func (r *Ring) FlushSQ() uint32 {
	tail := r.SQ.SQETail
	if r.SQ.SQEHead != tail {
		r.SQ.SQEHead = tail
		atomic.StoreUint32(r.SQ.KTail, tail)
	}

	// There is a potential race condition here, left
	// intentionally because it will not cause any issues
	// https://github.com/axboe/liburing/blob/liburing-2.4/src/queue.c#L219
	return tail - *r.SQ.KHead
}

// GetCQEvent is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/queue.c#L135
func (r *Ring) GetCQEvent(submit uint32, waitNR uint32, sigmask *unix.Sigset_t) (*CQEvent, error) {
	data := GetData{
		Submit:   submit,
		WaitNR:   waitNR,
		GetFlags: 0,
		Size:     _NSIG / 8,
		HasTS:    0,
		Arg:      unsafe.Pointer(sigmask),
	}

	cqe, err := r._GetCQEvent(&data)
	runtime.KeepAlive(data)

	return cqe, err
}

// _GetCQEvent is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/queue.c#L62
func (r *Ring) _GetCQEvent(data *GetData) (cqe *CQEvent, err error) {
	looped := false
	ret := uint(0)
	localErr := error(nil)
	for {
		needEnter := false
		nrAvailable := uint32(0)
		flags := uint32(0)
		cqe, localErr = r._PeekCQEvent(&nrAvailable)
		if localErr != nil {
			if err == nil {
				err = localErr
			}
			break
		}

		if cqe == nil && data.WaitNR == 0 && data.Submit == 0 {
			if looped || !r.CQNeedsEnter() {
				if err == nil {
					err = unix.EAGAIN
				}
				break
			}
			needEnter = true
		}

		if data.WaitNR > nrAvailable || needEnter {
			flags = uint32(EnterGetEvents) | data.GetFlags
			needEnter = true
		}
		if r.SQNeedsEnter(data.Submit, &flags) {
			needEnter = true
		}
		if !needEnter {
			break
		}
		if looped && (data.HasTS != 0) {
			arg := (*GetEventsArg)(data.Arg)
			if cqe == nil && arg.TS != 0 && err == nil {
				err = unix.ETIME
			}
			break
		}

		if r.IntFlags&uint8(IntFlagRegRing) != 0 {
			flags |= uint32(EnterRegisteredRing)
		}
		ret, localErr = r.Enter2(data.Submit, data.WaitNR, flags, data.Arg, data.Size)
		if localErr != nil {
			if err == nil {
				err = localErr
			}
			break
		}
		data.Submit -= uint32(ret)
		if cqe != nil {
			break
		}
		if !looped {
			looped = true
			err = localErr
		}
	}

	return
}

// _Submit is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/queue.c#L368
func (r *Ring) _Submit(submitted uint32, waitNR uint32, getEvents bool) (ret uint, err error) {
	cqNeedsEnter := getEvents || waitNR != 0 || r.CQNeedsEnter()

	flags := uint32(0)
	if r.SQNeedsEnter(submitted, &flags) || cqNeedsEnter {
		if cqNeedsEnter {
			flags |= uint32(EnterGetEvents)
		}

		if r.IntFlags&uint8(IntFlagRegRing) != 0 {
			flags |= uint32(EnterRegisteredRing)
		}

		ret, err = r.Enter(submitted, waitNR, flags, nil)
	} else {
		ret = uint(submitted)
	}
	return
}

// _SubmitAndWait is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/queue.c#L390
func (r *Ring) _SubmitAndWait(waitNR uint32) (uint, error) {
	return r._Submit(r.FlushSQ(), waitNR, false)
}

// Submit is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/queue.c#L400
func (r *Ring) Submit() (uint, error) {
	return r._SubmitAndWait(0)
}

// SubmitAndGetEvents is defined here: https://github.com/axboe/liburing/blob/liburing-2.4/src/queue.c#L415
func (r *Ring) SubmitAndGetEvents() (uint, error) {
	return r._Submit(r.FlushSQ(), 0, true)
}
