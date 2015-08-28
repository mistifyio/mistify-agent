/*
Package agent is the core Mistify Agent for managing a hypervisor. It runs local on the
hypervisor and exposes an HTTP API for managing virtual machines.

General Architecture

Mistify Agent  provides a REST-ish API for manipulating a hypervisor and
defines a general set of actions. It is primarilly an interface to the
sub-agents, which do most of the work.  Sub-agents are assumed to be running on
the same host as the agent and communication is done via JSON-RPC over HTTP.

The data structures for the REST API are defined in:
http://godoc.org/github.com/mistifyio/mistify-agent/client

The data structures for the RPC API are defined in:
http://godoc.org/github.com/mistifyio/mistify-agent/rpc

Sub-agents do not have to be written in Go, but the agent does provide helpers
for easily creating them. All of the official sub-agents are written in Go.

There are two types of sub-agents:

* images/storage - Mistify will provide an opinionated version that uses ZFS
ZVOLs for guest disks.

* guest "actions" - create/delete/reboot/etc.

Only one storage sub-agent is used, but multiple guest sub-agents may be used.

Actions

While there is a defined set of actions, the work performed is configurable.
Each action has a pipeline, a series of one or more steps that need to be
performed to complete the action, configurable in the config file. All steps
must succeed, in order, for an action to be considered successful.

There are three action types:

* Info - Information retrieval actions, such as getting a list of guests,
called synchronously at request time. A JSON result is returned to the
requesting client when the pipeline is complete.

* Async - Modification actions, such as rebooting a guest, called
asynchronously. One action per guest is performed at a time, while the rest are
queued. A response containing the job id in the header X-Guest-Job-ID is
returned after queueing the action, which can be used to check the status at a
later time.

* Stream - Data retrieval, such as downloading a zfs snapshot, called
synchronously at request time. Rather than a JSON response, data is streamed
back in chunks.

Valid actions are defined in:
http://godoc.org/github.com/mistifyio/mistify-agent/config

HTTP API Endpoints

	/debug/pprof
		* GET

	/debug/pprof/{profileName}
		* GET

	/debug/pprof/cmdline
		* GET

	/debug/pprof/symbol
		* GET

	/debug/pprof/symbol
		* GET

	/metadata
		* GET   - Retrieve the hypervisor's metadata
		* PATCH - Modify the hypervisor's metadata

	/images
		* GET  - Retrieve a list of disk images
		* POST - Fetch a disk image

	/images/{imageID}
		* GET    - Retrieve information about a disk image
		* DELETE - Delete a disk image

	/container_images
		* GET  - Retrieve a list of container images
		* POST - Fetch a container image

	/container_images/{imageID}
		* GET    -  Retrieve information about a container image
		* DELETE - Delete a container image

	/guests
		* GET  - Retrieve a list of guests
		* POST - Create a new guest

	/guests/{guestID}
		* GET - Retrieve information about a guest

	/guests/{guestID}/jobs
		* GET - Retrieve a list of recent action jobs for the guest

	/guests/{guestID}/jobs/{jobID}
		* GET - Retrieve information about a specific action job

	/guests/{guestID}/metadata
		* GET   - Retrieve a guest's metadata
		* PATCH - Modify the guest's metadata

	/guests/{guestID}/metrics/cpu
		* GET - Retrieve guest CPU metrics

	/guests/{guestID}/metrics/disk
		* GET - Retrieve guest disk metrics

	/guests/{guestID}/metrics/nic
		* GET - Retrieve guest NIC metrics

	/guests/{guestID}/{actionName}
		Actions: shutdown, reboot, restart, poweroff, start, suspend, delete
		* GET - Perform the specified action for the guest

	/guests/{guestID}/snapshots
	/guests/{guestID}/disks/{diskID}/snapshots
		* GET  - Retrieve a list of snapshots
		* POST - Create a new snapshot

	/guests/{guestID}/snapshots/{snapshotName}
	/guests/{guestID}/disks/{diskID}/snapshots/{snapshotName}
		* GET    - Retrieve information about a snapshot
		* DELETE - Delete a snapshot

	/guests/{guestID}/snapshots/{snapshotName}/rollback
	/guests/{guestID}/disks/{diskID}/snapshots/{snapshotName}/rollback
		* POST - Roll back to the snapshot

	/guests/{guestID}/snapshots/{snapshotName}/download
	/guests/{guestID}/disks/{diskID}/snapshots/{snapshotName}/download
		* GET - Download the snapshot

Contributing

See the guidelines:
https://github.com/mistifyio/mistify-agent/blob/master/CONTRIBUTING.md
*/
package agent
