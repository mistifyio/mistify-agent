PREFIX := /opt/mistify
BASEDIR=$(DESTDIR)$(PREFIX)
SBIN_DIR=$(BASEDIR)/sbin
SV_DIR=$(BASEDIR)/sv
ETC_DIR=$(BASEDIR)/etc

cmd/mistify-agent/mistify-agent: cmd/mistify-agent/main.go
	cd cmd/mistify-agent && \
	go get && \
	go build

clean:
	cd cmd/mistify-agent && \
	go clean

install: cmd/mistify-agent/mistify-agent
	mkdir -p $(SBIN_DIR)
	mkdir -p $(SV_DIR)

	install -D cmd/mistify-agent/mistify-agent $(SBIN_DIR)/mistify-agent
	install -D -m 0755 scripts/sv/run $(SV_DIR)/mistify-agent/run
	install -D -m 0755 scripts/sv/log $(SV_DIR)/mistify-agent/log/run
