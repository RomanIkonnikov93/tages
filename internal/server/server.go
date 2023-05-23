package server

import (
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/RomanIkonnikov93/tages/internal/config"
	"github.com/RomanIkonnikov93/tages/internal/grpcapi"
	pb "github.com/RomanIkonnikov93/tages/internal/proto"
	"github.com/RomanIkonnikov93/tages/pkg/pkg/logging"

	"google.golang.org/grpc"
)

func StartServer(service *grpcapi.KeeperServiceServer, cfg *config.Config, logger *logging.Logger) error {

	listen, err := net.Listen("tcp", cfg.GRPCAddress)
	if err != nil {
		logger.Fatal("net.Listen: ", err)
	}

	s := grpc.NewServer()
	pb.RegisterKeeperServer(s, service)

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		<-sigint
		logger.Println("server shutdown gracefully")
		s.GracefulStop()
		close(service.DownloadUploadChannel)
		select {
		case <-service.ShutdownChannel:
			break
		}
		wg.Done()
	}()

	logger.Info("gRPC server running")
	if err = s.Serve(listen); err != nil {
		logger.Fatal(err)
	}
	wg.Wait()

	return nil
}
