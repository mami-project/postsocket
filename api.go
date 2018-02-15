// Package postsocket specifies a Go interface for the TAPS API. This abstract
// interface to transport-layer service is described in
// https://taps-api.github.io/drafts/draft-trammell-taps-interface.html. For
// now, read that document to understand what's going on in this package.
// Eventually, this package will grow to contain a demonstration
// implementation of the API.
//
// A Note on Error Handling
//
// This API provides two ways for its client to learn of errors: through the
// Error event passed to a connection's EventHandler, and through error
// returns on various calls. In general, errors in networking are
// asynchronous, so almost every error involving the network will be passed
// through the Error event. The error returns on calls are therefore only used
// for immediately detectable errors, such as inconsistent arguments or
// states.
package postsocket

import (
	"crypto/tls"
	"io"
	"net"
	"time"
)

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

	// SetEventHandler sets the default connection handler for all
	// Connections created within this TransportContext.
	SetEventHandler(evh EventHandler)

	// SetFramingHandler sets the default framing handler for all
	// Connections created within this TransportContext.
	SetFramingHandler(fh FramingHandler)

	// Preconnect creates a Preconnection, which binds a connection and
	// framing handler to sets of related remote, local, transport and
	// security parameters (a Connection specifier) for Connection
	// instantiation. Any of Preconnect's arguments may  be nil, in which case
	// the default values for this TransportContext are used. Preconnection
	// allows the specification of multiple, disjoint sets of related
	// parameters for candidate transport protocol selection, as well as the
	// ability to initiate and send atomically for 0-RTT connection. The
	// Preconnection is initialized with one set of parameters; use
	// Preconnection.AddSpecifier to add more.

	Preconnect(evh EventHandler, fh FramingHandler, rem Remote, loc Local, tp TransportParameters, sp SecurityParameters) (Preconnection, error)

	// Initiate a connection with given remote, local, and parameters, and the
	// default connection and framing handlers. Any of these except Remote may
	// be nil, in which case context defaults will be used. Once the
	// Connection is initiated, the EventHandler's Ready callback will be
	// called with this connection and a nil antecedent. This is a shortcut
	// for creating a Preconnection with a single connection specifier and
	// initiating it.
	Initiate(rem Remote, loc Local, tp TransportParameters, sp SecurityParameters) (Connection, error)

	// Rendezvous with a given Remote using an appropriate peer to peer
	// rendezvous method, with
	// optional local, transport parameters, and security parameters. Each of
	// the optional arguments may be passed as nil; if so, the Context
	// defaults are used. Returns a Connection in the process of being
	// rendezvoused. The EventHandler's Ready callback will be called
	// with any established Connection(s), with a nil antecedent.
	// This is a shortcut for creating a Preconnection with a single
	// connection specifier and rendezvousing with it.
	Rendezvous(evh EventHandler, rem Remote, loc Local, tp TransportParameters, sp SecurityParameters) (Connection, error)

	// Listen on a given Local with a given Handler for connection events,
	// with optional transport and security parameters. Each of the optional
	// arguments may be passed as nil; if so, the Context defaults are used.
	// Returns a Listener in the process of being started. The
	// EventHandler's Ready callback will be called with any accepted
	// Connection(s), with this Connection as antecedent.
	Listen(evh EventHandler, loc Local, tp TransportParameters, sp SecurityParameters) (Connection, error)

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
// reported via the EventHandler when Intiate, Listen, or Rendezvous is
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
// resolution error will be reported via the EventHandler when Intiate,
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

// List of transport and security parameter names.
const (
	TransportFullyReliable = iota
	// ... and so on, FIXME fill this in. until then, see document for details
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
	// EventHandler's Ready callback will be called with this connection
	// and a nil antecedent.
	Initiate() (Connection, error)

	// Initiate a Connection with a Remote specified by this Preconnection,
	// using the Local and parameters supplied, while simultaneously sending
	// Content with the given SendParameters. Returns a connection in the
	// initiation process. Once the Connection is initiated, the
	// EventHandler's Ready callback will be called with this connection
	// and a nil antecedent.
	InitialSend(content interface{}, sp SendParameters) (Connection, error)

	// Rendezvous  using an appropriate peer to peer rendezvous method with a
	// Remote specified by this Preconnection, using the Local and parameters
	// supplied. Returns a connection in the rendezvous process. The
	// EventHandler's Ready callback will be called with any established
	// Connection(s), with a nil antecedent.
	Rendezvous() (Connection, error)

	// Listen for connections on the Local specified by this Preconnection
	// using the Local and parameters supplied. Returns a Listener in the
	// process of being started. The EventHandler's Ready callback will
	// be called with any accepted Connection(s), with this Connection as
	// antecedent.
	Listen() (Connection, error)
}

