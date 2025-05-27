package multipass

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"testing"
	"titan-vm/vms/virt/multipass/pb"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// func main() {
// 	// log.Println("hello world")
// 	testMultipass()
// }

func TestMultipassListInstance(t *testing.T) {
	serverAddr := "ws://localhost:7777/vm?uuid=98ffba3e-2bf6-11f0-a385-67a4bcb29f5d&transport=raw&vmapi=multipass"
	websocketConn, _, err := websocket.DefaultDialer.Dial(serverAddr, nil)
	if err != nil {
		log.Fatal("Dial failed:", err)
	}

	provider := CertProvider{CertFile: "../../etc/ec_cert.pem", KeyFile: "../../etc/ec_key.pem"}
	cert, err := tls.LoadX509KeyPair(provider.CertFile, provider.KeyFile)
	if err != nil {
		log.Fatal(err)
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
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewRpcClient(conn)

	// 调用 Multipass 服务的 list 方法
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := client.List(ctx)
	if err != nil {
		log.Fatalf("list: %v", err)
	}

	err = stream.Send(&pb.ListRequest{})
	if err != nil {
		log.Fatalf("send: %v", err)
	}

	reply, err := stream.Recv()
	if err != nil {
		log.Fatalf("Recv: %v", err)
	}

	instanceList := reply.GetInstanceList()
	// vmInfos := make([]*pb.VMInfo, 0, len(instanceList.Instances))

	for _, instance := range instanceList.Instances {
		fmt.Printf("instance name: %v\n", instance.GetName())
	}

}

func TestMultipassCreateVM(t *testing.T) {
	serverAddr := "ws://localhost:7777/vm?uuid=98ffba3e-2bf6-11f0-a385-67a4bcb29f5d&transport=raw&vmapi=multipass"
	websocketConn, _, err := websocket.DefaultDialer.Dial(serverAddr, nil)
	if err != nil {
		log.Fatal("Dial failed:", err)
	}

	provider := CertProvider{CertFile: "../../etc/ec_cert.pem", KeyFile: "../../etc/ec_key.pem"}
	cert, err := tls.LoadX509KeyPair(provider.CertFile, provider.KeyFile)
	if err != nil {
		log.Fatal(err)
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
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewRpcClient(conn)

	// 调用 Multipass 服务的 list 方法
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := client.Launch(ctx)
	if err != nil {
		log.Fatalf("list: %v", err)
	}

	launchRequest := &pb.LaunchRequest{
		InstanceName: "test",
		NumCores:     1,
		MemSize:      "1G",
		DiskSpace:    "70G",
		Image:        "file:///var/snap/multipass/common/data/multipassd/vault/instances/ubuntu-niulink/ubuntu-24.04-server-cloudimg-amd64.img",
	}

	err = stream.Send(launchRequest)
	if err != nil {
		log.Fatalf("send: %v", err)
	}

	for {
		reply, err := stream.Recv()
		if err != nil {
			log.Fatalf("Recv: %v", err)
		}

		progress := reply.GetLaunchProgress()
		log.Printf("progress %s:%s\n", progress.GetType().String(), progress.GetPercentComplete())

	}

}
