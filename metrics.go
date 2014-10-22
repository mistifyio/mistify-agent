package agent

import (
	"github.com/mistifyio/mistify-agent/client"
	"github.com/mistifyio/mistify-agent/rpc"
)

func (ctx *Context) getMetrics(guest *client.Guest, mtype string) (*rpc.GuestMetricsResponse, error) {
	handler := ctx.Metrics[mtype]

	req := rpc.GuestMetricsRequest{
		Guest: guest,
		Args:  handler.Args,
		Type:  mtype,
	}
	var resp rpc.GuestMetricsResponse

	err := handler.Service.Client.Do(handler.Method, &req, &resp)

	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (ctx *Context) getCpuMetrics(guest *client.Guest) ([]*client.GuestCpuMetrics, error) {
	resp, err := ctx.getMetrics(guest, "cpu")

	if err != nil {
		return nil, err
	}

	return resp.CPU, nil
}

func (ctx *Context) getNicMetrics(guest *client.Guest) (map[string]*client.GuestNicMetrics, error) {

	resp, err := ctx.getMetrics(guest, "nic")

	if err != nil {
		return nil, err
	}

	return resp.Nic, nil
}

func (ctx *Context) getDiskMetrics(guest *client.Guest) (map[string]*client.GuestDiskMetrics, error) {

	resp, err := ctx.getMetrics(guest, "disk")

	if err != nil {
		return nil, err
	}

	return resp.Disk, nil
}
