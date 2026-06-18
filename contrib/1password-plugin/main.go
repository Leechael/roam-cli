package main

import (
	"github.com/1Password/shell-plugins/sdk/rpc/proto"
	"github.com/1Password/shell-plugins/sdk/rpc/server"
	"github.com/1Password/shell-plugins/sdk/schema"
	"github.com/hashicorp/go-plugin"
)

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  proto.Version,
			MagicCookieKey:   proto.MagicCookieKey,
			MagicCookieValue: proto.MagicCookieValue,
		},
		Plugins: plugin.PluginSet{
			"plugin": &server.RPCPlugin{RPCPlugin: func() (schema.Plugin, error) {
				return New(), nil
			}},
		},
	})
}
