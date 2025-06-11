package multipass

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
	"titan-vm/vms/pb"

	multipassPb "titan-vm/vms/virt/multipass/pb"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	transport = "raw"
	vmapi     = "multipass"
)

type Multipass struct {
	serverURL    string
	certProvider CertProvider
	clients      sync.Map
}

type CertProvider struct {
	CertFile string
	KeyFile  string
	// CAFile   string
}

type RpcClient struct {
	client        multipassPb.RpcClient
	conn          *grpc.ClientConn
	websocketConn *websocket.Conn
}

func generateJwtToken(secret string, expire int64) (string, error) {
	claims := jwt.MapClaims{
		"user": "golibvirt",
		"exp":  time.Now().Add(time.Second * time.Duration(expire)).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))

}

func NewMultipass(serverURL string, certProvider CertProvider) *Multipass {
	return &Multipass{serverURL: serverURL, certProvider: certProvider}
}

func (m *Multipass) connectHost(hostID string) (*RpcClient, error) {
	v, ok := m.clients.Load(hostID)
	if ok {
		return v.(*RpcClient), nil
	}
	return m.newRpcClient(hostID)
}

func (m *Multipass) newRpcClient(hostID string) (*RpcClient, error) {
	url := fmt.Sprintf("%s?uuid=%s&transport=%s&vmapi=%s", m.serverURL, hostID, transport, vmapi)

	websocketConn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("websocket dial failed:%s", err.Error())
	}

	cert, err := tls.LoadX509KeyPair(m.certProvider.CertFile, m.certProvider.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("LoadX509KeyPair error: %s", err.Error())
	}

	caPool := x509.NewCertPool()
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caPool,
		ServerName:         "localhost",
		InsecureSkipVerify: true,
	}
	// conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient("localhost", grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)), grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
		return websocketConn.NetConn(), nil
	}))

	if err != nil {
		return nil, fmt.Errorf("new grpc client failed: %s", err.Error())
	}

	client := multipassPb.NewRpcClient(conn)
	return &RpcClient{client: client, conn: conn, websocketConn: websocketConn}, nil
}

func (m *Multipass) CreateVMWithMultipass(ctx context.Context, request *pb.CreateVMWithMultipassRequest, progressChan chan<- *multipassPb.LaunchProgress) error {
	defer close(progressChan)

	client, err := m.connectHost(request.GetId())
	if err != nil {
		return err
	}

	defer client.conn.Close()

	rpcClient := client.client
	stream, err := rpcClient.Launch(ctx)
	if err != nil {
		return err
	}

	lancheRequest := multipassPb.LaunchRequest{
		InstanceName: request.VmName,
		Image:        request.Image,
		NumCores:     request.Cpu,
		MemSize:      request.Memory,
		DiskSpace:    request.DiskSize,
	}
	err = stream.Send(&lancheRequest)
	if err != nil {
		return err
	}

	for {
		reply, err := stream.Recv()
		if err != nil {
			fmt.Printf("Multipass.CreateVMWithMultipass recv:%s\n", err.Error())
			if err != io.EOF {
				return err
			}
			break
		}

		progress := reply.GetLaunchProgress()
		progressChan <- progress
	}

	return nil
}

func (m *Multipass) CreateVMWithLibvirt(_ context.Context, request *pb.CreateVMWithLibvirtRequest) error {
	return nil
}

func (m *Multipass) StartVM(ctx context.Context, request *pb.StartVMRequest) error {
	client, err := m.connectHost(request.Id)
	if err != nil {
		return err
	}

	defer client.conn.Close()
	// todo: find vm if exist

	rpcClient := client.client
	stream, err := rpcClient.Start(ctx)
	if err != nil {
		return err
	}

	names := multipassPb.InstanceNames{InstanceName: []string{request.VmName}}
	err = stream.Send(&multipassPb.StartRequest{InstanceNames: &names})
	if err != nil {
		return err
	}

	reply, err := stream.Recv()
	if err != nil {
		fmt.Printf("Multipass.StartVM recv:%s\n", err.Error())
		return err
	}

	logx.Infof("start vm %s %s", request.VmName, reply.GetReplyMessage())

	return nil
}

func (m *Multipass) StopVM(ctx context.Context, request *pb.StopVMRequest) error {
	client, err := m.connectHost(request.Id)
	if err != nil {
		return err
	}
	defer client.conn.Close()
	// todo: find vm if exist

	rpcClient := client.client
	stream, err := rpcClient.Stop(ctx)
	if err != nil {
		return err
	}

	names := multipassPb.InstanceNames{InstanceName: []string{request.VmName}}
	err = stream.Send(&multipassPb.StopRequest{InstanceNames: &names})
	if err != nil {
		return err
	}

	_, err = stream.Recv()
	if err != nil {
		fmt.Printf("Multipass.StartVM StopVM:%s\n", err.Error())
		if err != io.EOF {
			return err
		}
	}
	// logx.Infof("start vm %s %s", request.VmName, reply.GetReplyMessage())

	return nil
}

