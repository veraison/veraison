package decoder

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

type Plugin struct {
	Impl IDecoder
}

func (p *Plugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &RPCServer{Impl: p.Impl}, nil
}

func (p *Plugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &RPCClient{client: c}, nil
}
