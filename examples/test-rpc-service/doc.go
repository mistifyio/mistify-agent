/*
test-rpc-service is a test Mistify sub-agent that implements stubs of all the
actions the agent may call.  It returns the guest received for Guest requests
and fake metrics for metric requests.  A sub-agent should generally only have
one area of concern and do one thing well.  This allows sub-agents to be
composited in various ways.

Run the mistify agent with the test-rpc-service agent.json to use this sub-agent
for all actions.

Usage

	Usage of ./test-rpc-service:
	-p, --port=9999: listen port
*/
package main