func (m *Multipass) DeleteVM(ctx context.Context, request *pb.DeleteVMRequest) error {
	client, err := m.connectHost(request.Id)
	if err != nil {
		return err
	}
	defer client.conn.Close()
	// todo: find vm if exist

	rpcClient := client.client
	stream, err := rpcClient.Delet(ctx)
	if err != nil {
		return err
	}

	instanceSnapshotPair := multipassPb.InstanceSnapshotPair{InstanceName: request.VmName}
	err = stream.Send(&multipassPb.DeleteRequest{InstanceSnapshotPairs: []*multipassPb.InstanceSnapshotPair{&instanceSnapshotPair}})
	if err != nil {
		return err
	}

	_, err = stream.Recv()
	if err != nil {
		return err
	}

	purgeStream, err := rpcClient.Purge(ctx)
	if err != nil {
		return err
	}

	err = purgeStream.Send(&multipassPb.PurgeRequest{})
	if err != nil {
		return err
	}

	_, err = purgeStream.Recv()
	if err != nil {
		return err
	}

	return nil
}

func (m *Multipass) UpdateVM(ctx context.Context, request *pb.UpdateVMRequest) error {
	return nil
}

func (m *Multipass) ListVMInstance(ctx context.Context, request *pb.ListVMInstanceReqeust) (*pb.ListVMInstanceResponse, error) {
	client, err := m.connectHost(request.Id)
	if err != nil {
		return nil, err
	}
	defer client.conn.Close()
	// todo: find vm if exist
	fmt.Printf("Multipass.ListVMInstance new Connect to host %s\n", request.Id)

	rpcClient := client.client
	stream, err := rpcClient.List(ctx)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Multipass.ListVMInstance request list\n")

	err = stream.Send(&multipassPb.ListRequest{})
	if err != nil {
		return nil, err
	}

	fmt.Printf("Multipass.ListVMInstance send request\n")

	reply, err := stream.Recv()
	if err != nil {
		return nil, err
	}

	fmt.Printf("Multipass.ListVMInstance recv\n")

	instanceList := reply.GetInstanceList()
	vmInfos := make([]*pb.VMInfo, 0, len(instanceList.Instances))

	for _, instance := range instanceList.Instances {
		// logx.Infof("delete vm %s", instance)
		vmInfo := &pb.VMInfo{Name: instance.GetName(), State: instance.GetInstanceStatus().String()}
		if len(instance.Ipv4) > 0 {
			vmInfo.Ip = instance.Ipv4[0]
		}

		vmInfos = append(vmInfos, vmInfo)
	}
	return &pb.ListVMInstanceResponse{VmInfos: vmInfos}, nil
}

func (m *Multipass) ListImage(_ context.Context, request *pb.ListImageRequest) (*pb.ListImageResponse, error) {
	return nil, nil
}

func (m *Multipass) DeleteImage(_ context.Context, request *pb.DeleteImageRequest) error {
	return nil
}

func (m *Multipass) CreateVolWithLibvirt(ctx context.Context, in *pb.CreateVolWithLibvirtReqeust) (*pb.CreateVolWithLibvirtResponse, error) {
	return nil, fmt.Errorf("not implement")
}
func (m *Multipass) GetVol(ctx context.Context, in *pb.GetVolRequest) (*pb.GetVolResponse, error) {
	return nil, fmt.Errorf("not implement")
}

//	func (m *Multipass) ListHostNetworkInterfaceWithLibvirt(ctx context.Context, in *pb.ListHostNetworkInterfaceRequest) (*pb.ListHostNetworkInterfaceResponse, error) {
//		return nil, nil
//	}
//
//	func (m *Multipass) ListVMNetwrokInterfaceWithLibvirt(ctx context.Context, in *pb.ListVMNetwrokInterfaceReqeust) (*pb.ListVMNetworkInterfaceResponse, error) {
//		return nil, nil
//	}
func (m *Multipass) AddNetworkInterfaceWithLibvirt(ctx context.Context, in *pb.AddNetworkInterfaceRequest) error {
	return nil
}
func (m *Multipass) DeleteNetworkInterfaceWithLibvirt(ctx context.Context, in *pb.DeleteNetworkInterfaceRequest) error {
	return nil
}

func (m *Multipass) AddHostdevWithLibvirt(ctx context.Context, in *pb.AddHostdevRequest) error {
	return nil
}
func (m *Multipass) DeleteHostdevWithLibvirt(ctx context.Context, in *pb.DeleteHostdevRequest) error {
	return nil
}

func (m *Multipass) GetVncPortWithLibvirt(ctx context.Context, in *pb.VMVncPortRequest) (*pb.VMVncPortResponse, error) {
	return nil, nil
}

//	func (m *Multipass) ListHostDiskWithLibvirt(ctx context.Context, in *pb.ListHostDiskRequest) (*pb.ListDiskResponse, error) {
//		return nil, nil
//	}
//
//	func (m *Multipass) ListVMDiskWithLibvirt(ctx context.Context, in *pb.ListVMDiskRequest) (*pb.ListVMDiskResponse, error) {
//		return nil, nil
//	}
func (m *Multipass) GetVMInfo(ctx context.Context, in *pb.GetVMInfoRequest) (*pb.GetVMInfoResponse, error) {
	return nil, nil
}
func (m *Multipass) AddDiskWithLibvirt(ctx context.Context, in *pb.AddDiskRequest) error {
	return nil
}
func (m *Multipass) DeleteDiskWithLibvirt(ctx context.Context, in *pb.DeleteDiskRequest) error {
	return nil
}
