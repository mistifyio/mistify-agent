PREFIX := /opt/mistify
SBIN_DIR=$(PREFIX)/sbin

cmd/mistify-agent/mistify-agent: cmd/mistify-agent/main.go
	cd cmd/mistify-agent && \
	go get && \
	go build


clean:
	cd cmd/mistify-agent && \
	go clean

install: cmd/mistify-agent/mistify-agent
	install -D cmd/mistify-agent/mistify-agent $(DESTDIR)$(SBIN_DIR)/mistify-agent

