package grpcapi

import (
	"context"
	"os"
	"path/filepath"

	"github.com/RomanIkonnikov93/tages/internal/fileos"
	"github.com/RomanIkonnikov93/tages/internal/models"
	pb "github.com/RomanIkonnikov93/tages/internal/proto"
	"github.com/RomanIkonnikov93/tages/pkg/logging"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type KeeperServiceServer struct {
	pb.UnimplementedKeeperServer
	DownloadUploadChannel chan models.Record
	RecordChannel         chan models.Record
	FilesInfoChannel      chan models.Record
	ListChannel           chan []models.Record
	ShutdownChannel       chan struct{}
	logger                *logging.Logger
}

func InitServices(logger *logging.Logger) *KeeperServiceServer {

	return &KeeperServiceServer{
		DownloadUploadChannel: make(chan models.Record, models.UploadDownloadParallelCount),
		RecordChannel:         make(chan models.Record),
		FilesInfoChannel:      make(chan models.Record, models.FilesInfoParallelCount),
		ListChannel:           make(chan []models.Record),
		ShutdownChannel:       make(chan struct{}),
		logger:                logger,
	}
}

func (k *KeeperServiceServer) Run() error {

	if _, err := os.Stat("storage"); os.IsNotExist(err) {
		err = os.Mkdir("storage", 0666)
		if err != nil {
			return err
		}
	}

	for {
		select {
		case r := <-k.DownloadUploadChannel:

			switch r.RequestType {

			case models.Upload:

				path := filepath.Clean("storage/" + r.FileName)
				err := os.WriteFile(path, r.File, 0666)
				if err != nil {
					return err
				}

			case models.Download:

				path := filepath.Clean("storage/" + r.FileName)
				file, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				k.RecordChannel <- models.Record{
					FileName: r.FileName,
					File:     file,
				}
			case "":
				k.ShutdownChannel <- struct{}{}
			}

		case req := <-k.FilesInfoChannel:

			switch req.RequestType {

			case models.Info:

				f, err := os.Open("storage")
				if err != nil {
					return err
				}

				files, err := f.Readdir(0)
				if err != nil {
					return err
				}

				info, err := fileos.FileInfo(files)
				if err != nil {
					return err
				}

				k.ListChannel <- info
			}

		}
	}
}

func (k *KeeperServiceServer) AddRecord(ctx context.Context, in *pb.Record) (*emptypb.Empty, error) {

	record := models.Record{
		RequestType: models.Upload,
		FileName:    in.FileName,
		File:        in.File,
	}

	k.DownloadUploadChannel <- record

	out := &emptypb.Empty{}
	return out, nil
}

func (k *KeeperServiceServer) GetRecord(ctx context.Context, in *pb.Record) (*pb.Record, error) {

	record := models.Record{
		RequestType: models.Download,
		FileName:    in.FileName,
	}

	k.DownloadUploadChannel <- record

	res := <-k.RecordChannel

	if len(res.File) == 0 {
		return nil, status.Error(codes.NotFound, "")
	}

	out := &pb.Record{
		FileName: res.FileName,
		File:     res.File,
	}

	return out, nil
}

func (k *KeeperServiceServer) GetInfo(ctx context.Context, in *emptypb.Empty) (*pb.List, error) {

	record := models.Record{
		RequestType: models.Info,
	}

	k.FilesInfoChannel <- record

	res := <-k.ListChannel

	if len(res) == 0 {
		return nil, status.Error(codes.NotFound, "")
	}

	out := &pb.List{}

	for _, val := range res {
		out.Note = append(out.Note, &pb.Record{
			FileName:  val.FileName,
			CreatedAt: val.Created,
			UpdatedAt: val.Updated,
		})
	}

	return out, nil
}
