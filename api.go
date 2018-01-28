package postsocket

import (
	"crypto/tls"
	"io"
)

///////////////////////////////////////////////////////////////////////////////
// Post Sockets API definition
//
// This is a work in progress, attempting to track developments in
// https://datatracker.ietf.org/doc/draft-trammell-taps-post-sockets. See
// https://github.com/mami-project/postsocket/issues for issues identified
// with this abstract API that need a solution here and/or in the document.
///////////////////////////////////////////////////////////////////////////////

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

	// Create a new Source associated with a given Local and Remote. Calls
	// associated with receiving on a Source  will result in a runtime
	// error.
	NewSource(loc Local, rem Remote, cfg Configuration) (Carrier, error)

	// Create a new Sink associated with a given Local. Calls
	// associated with sending on a Sink will result in a runtime
	// error.
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

	// Create a new Local from a string specification, optionally bound to a
	// provisioning domain, optionally identified by zero or more TLS
	// certificates.
	NewLocal(spec string, pvd ProvisioningDomain, identities []tls.Certificate) (Local, error)

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

	// Iterate over the Paths this Association has properties for. Not all
	// paths are presently in use.
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

	// Iterate over the Paths this Carrier is currently using.
	Paths() []Path

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
	OnReceiveError(fn ReceiveErrorFunc)

	// Register a deframer
	DeframeWith(fn DeframeFunc)
}

// ProvisioningDomain contains information about a single connection this
// endpoint has to remote networks and/or the Internet.
//
// FIXME the methods defined here may actually be implementation-specific;
// in this case, the PvD in the abstract API might be an interface{}.
type ProvisioningDomain interface {
	// Create a Local bound to this Provisioning Domain; i.e., that will only
	// use the interface it encapsulates.
	LocalFor(spec string, identities []tls.Certificate)

	// Determine whether a given resolved Remote is reachable using this PvD
	CanReach(rem Remote) bool
}

// Local represents information about a single local endpoint, i.e. to which a
// Listener can be bound, and which may have a coherent security identity.
type Local interface {
	// Return a canonicalized string representing this Local's specification.
	String() string
	// Return the certificates by which this Local identifies itself
	Identities() []tls.Certificate
}

type Remote interface {
	// Return a canonicalized string representing this Remote's specification.
	String() string

	// Return the certificates by which this Remote identifies itself
	Identities() []tls.Certificate

	// Resolve this Remote to a further stage of resolution, optionally scoping that resolution
	Resolve(loc Local) ([]Remote, error)

	// Return the Remote from which this one was resolved, or nil if the Remote is not the result of resolution.
	Antecedent() Remote

	// Determine whether this Remove can be used for establishing a connection, or requires further resolution to do so
	Complete() bool
}

type Path interface {
	// Return the Local associated with this Path
	Local() Local

	// Return the Remote associated with this Path
	Remote() Remote

	// Return the current value of a named path property FIXME replace with core path properties
	Get(property string) interface{}
}

type Configuration interface {
	// Get the value of a named configuration property
	Get(key string) interface{}

	// Set the value of a named configuration property to a given value
	Set(key string, value interface{})
}

type Listener interface {
	// Close this listener; i.e., stop listening
	Close() error
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
