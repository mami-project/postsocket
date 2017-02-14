package postsocket

import (
	"io"
	"crypto/tls"
	"time"
)

///////////////////////////////////////////////////////////
// Protocol Stack Instantiation MPI
///////////////////////////////////////////////////////////

// Encapsulates a connectable or connected instance of a protocol stack (e.g.
// TCP/IP, Websockets over TLS over TCP over IP, QUIC over PLUS over UDP over
// IP, and so on).
type ProtocolStackInstance interface {
	// Get a name for this instance for debugging purposes
	String()

	// Write a Message to this Instance
	Write(t Transient, msg Message) error

	// Read the next message into the given channel, or cancel when an object
	// is written to the cancellation channel
	Read(t Transient, mc chan-> Message, cancel chan<- struct{}) error

	// Ensure that the PSI is ready for reading and writing
	Start(t Transient, func() ready) error

	// Destroy this PSI
	Destroy(t Transient)
}

// Binds a stream to a protocol stack instance. 
type Transient interface {
	Path() PathDescriptor
}

///////////////////////////////////////////////////////////
// Pathfinder lower-level API -- probably not exposed
///////////////////////////////////////////////////////////

// A policy context describes a set
// of properties for "desirable" provisioning
// domains, protocol stack instances, etc.
// for a particular association.

// FIXME neat this up a bit.
type PolicyContext map[string]interface{}

// A path descriptor describes a single path through
// the network, including all addressing and 
// protocol stack instance information necessary to 
// establish an association on that path
type PathDescriptor interface {
	Identifier() string
	PolicyContext() PolicyContext
	LocalIdentity() LocalIdentity
	RemoteIdentity() RemoteIdentity	
}

// Given a policy describing which PvDs are acceptable,
// and a remote identity to connect to / rendezvous with,
// resolve the remote and determine possible
// Paths to between acceptable PvDs and the remote.
func CandidatePaths(pc PolicyContext, ri RemoteIdentity) []PathDescriptor

// Allocate an association given a policy between a local and a remote
func NewAssociation(pc PolicyContext, 
					li LocalIdentity, 
					ri RemoteIdentity) Association, error


///////////////////////////////////////////////////////////
// PostSockets API
///////////////////////////////////////////////////////////

type Association interface {
	PolicyContext() PolicyContext
	LocalIdentity() LocalIdentity
	RemoteIdentity() RemoteIdentity

	NewStream() Stream, error
}

// Contains a message along with metadata needed to send it
type OutMessage struct {
	Content Message
	Niceness uint
	Lifetime time.Duration
	Antecedent []*OutMessage
	Metadata map[string]interface{}
}

// FIXME does a in message have metadata? yes, PSI can fill this in...

// Contains a message received from a stream
type Message []byte

// A logical, two-way channel for messages
type Stream interface {
	// Retrieve the Association over which this stream is running
	Association() *Association

	// Send a message on this Stream
	Send(msg *OutMessage, fate func(err error))

	// Signal that the application is ready 
	// to receive a message on this stream to a given channel
	Receive(mc chan-> Message, cancel chan<- struct{})

	// Terminate the stream
	Close()
}

type Receiver interface {
	Receive(msg Message)
	Close()
}

type Accept func(stream *Stream)

type Respond func(msg Message, reply Reply)

type Reply func(msg *OutMessage, fate func(err error))

type Source interface {
	// Send a message via this Source
	Send(msg *OutMessage, fate func(err error))

	Close()
}

type PostTerminal interface {
	PolicyContext() PolicyContext

	NewStream(li LocalIdentity, ri RemoteIdentity, pc PolicyContext) Stream, error
	NewListener(li LocalIdentity, pc PolicyContext, accept Accept) io.Closer, error
	NewResponder(li LocalIdentity, pc PolicyContext, respond Respond) io.Closer, error
	NewSource(li LocalIdentity, ri RemoteIdentity, pc PolicyContext) Source, error
	NewSink(li LocalIdentity, pc PolicyContext, receiver Receiver) io.Closer, error
}

// Create a new PostSockets terminal. Terminals encapsulate 
// default policy contexts, and cache Associations for reuse.

func NewTerminal(pc PolicyContext) PostTerminal

// Encapsulate a local identity
type LocalIdentity struct {
	Port int 
	Interface string
	EndEntities []tls.Certificate
}

// Encapsulate a remote identity
type RemoteIdentity interface {
	AcceptableIssuers() []tls.Certificate
	AcceptableEndEntities() []tls.Certificate
	Resolve() []RemoteIdentity, error
}







