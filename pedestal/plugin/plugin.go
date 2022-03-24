package plugin

import (
	"context"
	"fmt"
	"github.com/plugin-ops/pedestal/pedestal/config"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"

	"github.com/plugin-ops/pedestal/pedestal/action"
	"github.com/plugin-ops/pedestal/pedestal/plugin/proto"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

var (
	plugins     = map[string] /*action name*/ *plugin.Client{}
	pluginFile  = map[string] /*action name*/ string /*file name*/ {}
	pluginMutex = &sync.Mutex{}
)

func SetPlugin(actionName, pluginPath string, p *plugin.Client) {
	pluginMutex.Lock()
	a, ok := plugins[actionName]
	if ok {
		a.Kill()
	}
	plugins[actionName] = p
	pluginFile[actionName] = pluginPath
	pluginMutex.Unlock()
}

func CleanAllPlugin() {
	pluginMutex.Lock()
	for _, client := range plugins {
		client.Kill()
	}
	plugins = map[string]*plugin.Client{}
	pluginFile = map[string]string{}
	pluginMutex.Unlock()
}

func RemovePlugin(actionName string) {
	pluginMutex.Lock()
	action.RemoveAction(actionName)
	plugins[actionName].Kill()
	delete(plugins, actionName)
	delete(pluginFile, actionName)
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
	c := raw.(proto.DriverClient)

	a := &driverGRPCClient{impl: c}
	action.RegisterAction(a)
	SetPlugin(a.Name(), path, client)

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

func CleanUselessPluginFile() error {
	files, err := ioutil.ReadDir(config.PluginDir)
	if err != nil {
		return err
	}

	errStrs := []string{}

	pluginMutex.Lock()
	for _, file := range files {
		has := false
		for _, f := range pluginFile {
			if path.Join(config.PluginDir, file.Name()) == f {
				has = true
				break
			}
		}
		if !has {
			err := os.RemoveAll(path.Join(config.PluginDir, file.Name()))
			if err != nil {
				errStrs = append(errStrs, err.Error())
			}
		}
	}
	pluginMutex.Unlock()

	if len(errStrs) > 0 {
		return fmt.Errorf("remove useless plugin file( %v ) faile.", strings.Join(errStrs, ", "))
	}
	return nil
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
	impl    proto.DriverClient
	name    string
	version float32
	desc    string
}

func (d *driverGRPCClient) Name() string {
	if d.name == "" {
		name, err := d.impl.Name(context.TODO(), &proto.Empty{})
		if err != nil {
			return StrPluginAbnormal
		}
		d.name = name.GetValue()
	}
	return d.name
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
	if d.version == 0 {
		version, err := d.impl.Version(context.TODO(), &proto.Empty{})
		if err != nil {
			return -1
		}
		d.version = version.GetValue()
	}
	return d.version
}

func (d *driverGRPCClient) Description() string {
	if d.desc == "" {
		desc, err := d.impl.Description(context.TODO(), &proto.Empty{})
		if err != nil {
			return StrPluginAbnormal
		}
		d.desc = desc.GetValue()
	}
	return d.desc
}
