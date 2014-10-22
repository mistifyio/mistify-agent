Mistify Agent
==============

## Overview ##

Mistify Agent is an agent for managing a hypervisor. It runs local on the hypervisor exposes a simple HTTP API for creating/deleting/modifying virtual machines.

## Status ##

This is still very early in the development process.

## General Architecture ##

Mistify Agent is really just a "core" agent that provides a simple
REST-ish API for manipulating a hypervisor. The "core" agent defines
the endpoints and the general "actions" that can be done, but the
majority of the "work" is done by sub-agents.  Communication between
the agent and sub-agents is done with JSON-RPC over HTTP.  These
sub-agents are assumed to be running on the same host as the agent.

The data structures for the REST API are defined in
http://godoc.org/github.com/mistifyio/mistify-agent/client

The data structures for the RPC API are defined in http://godoc.org/github.com/mistifyio/mistify-agent/rpc

Sub-agents do not have to be written in Go, but the agent does provide
helpers for easily creating them. All of the "official" sub-agents are
(or will be) written in Go.

There are two types of sub-agents:
- images/storage - Mistify will provide an opinionated version that
uses ZFS ZVOLs for guest disk
- guest "actions" - create/delete/reboot/etc

Only on storage sub-agent is used, but multiple "guest" sub-agents may
be used.  The Agent uses "pipelines" for guest actions. A pipeline is
a series of calls to sub-agents.  The pipeline continues until it
reaches the end or a sub-agent returns an error.

Each action
can have two pipelines (and must have at least one):
- sync - these  pipelines are called synchronously at request time.
  The agent does not return a result to the requesting client until
  the pipeline is done.  This can be used for validations or other
  work that needs to be done at initial request time.
- async - these pipelines are called "in the background" and will
be ran periodically, even if an error is returned.

An example is on guest create, the sync pipeline could try to reserve
resources. Then the async pipeline could actually allocate disks,
memory, etc.

The Agent will move the guest to the running state after the create
action is done.  The agent also persists the guest data at each stage
of each pipeline.

# Contributing #

See the [contributing guidelines](./CONTRIBUTING.md)


