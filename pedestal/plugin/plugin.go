package plugin

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/os/glog"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/plugin-ops/pedestal/pedestal/action"
	"github.com/plugin-ops/pedestal/pedestal/config"
	"github.com/plugin-ops/pedestal/pedestal/log"
	"github.com/plugin-ops/pedestal/pedestal/plugin/proto"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

var (
	plugins     = map[string] /*action name@action version*/ *plugin.Client{}
	pluginMutex = &sync.Mutex{}
)

func LoadPluginWithLocal(stage *log.Stage, path string) error {
	stage = stage.Go("LoadPluginWithLocal")
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
	action.RegisterAction(stage, a)

	{
		pluginMutex.Lock()
		key := action.GenerateActionKey(a)
		oldClient, ok := plugins[key]
		if ok {
			oldClient.Kill()
		}
		plugins[key] = client
		pluginMutex.Unlock()
	}

	glog.Infof(stage.Context(), "load local plugin %v, discover plugins %v@%v\n", path, a.Name(), a.Version())
	return nil
}

func LoadPluginWithLocalDir(stage *log.Stage, dir string) error {
	stage = stage.Go("LoadPluginWithLocalDir")
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, f := range fs {
		err = LoadPluginWithLocal(stage, path.Join(dir, f.Name()))
		if err != nil {
			return err
		}
	}
	return nil
}

func ReLoadPluginWithDir(stage *log.Stage, dir string) error {
	stage = stage.Go("ReLoadPluginWithDir")
	failedPlugin := []string{}
	for s := range plugins {
		nv := strings.Split(s, "@")
		if len(nv) != 2 {
			failedPlugin = append(failedPlugin, s)
			continue
		}
		ff, err := strconv.ParseFloat(nv[1], 10)
		if err != nil {
			failedPlugin = append(failedPlugin, s)
			continue
		}
		action.RemoveAction(stage, nv[0], float32(ff))
	}
	UninstallAllPlugin(stage)
	if len(failedPlugin) > 0 {
		return fmt.Errorf("plugins [%v] uninstall failed, please retry loading all plugin after processing", strings.Join(failedPlugin, ", "))
	}
	return LoadPluginWithLocalDir(stage, dir)
}

func CleanUselessPluginFile(stage *log.Stage) error {
	stage = stage.Go("CleanUselessPluginFile")
	files, err := ioutil.ReadDir(config.PluginDir)
	if err != nil {
		return err
	}

	errStrs := []string{}

	pluginMutex.Lock()
	for _, file := range files {
		has := false
		for f := range plugins {
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
			glog.Infof(stage.Context(), "clean useless plugin file %v in %v", file.Name(), config.PluginDir)
		}
	}
	pluginMutex.Unlock()

	if len(errStrs) > 0 {
		return fmt.Errorf("remove useless plugin file( %v ) faile.", strings.Join(errStrs, ", "))
	}
	return nil
}

func UninstallAllPlugin(stage *log.Stage) {
	stage = stage.Go("UninstallAllPlugin")
	pluginMutex.Lock()
	for _, client := range plugins {
		client.Kill()
	}

	current := []string{}
	for s := range plugins {
		current = append(current, s)
	}
	glog.Warningf(stage.Context(), "uninstall all plugin, current working plugins: %v", strings.Join(current, ";"))
	plugins = map[string]*plugin.Client{}
	pluginMutex.Unlock()
}

func UninstallPlugin(stage *log.Stage, actionName string, actionVersion float32) {
	stage = stage.Go("UninstallPlugin")
	pluginMutex.Lock()
	action.RemoveAction(stage, actionName, actionVersion)
	key := fmt.Sprintf("%v@%v", actionName, actionVersion)
	glog.Infof(stage.Context(), "uninstall the plugin for action %v\n", key)
	plugins[key].Kill()
	delete(plugins, key)
	pluginMutex.Unlock()
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
	out := map[string]string{}
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

func (d *driverGRPCClient) Do(params ...*action.Value) (map[string]*action.Value, error) {
	in := make([]string, len(params))
	for i, v := range params {
		in[i] = v.String()
	}
	resp, err := d.impl.Do(context.TODO(), &proto.DoInput{Value: in})
	if err != nil {
		return nil, err
	}
	out := map[string]*action.Value{}
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
