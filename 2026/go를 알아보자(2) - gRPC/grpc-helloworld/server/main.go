package main

import (
	"context"
	"log"
	"net"

	"grpc-helloworld/pb"

	"google.golang.org/grpc"
)

type server struct {
	/*
		"protoc --go_out=. --go-grpc_out=. greet.proto" 명령으로 생성된 코드에는 greeterServer, greeterClient 구현이 있음.
		UnimplementedGreeterServer 구조체 역시 protoc가 생성한 것으로 proto 파일에 정의된 Greeter 서비스의 모든 메서드에 대해 껍데기 구현을 가지고 있음.(호출시 codes.Unimplemented 오류 리턴)
		이는 추후 proto 파일에 새로운 메서드가 추가되고, 이에 대한 구현이 되지 않았을 때 컴파일 단계의 오류를 방지하기 위함.
		-> 컴파일은 잘 되지만, 해당 함수를 호출하면 codes.Unimplemented 오류 리턴.
	*/
	pb.UnimplementedGreeterServer
}

// pb.UnimplementedGreeterServer의 SayHello 함수를 오버라이드
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	// 50051번 포트에 대해 네트워크 리스너를 생성
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// gRPC 서버 생성
	s := grpc.NewServer()
	// Greeter grpc 서비스에 서버를 등록
	pb.RegisterGreeterServer(s, &server{})

	// 서버 시작
	log.Println("Server listening at :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
