package deviceplugin

import (
	"net"
	"time"
	"log"
	"context"
	"path"
	"os"
	"sync"

	"google.golang.org/grpc"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

type AllocateFunc func([]string) (*pluginapi.ContainerAllocateResponse, error)
type PreStartFunc func([]string) error

type generalDevicePlugin struct {
	lock sync.Mutex

	resourceName string
	socket string

	devs   map[string]*pluginapi.Device

	stop   chan struct{}
	update chan []*pluginapi.Device

	server *grpc.Server

	preStartFunc PreStartFunc
	allocateFunc AllocateFunc
}

func New(resourceName, socket string) DevicePlugin {
	return &generalDevicePlugin{
		resourceName: resourceName,
		socket:       socket,
		devs:         map[string]*pluginapi.Device{},
		stop:         make(chan struct{}),
		update:       make(chan []*pluginapi.Device),
	}
}

func (p *generalDevicePlugin) SetPreStartFunc(f PreStartFunc) {
	p.preStartFunc = f
}

func (p *generalDevicePlugin) SetAllocateFunc(f AllocateFunc) {
	p.allocateFunc = f
}

func (p *generalDevicePlugin) Start() error {
	err := p.startServer()
	if err != nil {
		return err
	}
	log.Println("Starting to serve on", p.socket)

	err = p.register()
	if err != nil {
		log.Printf("Could not register device plugin: %s", err)
		p.Stop()
		return err
	}

	log.Println("Registered device plugin with Kubelet")
	return nil
}

func (p *generalDevicePlugin) Stop() error {
	p.server.Stop()
	close(p.stop)
	return p.cleanup()
}

func (p *generalDevicePlugin) AddOrUpdateDevice(devs ...*pluginapi.Device) {
	p.lock.Lock()
	defer p.lock.Unlock()

	updated := false
	for _, dev := range devs {
		if d, has := p.devs[dev.ID]; !has {
			log.Println("Add device", dev)
			p.devs[dev.ID] = dev
			updated = true
		} else {
			if d.Health != dev.Health {
				log.Println("Update device", dev)
				d.Health = dev.Health
				updated = true
			}
		}
	}

	if updated {
		p.update <- p.listDeviceLocked()
	}
}

func (p *generalDevicePlugin) RemoveDevice(devs ...*pluginapi.Device) {
	p.lock.Lock()
	defer p.lock.Unlock()

	updated := false
	for _, dev := range devs {
		if _, has := p.devs[dev.ID]; has {
			log.Println("Delete device", dev.ID)
			delete(p.devs, dev.ID)
			updated = true
		}
	}

	if updated {
		p.update <- p.listDeviceLocked()
	}
}

func (p *generalDevicePlugin) ReplaceDevice(devs ...*pluginapi.Device) {
	p.lock.Lock()
	defer p.lock.Unlock()

	log.Println("Replace devices with", devs)
	p.devs = nil
	for _, dev := range devs {
		p.devs[dev.ID] = dev
	}

	p.update <- p.listDeviceLocked()
}

func (p *generalDevicePlugin) listDeviceLocked() []*pluginapi.Device {
	devs := make([]*pluginapi.Device, len(p.devs))
	for _, d := range p.devs {
		devs = append(devs, d)
	}
	return devs
}

func (p *generalDevicePlugin) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{PreStartRequired: p.preStartRequired()}, nil
}

func (p *generalDevicePlugin) Allocate(_ context.Context, r *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	resp := &pluginapi.AllocateResponse{}
	if p.allocateFunc == nil {
		return resp, nil
	}

	for _, creq := range r.ContainerRequests {
		cresp, err := p.allocateFunc(creq.DevicesIDs)
		if err != nil {
			return &pluginapi.AllocateResponse{}, err
		}
		resp.ContainerResponses = append(resp.ContainerResponses, cresp)
	}
	return resp, nil
}

func (p *generalDevicePlugin) ListAndWatch(_ *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	log.Println("ListAndWatch")
	for {
		select {
		case <- p.stop:
			return nil
		case updated := <- p.update:
			s.Send(&pluginapi.ListAndWatchResponse{Devices: updated})
		}
	}
}

func (p *generalDevicePlugin) PreStartContainer(_ context.Context, r *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	resp := &pluginapi.PreStartContainerResponse{}
	if p.preStartFunc == nil {
		return resp, nil
	}

	err := p.preStartFunc(r.DevicesIDs)
	return resp, err
}

func (p *generalDevicePlugin) startServer() error {
	err := p.cleanup()
	if err != nil {
		return err
	}

	sock, err := net.Listen("unix", p.socket)
	if err != nil {
		return err
	}

	p.server = grpc.NewServer([]grpc.ServerOption{}...)
	pluginapi.RegisterDevicePluginServer(p.server, p)

	go p.server.Serve(sock)

	// Wait for server to start by launching a blocking connexion
	conn, err := dial(p.socket)
	if err != nil {
		return err
	}
	conn.Close()

	log.Println("Starting to serve on", p.socket)
	return nil
}

func (p *generalDevicePlugin) register() error {
	conn, err := dial(pluginapi.KubeletSocket)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	reqt := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(p.socket),
		ResourceName: p.resourceName,
		Options:      &pluginapi.DevicePluginOptions{PreStartRequired: p.preStartRequired()},
	}

	_, err = client.Register(context.Background(), reqt)
	return err
}

func (p *generalDevicePlugin) preStartRequired() bool {
	return p.preStartFunc != nil
}

// dial establishes the gRPC communication with the registered device plugin.
func dial(unixSocketPath string) (*grpc.ClientConn, error) {
	return grpc.Dial(unixSocketPath, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithTimeout(10*time.Second),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)
}

func (p *generalDevicePlugin) cleanup() error {
	if err := os.Remove(p.socket); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
