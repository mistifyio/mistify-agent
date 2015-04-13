PREFIX := /opt/mistify
SBIN_DIR=$(PREFIX)/sbin
CONF_DIR := /etc/mistify

cmd/mistify-agent/mistify-agent: cmd/mistify-agent/main.go
	cd cmd/mistify-agent && \
	go get && \
	go build

clean:
	cd cmd/mistify-agent && \
	go clean

install_config:
	install -D -m 0444 -o root cmd/mistify-agent/agent.json $(CONF_DIR)/agent.json

install: cmd/mistify-agent/mistify-agent install_config
	install -D cmd/mistify-agent/mistify-agent $(DESTDIR)$(SBIN_DIR)/mistify-agent

