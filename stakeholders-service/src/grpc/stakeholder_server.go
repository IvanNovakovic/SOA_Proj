package grpc

import (
	"context"
	"strings"
	"time"

	pb "github.com/IvanNovakovic/SOA_Proj/protos"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"stakeholders-service/auth"
	"stakeholders-service/repository"
)

type StakeholderServer struct {
	pb.UnimplementedStakeholderServiceServer
	repo *repository.UserRepository
}

func NewStakeholderServer(repo *repository.UserRepository) *StakeholderServer {
	return &StakeholderServer{repo: repo}
}

func (s *StakeholderServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	username := strings.TrimSpace(req.Username)
	if username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password required")
	}

	// find by username or email
	filter := bson.M{"$or": []bson.M{{"username": username}, {"email": username}}}
	users, err := s.repo.List(ctx, filter, 0, 1)
	if err != nil || len(users) == 0 {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	u := users[0]
	if !auth.CheckPassword(u.Password, req.Password) {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	// Check if user is blocked
	if u.IsBlocked {
		return nil, status.Error(codes.PermissionDenied, "account has been blocked")
	}

	token, err := auth.IssueToken(u.ID.Hex(), u.Username, u.Roles, 60*time.Minute)
	if err != nil {
		return nil, status.Error(codes.Internal, "token generation error")
	}

	return &pb.LoginResponse{
		Token:  token,
		UserId: u.ID.Hex(),
	}, nil
}
