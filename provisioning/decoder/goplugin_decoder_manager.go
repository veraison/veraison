package decoder

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/hashicorp/go-plugin"
)

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "VERAISON_PROVISIONING_DECODER_PLUGIN",
	MagicCookieValue: "VERAISON",
}

// pluginMap is the map of plugins we can dispense.
var pluginMap = map[string]plugin.Plugin{
	"decoder": &Plugin{},
}

type GoPluginDecoderManager struct {
	DispatchTable map[string]*GoPluginDecoderContext
}

func (o *GoPluginDecoderManager) Init(dir string) error {
	// TODO(tho) might want to define a naming convention for endorsement
	// decoder plugins
	pPaths, err := plugin.Discover("*", dir)
	if err != nil {
		return err
	}

	o.DispatchTable = make(map[string]*GoPluginDecoderContext)

	for _, p := range pPaths {

		ctx, err := NewGoPluginDecoderContext(p)
		if err != nil {
			return err
		}

		for _, mt := range ctx.supportedMediaTypes {
			// TODO(tho) check if this same media type has been already
			// advertised by another plugin.  Should raise fatal error if this
			// is the case.
			o.DispatchTable[mt] = ctx
		}
	}

	return nil
}

func (o GoPluginDecoderManager) Close() error {
	for _, v := range o.DispatchTable {
		if v.client != nil {
			log.Printf("killing client %s", v.name)
			v.client.Kill()
		}
	}
	return nil
}

func (o GoPluginDecoderManager) Dispatch(
	mediaType string,
	data []byte,
) (*EndorsementDecoderResponse, error) {
	ctx, ok := o.DispatchTable[mediaType]
	if !ok || ctx.handle == nil {
		return nil, fmt.Errorf("no active plugin found for media type %s", mediaType)
	}

	return ctx.handle.Decode(data)
}

func (o GoPluginDecoderManager) IsSupportedMediaType(mediaType string) bool {
	_, ok := o.DispatchTable[mediaType]

	return ok
}

func (o GoPluginDecoderManager) SupportedMediaTypes() string {
	a := make([]string, len(o.DispatchTable))

	for k := range o.DispatchTable {
		a = append(a, k)
	}

	return strings.Join(a, ", ")
}

type GoPluginDecoderContext struct {
	path                string
	name                string
	supportedMediaTypes []string
	handle              IDecoder
	client              *plugin.Client
}

func NewGoPluginDecoderContext(path string) (*GoPluginDecoderContext, error) {
	client := plugin.NewClient(
		&plugin.ClientConfig{
			HandshakeConfig: handshakeConfig,
			Plugins:         pluginMap,
			Cmd:             exec.Command(path),
		},
	)

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf(
			"unable to create the RPC client for %s: %w",
			path, err,
		)
	}

	protocolClient, err := rpcClient.Dispense("decoder")
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf(
			"unable to create a new instance of plugin %s: %w",
			path, err,
		)
	}

	handle, ok := protocolClient.(IDecoder)
	if !ok {
		client.Kill()
		return nil, fmt.Errorf(
			"plugin %s does not provide an implementation of the endorsement decoder interface",
			path,
		)
	}

	return &GoPluginDecoderContext{
		path:                path,
		name:                handle.GetName(),
		supportedMediaTypes: handle.GetSupportedMediaTypes(),
		handle:              handle,
		client:              client,
	}, nil
}
