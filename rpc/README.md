# Mistify Agent RPC #

The Agent uses JSON-RPC over HTTP to communicate with sub-agents. The Agent uses `POST` only to the URI `/_mistify_RPC_`

There are currently two major types of RPC services that the Agent uses.  The Storage RPC and Guest services.

The data structures used in the RPC are documented using `godoc` in
the `mistify-agent/rpc` package.

## Storage RPC ##

The Storage RPC is used communicate with the Storage sub-agent.  The
Mistify team maintains the reference implentation at
https://github.com/mistifyio/mistify-agent-image/ . Currently, only
one Storage sub-agent may be used for image management.  The Storage
sub-agent is also generally used in Guest pipelines as well.

## Guest RPC ##

The Agent uses sub-agents for all interaction with guests including
creation and deletion.  Multiple sub-agents are generally used.

See (../cmd/mistify-test-rpc-service/)[mistify-test-rpc-service] for
an example sub-agent.

