#/usr/bin/env bash
set -e
set -o pipefail
set -x

trap "kill 0" SIGINT SIGTERM EXIT


pushd examples/test-rpc-service
go get
go clean
go build
./test-rpc-service &
RPC_PID=$!
trap "kill $RPC_PID" SIGINT SIGTERM EXIT
popd

pushd cmd/mistify-agent
go get
go clean
go build
./mistify-agent --config-file ../../examples/test-rpc-service/agent.json &
AGENT_PID=$!
trap "kill $RPC_PID; kill $AGENT_PID" SIGINT SIGTERM EXIT
popd


sleep 5

HOST=${1:-127.0.0.1}
PORT=${2:-8080}

http (){
        METHOD=$1
        URL=$2
        shift 2
        curl --fail -sv -X $METHOD -H 'Content-Type: application/json' http://$HOST:$PORT/$URL "$@" | jq .
}

http PATCH metadata --data-binary '{"foo": "bar", "hello": "world" }'

http PATCH metadata --data-binary '{"foo": null}'

http GET guests

ID=$(http POST guests --data-binary '{"metadata": { "foo": "bar"}, "memory": 512, "cpu": 2, "nics": [ { "model": "virtio", "address": "10.10.10.10", "netmask": "255.255.255.0", "gateway": "10.10.10.1", "network": "br0"} ], "disks": [ {"image": "ubuntu-14.04-server-mistify-amd64-disk1.zfs"}, {"size": 512} ] }' | jq -r .id)


for m in cpu disk nic; do
    http GET guests/$ID/metrics/$m
done

kill $AGENT_PID
sleep 1
kill $RPC_PID

trap - SIGINT SIGTERM EXIT

exit 0
