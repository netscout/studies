# go를 알아보자(1)

새해가 되고 뭔가 해야하지 않을까... 싶은 불안감이 커지던 중 평소에 `node.js의 다음은 go 아니겠나` 하고 생각하던 게 떠올랐습니다. 물론 node.js는 대체 불가입니다. 풀스택 개발자에게는 내 고향같은 편안함이 있죠. DB 부터 백엔드, 프론트까지 모든 스택을 하나로 다룰 수 있고 Typescript를 사용하면 정적 타입의 편안함까지 느낄 수 있죠.(물론 타입이 너무 복잡해서 아직도 잘 모릅니다... 하하하하...)

그런데 아무래도 만드는 제품을 on-premiss 솔루션화 하려다 보면, node.js 는 약간 어색함이 있습니다. 어쨌든 스크립트니까 코드가 그대로 노출된다는 단점이 있죠. 난독화를 하긴 하지만, 여전히 불안한 느낌은 있습니다. 물론 바이너리 역시도 100% 안전한 건 없지만, 뭐 그래도 말이죠.

하여간 이제 아주 천천히, 하던대로 go를 알아볼까 합니다. 평소에 많이 하던 코드를 go로 간단하게 작성해보면서 말이죠!

## mongodb야 반갑다!

예, mongodb를 이용할 겁니다. 그리고 아주 간단하게 도큐먼트를 하나 집어넣고, 집어 넣은 도큐먼트를 다시 꺼내오는 작업을 할 겁니다.

### go를 설치하자.

현재 go 최신버전은 1.25.5 입니다. 저는 맥에서 작성중이기 때문에 homebrew로 그냥 설치하면 됩니다.

```bash
> brew install go
> go version
go version go1.25.5 darwin/arm64
```

### 프로젝트 셋업

그리고 적당한 곳에 폴더를 생성합니다.

```
> mkdir mongodb-example
> cd mongodb-example
# go 모듈 초기화
> go mod init mongodb-example
```

그리고 가장 좋아하는 ide를 띄웁니다. 그러면 go.mod 파일이 하나 보일겁니다.

```
module test

go 1.25.5
```

모듈 이름과 go 버전이 보이는데요, 이 파일은 node.js의 package.json 같은 역할을 합니다. 일단 이번에 우리가 사용할 모듈을 설치해볼까요?

```bash
> go get go.mongodb.org/mongo-driver/v2/mongo
go: added github.com/golang/snappy v1.0.0
go: added github.com/klauspost/compress v1.16.7
go: added github.com/xdg-go/pbkdf2 v1.0.0
go: added github.com/xdg-go/scram v1.1.2
go: added github.com/xdg-go/stringprep v1.0.4
go: added github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78
go: added go.mongodb.org/mongo-driver/v2 v2.4.1
go: added golang.org/x/crypto v0.33.0
go: added golang.org/x/sync v0.11.0
go: added golang.org/x/text v0.22.0
```

자 그리고 다시 go.mod 파일을 보면!

```
module test

go 1.25.5

require (
	github.com/golang/snappy v1.0.0 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.mongodb.org/mongo-driver/v2 v2.4.1 // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/text v0.22.0 // indirect
)
```

설치된 패키지 목록이 보입니다. package.json 이랑 진짜 비슷하죠? 근데 저 많은 `// indirect` 는 뭘까요?

그러니까 직접적으로 사용하는 모듈(`go.mongodb.org/mongo-driver/v2`)가 간접적으로 사용하는 모듈이라는 뜻입니다. 근데 왜 `go.mongodb.org/mongo-driver/v2` 도 `// indirect` 냐고요? 아직 이 모듈을 사용하는 코드가 없기 때문입니다!

### 그럼 코드를 작성해보자.

main.go를 작성할 차례입니다.

```go
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
```

그리고 `go mod tidy` 를 실행한 뒤 다시 go.mod 파일을 보면 다음과 같습니다.

```
module test

go 1.25.5

require go.mongodb.org/mongo-driver/v2 v2.4.1

require (
	github.com/golang/snappy v1.0.0 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/text v0.22.0 // indirect
)
```

직접 임포트해서 사용하는 `go.mongodb.org/mongo-driver/v2` 모듈이 `// indirect` 에서 독립했습니다!

### mongodb를 띄우자.

간단한 `docker-compose.yml` 을 사용해서 mongodb를 띄울 겁니다.

```yaml
services:
  mongodb:
    image: mongo:latest
    ports:
      - "27017:27017"
    volumes:
      - mongodb_data:/data/db

volumes:
  mongodb_data:
```

이걸 이렇게 띄우죠.

```bash
> docker-compose up -d
```

### 이제 결과를 보자

이제 결과를 확인해볼까요?

```bash
> go run main.go
Connected to MongoDB
User inserted with ID: ObjectID("6961a3ea4f62d4037b2611bb")
User found: {"_id":{"$oid":"6961a3ea4f62d4037b2611bb"},"name":"John Doe","email":"john.doe@example.com","age":{"$numberInt":"30"}}
```

예... 이게 끝입니다. 처음이니까요. 하하하하하하.

## 참고자료

- [MongoDB Go Driver](https://pkg.go.dev/go.mongodb.org/mongo-driver/v2#section-readme)
- Perplexity와 Claude