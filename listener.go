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
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"net"
	"syscall"
)

var _ net.Listener = (*Listener)(nil)

const (
	AcceptEntries = 256
)

type Event uint64

const (
	EventAccept Event = iota
	EventRead
	EventWrite
	EventClose
)

type Listener struct {
	ring *Ring
}

func NewListener(addr string) (*Listener, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		return nil, fmt.Errorf("error while resolving listen address: %w", err)
	}

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, fmt.Errorf("error while opening listening socket: %w", err)
	}
	if fd < 0 {
		return nil, fmt.Errorf("error while opening listening socket: fd is %d", fd)
	}

	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, unix.SO_REUSEADDR, 1)
	if err != nil {
		return nil, fmt.Errorf("error while setting SO_REUSEADDR on listening socket with fd %d: %w", fd, err)
	}

	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1)
	if err != nil {
		return nil, fmt.Errorf("error while setting SO_REUSEPORT on listening socket with fd %d: %w", fd, err)
	}

	err = syscall.Bind(fd, &syscall.SockaddrInet4{
		Port: tcpAddr.Port,
		Addr: *(*[4]byte)(tcpAddr.IP),
	})
	if err != nil {
		return nil, fmt.Errorf("error binding listening socket with fd %d to listen address %s: %w", fd, addr, err)
	}

	err = syscall.SetNonblock(fd, false)
	if err != nil {
		return nil, fmt.Errorf("error while setting listening socket with fd %d to blocking: %w", fd, err)
	}

	err = syscall.Listen(fd, AcceptEntries/2)
	if err != nil {
		return nil, fmt.Errorf("error while starting to listen on socket with fd %d: %w", fd, err)
	}

	ring := NewRing()
	err = ring.QueueInit(AcceptEntries, 0)
	if err != nil {
		return nil, fmt.Errorf("error while creating ringbuffer for listening socket with fd %d: %w", fd, err)
	}

	clientAddr := NewClientAddress()
	sqe := ring.GetSQEntry()
	sqe.PrepareAccept(fd, clientAddr.AddressPointer, clientAddr.LengthPointer, 0)
	sqe.UserData = uint64(EventAccept)

	cqeNR, err := ring.Submit()
	if err != nil {
		return nil, fmt.Errorf("error while submitting SQE for listening socket with fd %d: %w", fd, err)
	}

	if cqeNR != 1 {
		return nil, fmt.Errorf("error while submitting SQE for listening socket with fd %d: expected 1 CQE, got %d", fd, cqeNR)
	}

	for {
		cqe, err := ring.WaitCQEvent()
		if errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EINTR) || errors.Is(err, syscall.ETIME) {
			return nil, fmt.Errorf("error while waiting for CQE for listening socket with fd %d: %w", fd, err)
		}
		if err != nil {
			return nil, fmt.Errorf("error while waiting for CQE for listening socket with fd %d: unknown - %w", fd, err)
		}

		ring.CQESeen(cqe)
	}

	err = ring.Close()
	if err != nil {
		return nil, fmt.Errorf("error while closing ringbuffer for listening socket with fd %d: %w", fd, err)
	}

	return &Listener{
		ring: ring,
	}, nil
}

func (l *Listener) Accept() (net.Conn, error) {
	//TODO implement me
	panic("implement me")
}

func (l *Listener) Close() error {
	//TODO implement me
	panic("implement me")
}

func (l *Listener) Addr() net.Addr {
	//TODO implement me
	panic("implement me")
}
