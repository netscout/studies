// 이 파일은 실행가능한 모듈이라는 의미
package main

// 사용할 모듈을 임포트
import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// 프로그램의 진입점
func main() {
	/**
		 * 컨텍스트는 일종의 프로그램의 실행 환경을 제어하는 객체
		 * 컨텍스트에 10초 동안 아무런 신호가 전달되지 않으면 컨텍스트를 취소하고 프로그램을 종료
		 * 특정 연산이 너무 오래걸리는 경우 무작정 대기하지 않도록 함
	     *
		 * := 는 한 줄로 변수를 선언하고 초기화 한다
	*/
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer는 프로그램이 종료될 때 실행되도록 지연시킨다
	defer cancel()

	// mongodb에 연결
	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		// log.Fatal은 오류가 발생하면 프로그램을 종료하고 오류 메시지를 출력
		log.Fatal(err)
	}
	// 프로그램이 종료될 때 연결을 끊기
	defer client.Disconnect(ctx)

	// mongodb 서버에 연결 확인, ping 성공시 컨텍스트가 취소되지 않고 계속 실행
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB")

	// users 컬렉션을 가져옴
	collection := client.Database("testdb").Collection("users")

	// bson.M은 자바스크립트 객체와 유사한 형태의 데이터를 저장하는 Binary JSON 형식
	user := bson.M{
		"name":  "John Doe",
		"email": "john.doe@example.com",
		"age":   30,
	}

	// 사용자를 컬렉션에 삽입
	insertResult, err := collection.InsertOne(ctx, user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("User inserted with ID:", insertResult.InsertedID)

	// 이메일로 사용자를 찾음
	filter := bson.M{"email": "john.doe@example.com"}
	var result bson.M
	err = collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}

	// 찾은 사용자 정보 출력
	fmt.Println("User found:", result)
}
