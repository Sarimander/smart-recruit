package grpcclient

import (
	"fmt"

	recruitv1 "logic-grpc-service/proto/gen/recruit/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	Auth       recruitv1.AuthServiceClient
	Job        recruitv1.JobServiceClient
	Candidate  recruitv1.CandidateServiceClient
	Application recruitv1.ApplicationServiceClient
	OSS        recruitv1.OSSServiceClient
	AI         recruitv1.AIServiceClient
	conn       *grpc.ClientConn
}

func New(address string) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial grpc: %w", err)
	}
	return &Client{
		Auth:        recruitv1.NewAuthServiceClient(conn),
		Job:         recruitv1.NewJobServiceClient(conn),
		Candidate:   recruitv1.NewCandidateServiceClient(conn),
		Application: recruitv1.NewApplicationServiceClient(conn),
		OSS:         recruitv1.NewOSSServiceClient(conn),
		AI:          recruitv1.NewAIServiceClient(conn),
		conn:        conn,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
