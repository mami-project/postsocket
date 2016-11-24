package postsocket

import (
	"io"
)

type Local interface {
}

type Remote interface {
}

type Path interface {
}

type Receiver interface {
	Receive(msg []byte)
}

type SendEventHandler interface {
	HandleAck(association *Association, oid int)
	HandleExpired(association *Association, oid int)
}

type PathEventHandler interface {
	HandlePathUp(association *Association, path *Path)
	HandlePathDown(association *Association, path *Path)
	HandleDormant()
}

type Association interface {
	Send(msg []byte) (oid int, err error)
	Sendx(msg []byte, nice int, lifetime int, oid int, antecedents []int) (err error)
	OpenStream() (stream io.ReadWriteCloser, err error)
	SetReceiver(receiver Receiver)
	SetSendEventHandler(sevhandler SendEventHandler)
	SetPathEventHandler(pevhandler PathEventHandler)
}

type AcceptHandler interface {
	HandleAccept(listener *Listener, local *Local, remote *Remote)
}

type Listener interface {
	Accept(local *Local, remote *Remote, receiver Receiver) (association *Association, err error)
}

type PostImpl interface {
	Associate(local *Local, remote *Remote, receiver Receiver) (association *Association, err error)
	Listen(local *Local, achandler AcceptHandler) (listener *Listener, err error)
}
