package faker

import (
	"context"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/samber/lo"
	"github.com/seth-shi/go-zero-testing-example/app/id/rpc/id"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type idServer struct {
	id.UnimplementedIdServer
	id     uint64
	locker sync.Locker
	conn   *bufconn.Listener
	srv    *grpc.Server
}

func (r *idServer) Get(ctx context.Context, request *id.IdRequest) (*id.IdResponse, error) {
	r.locker.Lock()
	defer r.locker.Unlock()

	r.id++

	return &id.IdResponse{
		Id:   r.id,
		Node: 1,
	}, nil
}

const (
	buffSize = 1024 * 1024
)

func newIdServer() *idServer {
	service := &idServer{
		conn:   bufconn.Listen(buffSize),
		srv:    grpc.NewServer(),
		id:     uint64(rand.Int() + startId),
		locker: &sync.RWMutex{},
	}
	id.RegisterIdServer(service.srv, service)
	go service.Serve()
	time.Sleep(time.Second)

	return service
}

func (r *idServer) Serve() {
	lo.Must0(r.srv.Serve(r.conn))

}

func (r *idServer) Connect() *grpc.ClientConn {

	// Create the dialer
	dialer := func(context.Context, string) (net.Conn, error) {
		return r.conn.Dial()
	}

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithContextDialer(dialer))
	// Disable transport security
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	// Connect https://stackoverflow.com/questions/78485578/how-to-use-the-bufconn-package-with-grpc-newclient
	return lo.Must(grpc.NewClient("passthrough://bufnet", opts...))
}
