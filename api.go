package postsocket

import (
	"crypto/tls"
	"io"
)

///////////////////////////////////////////////////////////////////////////////
// Post Sockets API definition
// Transport Services (TAPS) Abstract Interface edition
/////////////////////////////////////////////////////////////////////////////////

// TransportContext encapsulates all the state kept by the API at a single
// endpoint, and is the "root" of the API. It can be used to create new
// default transport and security parameters, locals and remotes bound to the
// available provisioning domains and resolution, as well as new Connections
// with these properties. It stores path (address pair) and association
// (endpoint pair) properties and ephemeral as well as durable state. An
// application will generally create a single TransportContext instance on
// startup, optionally Load() state from disk, and can checkpoint in-memory
// state to disk with Save().
type TransportContext interface {
	NewTransportParameters() TransportParameters
	NewSecurityParameters() SecurityParameters
	NewRemote() Remote
	NewLocal() Local

	DefaultSendParameters() SendParameters

	// Initiate a connection to a given Remote, with a given Handler for
	// connection events, with optional local, transport parameters, and
	// security parameters. Each of the optional arguments may be passed as
	// nil; if so, the Context defaults are used. Returns a Connection in the
	// process of being initiated. Once the Connection is initiated, the
	// ConnectionHandler's Ready callback will be called with this connection
	// and a nil antecedent.
	Initiate(ch ConnectionHandler, rem Remote, loc Local, tp TransportParameters, sp SecurityParameters) (Connection, error)

	// Rendezvous with a given Remote using an appropriate peer to peer
	// rendezvous method, with a given Handler for connection events, with
	// optional local, transport parameters, and security parameters. Each of
	// the optional arguments may be passed as nil; if so, the Context
	// defaults are used. Returns a Connection in the process of being
	// rendezvoused. The ConnectionHandler's Ready callback will be called
	// with any established Connection(s), with this Connection as antecedent.
	Rendezvous(ch ConnectionHandler, rem Remote, loc Local, tp TransportParameters, sp SecurityParameters) (Connection, error)

	// Listen on a given Local with a given Handler for connection events,
	// with optional transport and security parameters. Each of the optional
	// arguments may be passed as nil; if so, the Context defaults are used.
	// Returns a Listener in the process of being started. The
	// ConnectionHandler's Ready callback will be called with any accepted
	// Connection(s), with this Connection as antecedent.
	Listen(ch ConnectionHandler, loc Local, tp TransportParameters, sp SecurityParameters) (Connection, error)

	// Save this context's state to a file on disk. The format of this state
	// file is not specified and not necessarily portable across
	// implementations of the API.
	Save(filename string) error

	// Replace this context's state with state loaded from a file on disk. The
	// format of this state file is not specified and not necessarily portable
	// across implementations of the API.
	Restore(filename string) error
}

type TransportParameter int

const (
	TransportFullyReliable = iota
	// ... and so on
)

type SecurityParameter int

const (
	SecuritySupportedGroup = iota
	SecurityCiphersuite
	SecuritySignatureAlgorithm
	// ... and so on
)

type TransportParameters interface {
	Require(p TransportParameter, v int) TransportParameters
	Prefer(p TransportParameter, v int) TransportParameters
	Avoid(p TransportParameter, v int) TransportParameters
	Prohibit(p TransportParameter, v int) TransportParameters
}

type SecurityParameters interface {
	AddIdentity(c tls.Certificate) SecurityParameters
	AddPSK(c tls.Certificate, k []byte) SecurityParameters
	VerifyTrustWith(func() (bool, error)) SecurityParameters     // FIXME needs useful args
	HandleChallengeWith(func() (bool, error)) SecurityParameters // FIXME needs useful args
	Require(p SecurityParameter, v int) SecurityParameters
	Prefer(p SecurityParameter, v int) SecurityParameters
	Avoid(p SecurityParameter, v int) SecurityParameters
	Prohibit(p SecurityParameter, v int) SecurityParameters
}

type SendParameters struct {
	ContentRef         interface{}
	Lifetime           uint
	Niceness           uint
	Ordered            bool
	Immediate          bool
	Idempotent         bool
	CorruptionTolerant bool
}

type Content interface {
	Bytes() []byte
}

type Remote interface {
}

type Local interface {
}

type Connection interface {
	Send(c interface{}, sp SendParameters) error
	Clone() (Connection, error)
}

type Listener interface {
}

type ContentRef interface {
}

type ConnectionHandler struct {
	Ready           func(conn Connection, ante Connection)
	Received        func(content Content, conn Connection)
	Sent            func(conn Connection, cr interface{})
	Expired         func(conn Connection, cr interface{})
	SendError       func(conn Connection, err error)
	ReceiveError    func(conn Connection, err error)
	InitiateError   func(conn Connection, err error)
	RendezvousError func(conn Connection, err error)
	ListenError     func(conn Connection, err error)
	Frame           func(content interface{}) ([]byte, error)
	Deframe         func(in io.Reader) (Content, error)
}
