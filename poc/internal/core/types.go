package core

type Extension struct {
	Type  uint64
	Value []byte
}

type Envelope struct {
	Version   uint64
	ProfileID uint64
	MsgType   uint64
	Flags     uint64
	TsUnixMs  uint64
	MsgID     []byte

	Extensions []Extension
	Payload    []byte
}

type Limits struct {
	MaxFrameBytes   uint32
	MaxPayloadBytes uint32
	MinMsgIDBytes   int
	MaxMsgIDBytes   int
	MaxExtBytes     int
}

const (
	CoreVersion           = 1
	DefaultMaxFrameBytes  = 8 * 1024 * 1024
	DefaultMinMsgIDBytes  = 8
	DefaultMaxMsgIDBytes  = 64
	DefaultMaxExtBytes    = 4096
	DefaultMaxClockSkewMs = 300000
)

func DefaultLimits() Limits {
	return Limits{
		MaxFrameBytes:   DefaultMaxFrameBytes,
		MaxPayloadBytes: DefaultMaxFrameBytes,
		MinMsgIDBytes:   DefaultMinMsgIDBytes,
		MaxMsgIDBytes:   DefaultMaxMsgIDBytes,
		MaxExtBytes:     DefaultMaxExtBytes,
	}
}