// Connection encapsulates a connection to another endpoint. All events on the
// connection will be passed to its associated EventHandler.
type Connection interface {

	// Send sends some content on this connection, with an optional content
	// reference, an object that will be used to refer to the content on any
	// event related to it, and a set of send parameters to govern how it will
	// be sent. If content is a []byte, the bytes it contains will be sent to
	// over the connection as a single Content. If content implements the
	// Content interface, the Bytes() method will be invoked and the resulting
	// []byte will be transmitted. Otherwise, content will be passed to the
	// framing handler's Frame() method to convert it to a []byte for
	// transmission.
	Send(content interface{}, contentref interface{}, sp SendParameters) error

	// Clone clones this Connection, creating a new Connection to the same
	// remote endpoint. If the underlying protocol stack supports
	// multistreaming, then this will create a new stream; otherwise, a new
	// transport connection (flow) will be created.
	Clone() (Connection, error)

	// Close closes this connection.
	Close() error

	// GetEventHandler returns this connection's event handler.
	GetEventHandler() EventHandler

	// SetEventHandler replaces this connection's event handler.
	SetEventHandler(eh EventHandler)

	// GetEventHandler returns this connection's framing handler.
	GetFramingHandler() FramingHandler

	// SetEventHandler replaces this connection's framing handler.
	SetFramingHandler(fh FramingHandler)
}

// Content provides the interface implemented by content passed to a Received
// event.
type Content interface {
	Bytes() []byte
}

// EventHandler defines the interface for connection event handlers.
type EventHandler interface {

	// Ready occurs when a new connection is ready for use. The conn argument
	// contains the new connection, and the ante argument contains the
	// connection from which this connection was created. In the case of
	// Connections opened with Initate() and Rendezvous(), ante will be nil.
	// In the case of passively-opened Connections (i.e., created after
	// Listen()), ante will contain the listening Connection. In the case of
	// Connections created by Clone(), ante will contain the connection on
	// which Clone() was called. In the case of Connections created because a
	// remote endpoint created new streams with a multistreaming transport
	// protocol, ante will contain a Connection wrapped around one of the
	// other streams.
	Ready(conn Connection, ante Connection)

	// Received occurs when Content had been received. The Content received is
	// given as an implementation of the Content interface, on which the
	// receiver can call Bytes(). For protocol stacks
	Received(content Content, conn Connection)

	// Sent occurs when Content has been sent. The contentref argument
	// contains the content reference given on Send().
	Sent(conn Connection, contentref interface{})

	// Expired occurs when a Content's expires without having been sent. The
	// contentref argument contains the content reference given on Send().
	Expired(conn Connection, contentref interface{})

	// Error occurs when an error occurs on a connection. If the error refers
	// to an attempt to send content, the contentref argument contains the
	// content reference given on Send(). Error is only occurs for errors
	// which do not also cause the connection to close.
	Error(conn Connection, contentref interface{}, err error)

	// Closed occurs when a connection is closed, either actively through
	// Close(), passively because the remote side ended the connection, or
	// because a connection-ending error occurred. In this last case, the
	// error is passed as the err argument.
	Closed(conn Connection, err error)
}

// FramingHandler defines the interface for application-assisted framing and deframing
type FramingHandler interface {
	// Frame converts a content object as passed to the Send() call into a
	// []byte to be passed down to the protocol stack.
	Frame(content interface{}) ([]byte, error)

	// Deframe reads the next object from a given reader, and returns it as an
	// object of a type implementing Content. Deframe will only be called when
	// receiving content via a transport protocol which does not provide its
	// own framing (e.g. TCP)
	Deframe(in io.Reader) (Content, error)
}
