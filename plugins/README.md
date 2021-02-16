# VERAISON plugins

This directory contains the code for plugin implementations that come as part
of VERAISON. Plugins implement functionality such as support various store
backends and attestation token formats.

## Directory layout

Each subdirectory (except for `bin/` -- see later) contains code for a plugin.
The name of the directory indicates what type of the plugin it is.  The GNU
`make(1)` command can be invoked with one of the predefined targets (`all`,
`test`, `lint`, or `clean`), either at this level or in a plugin directory of
choice.  The `all` target will build the plugin binary and copy it into `bin/`
directory at this level, creating it if necessary.

## Plugin types

The following plugin types are currently supported, each implementing a
corresponding interface

| plugin type name  | plugin class            | implements interface |
| :---------------- | :---------------------- | :------------------- |
| endorsementstore  | EndorsementStorePlugin  | IEndrosementStore    |
| policyengine      | PolicyEnginePlugin      | IPolicyEngine        |
| policystore       | PolicyStorePlugin       | IPolicyStore         |
| evidenceextractor | EvidenceExtractorPlugin | IEvidenceExtractor   |
| trustanchorstore  | TrustAnchorStorePlugin  | ITrustAnchorStore    |

Definitions for these can be found under `../common/`

## Adding a new plugin

The plugins are implemented using the Hashicorp plugin framework:

    https://github.com/hashicorp/go-plugin

Each plugin is compiled into an executable and is run in a separate process,
removing the need to synchronise builds between the plugin and the core
application.

To implement a new plugin:

1. Create a directory for the plugin and initialize go module in it.
2. Create a main.go with an object implementing one of the interfaces above.
3. GetName() (which is included in all interfaces) should return the name of
   the plugin as as string. This name  is used in configuration to indicate
   which plugins should be loaded in a deployment. The name must be unique
   across plugins of the same time. To ensure portability across multiple
   deployments, it is recommended that you add a prefix unique to your
   organisation.
4. Create an instance of plugin.HandshakeConfig, setting the PluginVersion to
   1, MagicCookieKey to "VERAISON_PLUGIN" and MagicCookieValue to "VERAISON".
5. Create amap[string]plugin.Plugin. For each of the plugin types you've
   implemented, add a key with the plugin name (see table above), and an
   instance of the appropriate plugin object (again see table) as the value.
   St Impl attribute of the plugin object to a pointer to an instance of your
   object.
6. Add a main() that invokes plugin.Serve with anew plugin.ServerConfig
   instance, setting the HandshakeConfig and Plugins attributes to the values
   created in steps (4) and (5).

See existing plugin implementations for examples.


Note: it is possible to have more than one plugin contained in the same
executable --just add those

## Running plugin binaries independently

Even though the plugin binaries are regular ELF executables, they are not
intended to be run on their own. If you try to do so, you'll get a message to
that effect, and the binary will terminate. During debugging, however, it is
sometimes useful to be able to run the executables without the surrounding
code.

This check is performed by the plugin server initialization code based on the
magic cookie set in the handshake. In addition to the cookie, server code also
expects some plugin configuration variables to be set in the environment. The
snippet below gives an example of starting the plugin binary execution with a
debugger:

    export MY_COOKIE_KEY=my_cookie_value
    export PLUGIN_MIN_PORT=10000
    export PLUGIN_MAX_PORT=25000
    export PLUGIN_PROTOCOL_VERSIONS=1

    dlv test

The cookie variable and the protocol versions must match the values set in the
handshake (see above).
