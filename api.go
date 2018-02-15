package postsocket

import (
	"crypto/tls"
	"io"
	"net"
	"time"
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

	// DefaultSendParameters returns a SendParameters object with default values.
	DefaultSendParameters() SendParameters

	// Preconnect creates a Preconnection, a container for sets of related
	// remote, local, transport and security parameters (a Connection
	// specifier), which can be instantiated into a Connection. Preconnection
	// allows the specification of multiple, disjoint sets of related
	// parameters for candidate transport protocol selection, as well as the
	// ability to initiate and send atomically for 0-RTT connection. The
	// Preconnection is initialized with one set of parameters; use Preconnection.
	Preconnect(ch ConnectionHandler, rem Remote, loc Local, tp TransportParameters, sp SecurityParameters) (Preconnection, error)

	// Initiate a connection with a given handler, remote, local, and
	// parameters. Any of these except Remote may be nil, in which case
	// context defaults will be used. Once the Connection is initiated, the
	// ConnectionHandler's Ready callback will be called with this connection
	// and a nil antecedent. This is a shortcut for creating a Preconnection
	// with a single connection specifier and initiating it.
	Initiate(ch ConnectionHandler, rem Remote, loc Local, tp TransportParameters, sp SecurityParameters) (Connection, error)

	// Rendezvous with a given Remote using an appropriate peer to peer
	// rendezvous method, with a given Handler for connection events, with
	// optional local, transport parameters, and security parameters. Each of
	// the optional arguments may be passed as nil; if so, the Context
	// defaults are used. Returns a Connection in the process of being
	// rendezvoused. The ConnectionHandler's Ready callback will be called
	// with any established Connection(s), with a nil antecedent.
	// This is a shortcut for creating a Preconnection with a single
	// connection specifier and rendezvousing with it.
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

// ParameterIdentifier identifies a Transport or Security Parameter
type ParameterIdentifier int

// List of transport and security parameter names. FIXME reference to documentation
const (
	TransportFullyReliable = iota
	SecuritySupportedGroup
	SecurityCiphersuite
	SecuritySignatureAlgorithm
	// ... and so on
)

// TransportParameters contains a set of parameters used in the selection of
// transport protocol stacks and paths during connection pre-establishment.
// Get a new TransportParameters bound to a context with
// NewTransportParameters(), then set preferences through the Require(),
// Prefer(), Avoid(), and Prohibit() methods.
type TransportParameters interface {

	// Require protocols and paths selected to fulfill this parameter. If
	// no protocols and paths available fulfill this parameter, then no
	// connection is possible. v is an optional value whose meaning is
	// parameter-specific.
	Require(p ParameterIdentifier, v int) TransportParameters

	// Prefer protocols and paths selected to fulfill this parameter.
	// Preferences are considered after requirements and prohibitions. v is an
	// optional value whose meaning is parameter-specific.
	Prefer(p ParameterIdentifier, v int) TransportParameters

	// Avoid protocols and paths selected to fulfill this parameter.
	// Avoidences are considered after requirements and prohibitions. v is an
	// optional value whose meaning is parameter-specific.
	Avoid(p ParameterIdentifier, v int) TransportParameters

	// Prohibit the selection of protocols and paths that fulfill this
	// parameter. If the protocols and paths available all fulfill this
	// parameter, then no connection is possible. v is an optional value
	// whose meaning is parameter-specific.
	Prohibit(p ParameterIdentifier, v int) TransportParameters
}

// SecurityParameters contains a set of parameters used in the establishment
// of security associations. Get a new SecurityParameters bound to a context
// with NewSecurityParameters(), then set preferences and associate identity
// and callbacks through the methods on the object.
type SecurityParameters interface {

	// AddIdentity adds an local identity (as an X.509 certificate with
	// private key) to this parameter set.
	AddIdentity(c tls.Certificate) SecurityParameters

	// AddPSK adds an preshared key associated with the given certificate to
	// the parameter set.
	AddPSK(c tls.Certificate, k []byte) SecurityParameters

	// VerifyTrustWith registers a callback to verify trust. This callback
	// takes a certificate and returns true if the certificate is trusted.
	VerifyTrustWith(func(c tls.Certificate) (bool, error)) SecurityParameters

	// HandleChallengeWith registers a callback to handle identity challenges.
	// FIXME needs useful args
	HandleChallengeWith(func() (bool, error)) SecurityParameters

	// Require any established security association to fulfill this parameter.
	// If no available security associations fulfill this parameter, then no
	// connection is possible. v is an optional value whose meaning is
	// parameter-specific.
	Require(p ParameterIdentifier, v int) SecurityParameters

	// Prefer to establish security associations that fulfill this parameter.
	// Preferences are considered after requirements and prohibitions. v is an
	// optional value whose meaning is parameter-specific.
	Prefer(p ParameterIdentifier, v int) SecurityParameters

	// Avoid the establishment of security associations that fulfill this parameter.
	// Avoidences are considered after requirements and prohibitions. v is an
	// optional value whose meaning is parameter-specific.
	Avoid(p ParameterIdentifier, v int) SecurityParameters

	// Prohibit the establishment of security associations that fulfill this
	// parameter. If the security associations available all fulfill this
	// parameter, then no connection is possible. v is an optional value
	// whose meaning is parameter-specific.
	Prohibit(p ParameterIdentifier, v int) SecurityParameters
}

// SendParameters contains a set of parameters used for sending content.
// DefaultSendParameters() returns the defaults for this context.
type SendParameters struct {
	// ContentRef is an opaque object referencing a Content that will be
	// passed on events pertaining to the Content.
	ContentRef interface{}
	// Lifetime after which the object is no longer relevant. Used for
	// unreliable and partially reliable transports; set to zero or less to
	// specify fully reliable transport, if available.
	Lifetime time.Duration
	// Niceness is the inverse priority of this Content relative to others on
	// this Connection or within this ConnectionGroup. Niceness 0 messages are
	// the highest priority.
	Niceness uint
	// Ordered is true if the Content must be sent before the next Content
	// sent on this Connection
	Ordered bool
	// Immediate is true if this Content should not be held for coalescing
	// with other Content in a transport-layer datagram.
	Immediate bool
	// Idempotent is true if this Content may be sent to the application more
	// than once without ill effects. Use this together with SendInitial() for
	// 0RTT session resumption.
	Idempotent bool
	// CorruptionTolerant is true if this Content may be sent to the
	// application even if checksums fail; it is used to explicitly disable
	// checksums on sent content.
	CorruptionTolerant bool
}

// Preconnection is a container for sets of related remote, local, transport
// and security parameters (a Connection specifier), which can be instantiated
// into a Connection. Use Preconnect() to create one with an initial
type Preconnection interface {

	// AddSpecifier adds a related set of remote, local, transport and
	// security parameters to this Preconnection.
	AddSpecifier(rem Remote, loc Local, tp TransportParameters, sp SecurityParameters)

	// Initiate a Connection with a Remote specified by this Preconnection,
	// using the Local and parameters supplied. Returns a connection in the
	// initiation process. Once the Connection is initiated, the
	// ConnectionHandler's Ready callback will be called with this connection
	// and a nil antecedent.
	Initiate() (Connection, error)

	// Initiate a Connection with a Remote specified by this Preconnection,
	// using the Local and parameters supplied, while simultaneously sending
	// Content with the given SendParameters. Returns a connection in the
	// initiation process. Once the Connection is initiated, the
	// ConnectionHandler's Ready callback will be called with this connection
	// and a nil antecedent.
	InitialSend(content interface{}, sp SendParameters) (Connection, error)

	// Rendezvous  using an appropriate peer to peer rendezvous method with a
	// Remote specified by this Preconnection, using the Local and parameters
	// supplied. Returns a connection in the rendezvous process. The
	// ConnectionHandler's Ready callback will be called with any established
	// Connection(s), with a nil antecedent.
	Rendezvous() (Connection, error)

	// Listen for connections on the Local specified by this Preconnection
	// using the Local and parameters supplied. Returns a Listener in the
	// process of being started. The ConnectionHandler's Ready callback will
	// be called with any accepted Connection(s), with this Connection as
	// antecedent.
	Listen() (Connection, error)
}

type Connection interface {
	Send(c interface{}, sp SendParameters) error
	Clone() (Connection, error)
	Close() error
}

type Content interface {
	Bytes() []byte
}

type ConnectionHandler struct {
	Ready    func(conn Connection, ante Connection)
	Received func(content Content, conn Connection)
	Sent     func(conn Connection, contentref interface{})
	Expired  func(conn Connection, contentref interface{})
	Error    func(conn Connection, contentref interface{}, err error)
	Closed   func(conn Connection)
	Frame    func(content interface{}) ([]byte, error)
	Deframe  func(in io.Reader) (Content, error)
}
