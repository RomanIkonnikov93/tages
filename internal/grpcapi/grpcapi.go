package grpcapi

import (
	"context"
	"os"

	"github.com/RomanIkonnikov93/tages/internal/fileos"
	"github.com/RomanIkonnikov93/tages/internal/models"
	pb "github.com/RomanIkonnikov93/tages/internal/proto"
	"github.com/RomanIkonnikov93/tages/pkg/pkg/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/protobuf/types/known/emptypb"
)

type KeeperServiceServer struct {
	pb.UnimplementedKeeperServer
	DownloadUploadChannel chan models.Record
	RecordChan            chan models.Record
	FilesInfoChannel      chan models.Record
	ListChan              chan []models.Record
	logger                *logging.Logger
}

func InitServices(logger *logging.Logger) *KeeperServiceServer {

	return &KeeperServiceServer{
		DownloadUploadChannel: make(chan models.Record, models.UploadDownloadParallelCount),
		RecordChan:            make(chan models.Record),
		FilesInfoChannel:      make(chan models.Record, models.FilesInfoParallelCount),
		ListChan:              make(chan []models.Record),
		logger:                logger,
	}
}

func (k *KeeperServiceServer) Run() {

	if _, err := os.Stat("storage"); os.IsNotExist(err) {
		err = os.Mkdir("storage", 0666)
		k.logger.Error(err)
	}

	for {
		select {
		case r := <-k.DownloadUploadChannel:

			switch r.RequestType {

			case models.Upload:

				err := os.WriteFile("storage/"+r.FileName, r.File, 0666)
				if err != nil {
					k.logger.Error(err)
				}

			case models.Download:

				file, err := os.ReadFile("storage/" + r.FileName)
				if err != nil {
					k.logger.Error(err)
				}

				k.RecordChan <- models.Record{
					FileName: r.FileName,
					File:     file,
				}
			}

		case req := <-k.FilesInfoChannel:

			switch req.RequestType {

			case models.Info:

				f, err := os.Open("storage")
				if err != nil {
					k.logger.Error(err)
				}

				files, err := f.Readdir(0)
				if err != nil {
					k.logger.Error(err)
				}

				k.ListChan <- fileos.FileInfo(files)
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

	res := <-k.RecordChan

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

	res := <-k.ListChan

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
