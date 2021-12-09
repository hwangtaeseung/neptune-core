package grpcwrapper

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"time"
)

// GrpcRequest Grpc Client 요청
func GrpcRequest(uri string, timeout int64, callback func(conn *grpc.ClientConn, ctx context.Context) (interface{}, error)) (interface{}, error) {

	// connection 생성
	connection, err := grpc.Dial(uri, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Println("pt cmd connect fail")
		return nil, err
	}
	defer func() {
		err := connection.Close()
		if err != nil {
			log.Printf("grpc client close error (err:%v)", err)
		}
	}()

	// timer 설정
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout) * time.Second)
	defer cancel()

	// stub callback 실행
	response, err := callback(connection, ctx)
	if err != nil {
		log.Println("pt stub error")
		return nil, err
	}

	return response, nil
}
