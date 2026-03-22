package main

import (
	"context"
	"grpc-helloworld/pb"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// 1. 서버에 연결하기
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()

	// 2. GreeterClient 생성
	client := pb.NewGreeterClient(conn)

	// 3. SayHello 메서드 호출(1초 동안 서버가 응답하지 않으면 컨텍스트 취소)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := client.SayHello(ctx, &pb.HelloRequest{Name: "World"})
	if err != nil {
		log.Fatalf("failed to say hello: %v", err)
	}

	log.Printf("Server response: %s", r.GetMessage())
}
