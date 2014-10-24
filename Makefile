cmd/mistify-agent/mistify-agent: cmd/mistify-agent/main.go
	cd cmd/mistify-agent
	go build


clean:
	cd cmd/mistify-agent
	go clean

install: cmd/mistify-agent/mistify-agent
	install -d cmd/mistify-agent/mistify-agent $(DESTDIR)/usr/local/bin/mistify-agent

