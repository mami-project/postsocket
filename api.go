package postsocket

import (
	"crypto/tls"
	"io"
)

///////////////////////////////////////////////////////////
// Post Sockets API definition
///////////////////////////////////////////////////////////

// PostContext encapsulates all the state kept by the Post Sockets API at a
// single endpoint, and is the "root" of the API. It contains Policies and
// Configurations governing connection initiation and listening, available
// Locals, stored Associations and Paths, and currently open Carriers and
// Transients. An application will generally create a single PostContext
// instance on startup, optionally Load() state from disk, and can checkpoint
// in-memory state to disk with Save().
type PostContext interface {
	// Initiate a connection from a given Local to a given Remote, with an
	// optional Configuration to override association- and context-level
	// Conifigurations, returning a Carrier on which messages can be sent and
	// received.
	Initiate(loc Local, rem Remote, cfg Configuration) (Carrier, error)

	// Listen for connections on a given Local which will pass connection
	// requests to a given ListenFunc.
	Listen(loc Local, lfn ListenFunc, cfg Configuration) (Listener, error)

	// Create a new Source associated with a given Local and Remote
	NewSource(loc Local, rem Remote, cfg Configuration) (Carrier, error)

	// Create a new Sink associated with a given Local
	NewSink(loc Local, cfg Configuration) (Carrier, error)

	// Save this context's state to a file on disk.
	Save(filename string) error

	// Replace this context's state with state loaded from a file on disk.
	Load(filename string) error

	// Bind a configuration to this context
	Configure(cfg Configuration) error

	// Retrieve the configuration in effect for this context.
	Configuration() Configuration

	// Iterate over currently active provisioning domains.
	ProvisioningDomains() []ProvisioningDomain

	// Create a new Local, optionally bound to a provisioning domain,
	// optionally identified by zero or more TLS certificates, optionally
	// identified by one or more transport-layer ports
	NewLocal(pvd ProvisioningDomain, identities []tls.Certificate, ports []uint16) (Local, error)

	// Create a new Remote from a string specification. The Remote will be
	// optionally scoped for use with a given Local during resolution, and
	// optionally subject to a given identity check on any identity presented.
	NewRemote(spec string, loc Local, identityCheck IdentityCheckFn)
}

// Association encapsulates state about communications between a local
// endpoint and a remote endpoint over a set of paths. It includes local and
// remote identity information, cached cryptographic resumption parameters and
// cached path properties.
type Association interface {

	// Iterate over the Carriers currently bound to this Association
	Carriers() []Carrier

	// Iterate over the Paths this carrier knows about
	Paths() []Path

	// Initiate a new Carrier on this Association
	Initiate() (Carrier, error)

	// Bind a configuration to this Association. Association-level
	// configurations override context-level configurations where permitted by
	// policy.
	Configure(cfg Configuration) error

	// Return the Configuration in effect for this Association. This
	// configuration will include configuration directives inherited from the
	// PostContext in which the association was created.
	Configuration() Configuration
}

// Carrier provides an interface over which messages can be sent to and
// received from a remote endpoint.
type Carrier interface {
	// Access the association backing this Carrier
	Association() Association

	// Close this Carrier
	Close() error

	// Register an event handler to be notified when this carrier closes
	OnClosed(fn CloseEventFunc)

	// Send a Message on this carrier with given lifetime, niceness,
	// idempodence, and immediacy
	Send(msg []byte, lifetime int, nice int, idem bool, immed bool) error

	// Register an event handler to be notified when a message expires before
	// being sent
	OnExpired(fn SendEventFunc) error

	// Register an event handler to be notified when a message is sent
	OnSent(fn SendEventFunc) error

	// Register an event handler to be notified when a message is acknowledged
	OnAcked(fn SendEventFunc) error

	// Register a receiver that will be called when a message is received
	Ready(fn ReceiveFunc) error

	// Register an event handler that will be called when a receive error occurs
	OnError(fn ReceiveErrorFunc)

	// Register a deframer
	DeframeWith(fn DeframeFunc)
}

type ProvisioningDomain interface {
}

type Local interface {
}

type Remote interface {
	Resolve(loc Local) ([]Remote, error)
	Complete() bool
}

type Path interface {
}

type Configuration interface {
}

type Listener interface {
}

type ListenFunc func()

// DeframeFunc is used to deframe a byte stream into discrete messages. When
// Post Sockets is used with a transport protocol which does not support
// message boundary preservation, this allows an application to push message
// deframing down into the API implementation. DeframeFunc, when called,
// should read a single application-layer message into a byte slice, and leave
// the reader at the stream position of the start of the next message, if any.
type DeframeFunc func(in io.Reader) ([]byte, error)

type ReceiveFunc func(msg []byte, carrier Carrier)

type ReceiveErrorFunc func(err error)

// SendEventFunc is a callback type for events on a sent message.
type SendEventFunc func(msg []byte, carrier Carrier)

type CloseEventFunc func(carrier Carrier)

type IdentityCheckFn func(certs []tls.Certificate) error
