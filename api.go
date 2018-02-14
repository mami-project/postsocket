package postsocket

import (
	"crypto/tls"
	"io"
	"net"
)

///////////////////////////////////////////////////////////////////////////////
// Post Sockets API
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
	// NewTransportParameters creates a new TransportParameters object with
	// system and user defaults for this TransportContext. The specification
	// of system and user defaults is implementation-specific.
	NewTransportParameters() TransportParameters

	// NewSecurityParameters creates a new SecurityParameters object with
	// system and user defaults for this TransportContext. The specification
	// of system and user defaults is implementation-specific.
	NewSecurityParameters() SecurityParameters

	// NewRemore creates a new, empty Remote specifier.
	NewRemote() Remote

	// NewLocal creates a new Local specifier initialized with defaults for
	// this TransportContext.
	NewLocal() Local

	// DefaultSendParameters returns a SendParameters object with
	DefaultSendParameters() SendParameters

	// Initiate a connection to a given Remote, with a given Handler for
	// connection events, with optional local specifier, transport parameters,
	// and security parameters. Each of the optional arguments may be passed
	// as nil; if so, the Context defaults are used. Returns a Connection in
	// the process of being initiated. Once the Connection is initiated, the
	// ConnectionHandler's Ready callback will be called with this connection
	// and a nil antecedent.
	Initiate(
		ch ConnectionHandler,
		rem Remote,
		loc Local,
		tp TransportParameters,
		sp SecurityParameters) (Connection, error)

	// Rendezvous with a given Remote using an appropriate peer to peer
	// rendezvous method, with a given Handler for connection events, with
	// optional local, transport parameters, and security parameters. Each of
	// the optional arguments may be passed as nil; if so, the Context
	// defaults are used. Returns a Connection in the process of being
	// rendezvoused. The ConnectionHandler's Ready callback will be called
	// with any established Connection(s), with this Connection as antecedent.
	Rendezvous(
		ch ConnectionHandler,
		rem Remote,
		loc Local,
		tp TransportParameters,
		sp SecurityParameters) (Connection, error)

	// Listen on a given Local with a given Handler for connection events,
	// with optional transport and security parameters. Each of the optional
	// arguments may be passed as nil; if so, the Context defaults are used.
	// Returns a Listener in the process of being started. The
	// ConnectionHandler's Ready callback will be called with any accepted
	// Connection(s), with this Connection as antecedent.
	Listen(
		ch ConnectionHandler,
		loc Local,
		tp TransportParameters,
		sp SecurityParameters) (Connection, error)

	// Save this context's state to a file on disk. The format of this state
	// file is not specified and not necessarily portable across
	// implementations of the API.
	Save(filename string) error

	// Replace this context's state with state loaded from a file on disk. The
	// format of this state file is not specified and not necessarily portable
	// across implementations of the API.
	Restore(filename string) error
}

type ParameterIdentifier int

const (
	TransportFullyReliable = iota
	SecuritySupportedGroup
	SecurityCiphersuite
	SecuritySignatureAlgorithm
	// ... and so on
)

type TransportParameters interface {
	Require(p ParameterIdentifier, v int) TransportParameters
	Prefer(p ParameterIdentifier, v int) TransportParameters
	Avoid(p ParameterIdentifier, v int) TransportParameters
	Prohibit(p ParameterIdentifier, v int) TransportParameters
}

type SecurityParameters interface {
	AddIdentity(c tls.Certificate) SecurityParameters
	AddPSK(c tls.Certificate, k []byte) SecurityParameters
	VerifyTrustWith(func() (bool, error)) SecurityParameters     // FIXME needs useful args
	HandleChallengeWith(func() (bool, error)) SecurityParameters // FIXME needs useful args
	Require(p ParameterIdentifier, v int) SecurityParameters
	Prefer(p ParameterIdentifier, v int) SecurityParameters
	Avoid(p ParameterIdentifier, v int) SecurityParameters
	Prohibit(p ParameterIdentifier, v int) SecurityParameters
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

// Remote specifies a remote endpoint by hostname, address, port, and/or
// service name. Multiple of each of these may be given; this will result in a
// set of candidate endpoints assumed to be equivalent from the application's
// standpoint to be resolved and connected to. Resolution of the remote need
// not occur until a connection is created; any resolution error will be
// reported via the ConnectionHandler when Intiate, Listen, or Rendezvous is
// called.
type Remote interface {
	// Return a remote specifier with the given hostname added to this specifier.
	WithHostname(hostname string) Remote

	// Return a remote specifier with the given IPv4 or IPv6 address added to this specifier
	WithAddress(address net.IP) Remote

	// Return a remote specifier with the given transport port added to this specifier
	WithPort(port uint16) Remote

	// Return a remote specifier with the given service name added to this specifier
	WithServiceName(svc string) Remote
}

// Local specifies a remote endpoint by interface name, hostname, address,
// port, and/or service name. Multiple of each of these may be given; this
// will result in a set of candidate endpoints assumed to be equivalent from
// the application's standpoint to be connected from or listened on. Any
// resolution error will be reported via the ConnectionHandler when Intiate,
// Listen, or Rendezvous is called.
type Local interface {
	// Return a local specifier with the given local network interface name or alias added to this specifier
	WithInterface(iface string) Local

	// Return a local specifier with the given hostname added to this specifier
	WithHostname(hostname string) Local

	// Return a local specifier with the given IPv4 or IPv6 address added to this specifier
	WithAddress(address net.IP) Local

	// Return a local specifier with the given transport port added to this specifier
	WithPort(port uint16) Local

	// Return a local specifier with the given service name added to this specifier
	WithServiceName(svc string) Local
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
	Ready    func(conn Connection, ante Connection)
	Received func(content Content, conn Connection)
	Sent     func(conn Connection, contentref interface{})
	Expired  func(conn Connection, contentref interface{})
	Error    func(conn Connection, contentref interface{}, err error)
	Frame    func(content interface{}) ([]byte, error)
	Deframe  func(in io.Reader) (Content, error)
}
