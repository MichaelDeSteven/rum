package rum

// ErrorType is an unsigned 64-bit error code as defined in the rum spec.
type ErrorType uint64

// Error represents a error's specification.
type Error struct {
	Err  error
	Type ErrorType
	Meta interface{}
}

type errorMsgs []*Error
