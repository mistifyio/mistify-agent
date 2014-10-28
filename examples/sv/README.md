Example runit scripts and config.

To install (assuming binaries are in /opt/mistify/sbin):

```
install -d -m 0644 agent.json /etc/mistify-agent/agent.json
install -d -m 0755 run /etc/sv/mistify-agent/run
install -d -m 0755 log /etc/sv/mistify-agent/log/run
ln -sf /etc/sv/mistify-agent /etc/service
```
