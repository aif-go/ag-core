package aerrors

import "errors"

var (
	// ErrEmptyEngine occurs when trying to do something with an empty engine.
	ErrEmptyEngine = errors.New("agonet: the internal engine is empty")
	// ErrEngineShutdown occurs when server is closing.
	ErrEngineShutdown = errors.New("agonet: server is going to be shutdown")
	// ErrEngineInShutdown occurs when attempting to shut the server down more than once.
	ErrEngineInShutdown = errors.New("agonet: server is already in shutdown")
	// ErrAcceptSocket occurs when acceptor does not accept the new connection properly.
	ErrAcceptSocket = errors.New("agonet: accept a new connection error")
	// ErrTooManyEventLoopThreads occurs when attempting to set up more than 10,000 event-loop goroutines under LockOSThread mode.
	ErrTooManyEventLoopThreads = errors.New("agonet: too many event-loops under LockOSThread mode")
	// ErrUnsupportedProtocol occurs when trying to use protocol that is not supported.
	// ErrUnsupportedProtocol = errors.New("agonet: only unix, tcp/tcp4/tcp6, udp/udp4/udp6 are supported")
	ErrUnsupportedProtocol = errors.New("agonet: only unix, tcp/tcp4/tcp6, tlcp are supported")
	// ErrUnsupportedTCPProtocol occurs when trying to use an unsupported TCP protocol.
	ErrUnsupportedTCPProtocol = errors.New("agonet: only tcp/tcp4/tcp6 are supported")
	// ErrUnsupportedUDPProtocol occurs when trying to use an unsupported UDP protocol.
	// ErrUnsupportedUDPProtocol = errors.New("agonet: only udp/udp4/udp6 are supported")
	// ErrUnsupportedUDSProtocol occurs when trying to use an unsupported Unix protocol.
	ErrUnsupportedUDSProtocol = errors.New("agonet: only unix is supported")
	// ErrUnsupportedOp occurs when calling some methods that are either not supported or have not been implemented yet.
	ErrUnsupportedOp = errors.New("agonet: unsupported operation")
	// ErrNegativeSize occurs when trying to pass a negative size to a buffer.
	ErrNegativeSize = errors.New("agonet: negative size is not allowed")
	// ErrNoIPv4AddressOnInterface occurs when an IPv4 multicast address is set on an interface but IPv4 is not configured.
	ErrNoIPv4AddressOnInterface = errors.New("agonet: no IPv4 address on interface")
	// ErrInvalidNetworkAddress occurs when the network address is invalid.
	ErrInvalidNetworkAddress = errors.New("agonet: invalid network address")
	// ErrInvalidNetConn occurs when trying to do something with an empty net.Conn.
	ErrInvalidNetConn = errors.New("agonet: the net.Conn is empty")
	// ErrNilRunnable occurs when trying to execute a nil runnable.
	ErrNilRunnable = errors.New("agonet: nil runnable is not allowed")

	// ErrTLSConfigIsNil occurs when trying to use TLS without providing a valid TLSConfig.
	ErrTLSConfigIsNil = errors.New("agonet: TLSConfig is nil")

	// ErrTLCPConfigIsNil occurs when trying to use TLCP without providing a valid TLCPConfig.
	ErrTLCPConfigIsNil = errors.New("agonet: TLCPConfig is nil")
)
