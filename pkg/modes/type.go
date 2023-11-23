package modes

type Type string

const (
	GetBlocks  Type = "GET_BLOCKS"
	LongSender Type = "LONG_SENDER"
	unknown    Type = "UNKNOWN"
)

// IsValidMode checks if the passed in mode is valid
func IsValidMode(runtime Type) bool {
	return runtime == GetBlocks || runtime == LongSender
}

// String returns a string representation of the mode
func (r Type) String() string {
	switch r {
	case GetBlocks:
		return string(GetBlocks)
	case LongSender:
		return string(LongSender)
	default:
		return string(unknown)
	}
}
