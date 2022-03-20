package plugin

import (
	"context"
	"io/ioutil"
	"os/exec"
	"path"
	"sync"

	"github.com/plugin-ops/pedestal/pedestal/action"
	"github.com/plugin-ops/pedestal/pedestal/plugin/proto"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

var (
	plugins     = []*plugin.Client{}
	pluginMutex = &sync.Mutex{}
)

func SetPlugin(p *plugin.Client) {
	pluginMutex.Lock()
	plugins = append(plugins, p)
	pluginMutex.Unlock()
}

func CleanAllPlugin() {
	pluginMutex.Lock()
	for _, client := range plugins {
		client.Kill()
	}
	plugins = []*plugin.Client{}
	pluginMutex.Unlock()
}

func AddPlugin(path string) error {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: handshakeConfig,
		Plugins: map[string]plugin.Plugin{
			PluginName: &actionImpl{},
		},
		Cmd:              exec.Command(path),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})
	grpcClient, err := client.Client()
	if err != nil {
		return err
	}
	raw, err := grpcClient.Dispense(PluginName)
	// srv can only be proto.DriverClient
	a := raw.(proto.DriverClient)

	action.RegisterAction(&driverGRPCClient{impl: a})
	SetPlugin(client)

	return nil
}

func AddPluginWithDir(dir string) error {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, f := range fs {
		err = AddPlugin(path.Join(dir, f.Name()))
		if err != nil {
			return err
		}
	}
	return err
}

func ReLoadPluginWithDir(dir string) error {
	action.CleanAllAction()
	return AddPluginWithDir(dir)
}

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

var pluginMap = map[string]plugin.Plugin{
	PluginName: actionImpl{},
}

type actionImpl struct {
	plugin.NetRPCUnsupportedPlugin
	Srv *driverGRPCServer
}

func (a *actionImpl) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	proto.RegisterDriverServer(server, a.Srv)
	return nil
}

func (a *actionImpl) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	return proto.NewDriverClient(conn), nil
}

// driverPlugin use for hide gRPC detail.
type driverGRPCServer struct {
	impl action.Action
}

func (d *driverGRPCServer) Name(ctx context.Context, empty *proto.Empty) (*proto.String, error) {
	name := d.impl.Name()
	return &proto.String{Value: name}, nil
}

func (d *driverGRPCServer) Version(ctx context.Context, empty *proto.Empty) (*proto.Float32, error) {
	version := d.impl.Version()
	return &proto.Float32{Value: version}, nil
}

func (d *driverGRPCServer) Do(ctx context.Context, input *proto.DoInput) (*proto.DoOutput, error) {
	in := make([]interface{}, len(input.GetValue()))
	for i, v := range input.GetValue() {
		in[i] = v
	}
	resp, err := d.impl.Do(action.ConvertSliceToValueSlice(in)...)
	if err != nil {
		return &proto.DoOutput{}, err
	}
	out := make([]string, len(resp))
	for i, value := range resp {
		out[i] = value.String()
	}

	return &proto.DoOutput{Value: out}, nil
}

func (d *driverGRPCServer) Description(ctx context.Context, empty *proto.Empty) (*proto.String, error) {
	desc := d.impl.Description()
	return &proto.String{Value: desc}, nil
}

const (
	StrPluginAbnormal = "abnormal plugin"
	PluginName        = "action"
)

type driverGRPCClient struct {
	impl proto.DriverClient
}

func (d *driverGRPCClient) Name() string {
	name, err := d.impl.Name(context.TODO(), &proto.Empty{})
	if err != nil {
		return StrPluginAbnormal
	}
	return name.GetValue()
}

func (d *driverGRPCClient) Do(params ...*action.Value) ([]*action.Value, error) {
	in := make([]string, len(params))
	for i, v := range params {
		in[i] = v.String()
	}
	resp, err := d.impl.Do(context.TODO(), &proto.DoInput{Value: in})
	if err != nil {
		return nil, err
	}
	out := make([]*action.Value, len(resp.GetValue()))
	for i, value := range resp.GetValue() {
		out[i] = action.NewValue(value)
	}

	return out, nil
}

func (d *driverGRPCClient) Version() float32 {
	version, err := d.impl.Version(context.TODO(), &proto.Empty{})
	if err != nil {
		return -1
	}
	return version.GetValue()
}

func (d *driverGRPCClient) Description() string {
	desc, err := d.impl.Description(context.TODO(), &proto.Empty{})
	if err != nil {
		return StrPluginAbnormal
	}
	return desc.GetValue()
}
