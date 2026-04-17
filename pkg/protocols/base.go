package protocols

import "fmt"

type ProtocolType string

const (
	ProtocolMCP ProtocolType = "mcp"
	ProtocolA2A ProtocolType = "a2a"
	ProtocolANP ProtocolType = "anp"
)

type Protocol interface {
	GetProtocolName() string
	GetVersion() string
}

type BaseProtocol struct {
	ProtocolType_ ProtocolType
	Version_      string
}

func NewBaseProtocol(protocolType ProtocolType, version string) *BaseProtocol {
	if version == "" {
		version = "1.0.0"
	}
	return &BaseProtocol{
		ProtocolType_: protocolType,
		Version_:      version,
	}
}

func (p *BaseProtocol) GetProtocolName() string {
	return string(p.ProtocolType_)
}

func (p *BaseProtocol) GetVersion() string {
	return p.Version_
}

func (p *BaseProtocol) String() string {
	return fmt.Sprintf("%s(version=%s)", p.GetProtocolName(), p.GetVersion())
}
