package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("사용법:")
		fmt.Println("  go run . cluster              -- 클러스터 모드 (3개 노드 실행)")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "node":
		runNode() // 클러스터 관리 cli가 각 노드를 실행할 때 사용
	case "cluster":
		runCluster() // 클러스터 관리 cli를 실행
	default:
		fmt.Printf("알 수 없는 명령어: %s\n", os.Args[1])
		os.Exit(1)
	}
}
