package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type nodeConfig struct {
	id   int
	port int
}

// 각 노드 아이디와 포트 번호 정의
var nodes = []nodeConfig{
	{id: 1, port: 3001},
	{id: 2, port: 3002},
	{id: 3, port: 3003},
}

// 가독성을 위해 각 노드의 아이디별 출력 색상 정의
var colors = map[int]string{
	1: "\033[36m", // 파란색 (cyan)
	2: "\033[33m", // 노란색
	3: "\033[35m", // 핑크색 (magenta)
}

const colorReset = "\033[0m"

// 노드 프로세스 관리를 위한 맵 정의
var (
	// 키 타입이 int이고, 값 타입이 *exec.Cmd인 맵
	processes = make(map[int]*exec.Cmd)
	// 맵에 대한 동시성 접근을 위한 뮤텍스(lock)
	processesMu sync.Mutex
)

// 노드 프로세스 시작 처리 함수
func startNode(nc nodeConfig) {
	// 현재 노드가 통신해야 할 다른 노드의 URL 목록 생성
	var peers []string
	for _, n := range nodes {
		if n.id != nc.id {
			peers = append(peers, fmt.Sprintf("http://localhost:%d", n.port))
		}
	}

	color := colors[nc.id]
	fmt.Printf("%s노드 %d 시작...%s\n", color, nc.id, colorReset)

	// 노드 프로세스 생성
	args := []string{"run", ".", "node", strconv.Itoa(nc.id), strconv.Itoa(nc.port)}
	args = append(args, peers...)

	cmd := exec.Command("go", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("노드 %d stdout pipe 실패: %v\n", nc.id, err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Printf("노드 %d stderr pipe 실패: %v\n", nc.id, err)
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("노드 %d 시작 실패: %v\n", nc.id, err)
		return
	}

	// 여러 goroutine에서 동시에 접근할 수 있으므로, 동시성 접근을 위해 뮤텍스를 사용
	processesMu.Lock()
	processes[nc.id] = cmd
	processesMu.Unlock()

	// goroutine: 노드 프로세스의 출력에 색상을 적용하여 출력
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Printf("%s%s%s\n", color, scanner.Text(), colorReset)
		}
	}()

	// goroutine: 노드 프로세스의 에러 출력에 색상을 적용하여 출력
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Printf("%s[노드 %d 에러] %s%s\n", color, nc.id, scanner.Text(), colorReset)
		}
	}()

	// goroutine: 노드 프로세스가 종료될 때 처리
	go func() {
		err := cmd.Wait()
		exitCode := -1
		if cmd.ProcessState != nil {
			exitCode = cmd.ProcessState.ExitCode()
		}
		fmt.Printf("%s[노드 %d] 프로세스 종료 (code: %d, err: %v)%s\n",
			color, nc.id, exitCode, err, colorReset)

		processesMu.Lock()
		delete(processes, nc.id)
		processesMu.Unlock()
	}()
}

// 노드 프로세스 종료 처리 함수
func killNode(id int) {
	processesMu.Lock()
	// 맵에서 노드 프로세스 조회
	cmd, exists := processes[id]
	processesMu.Unlock()

	// 노드 프로세스가 존재하면 종료
	if exists && cmd.Process != nil {
		if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
			fmt.Printf("[클러스터] 노드 %d 시그널 전송 실패: %v\n", id, err)
		}
		color := colors[id]
		fmt.Printf("\n%s[클러스터] 노드 %d 종료%s\n\n", color, id, colorReset)
	} else {
		fmt.Printf("\n[클러스터] 노드 %d 실행 중이 아님\n\n", id)
	}
}

// 클러스터 상태 출력 함수
func printStatus() {
	fmt.Println("\n=== 클러스터 상태 ===")

	// 타임아웃이 1초인 http.Client 생성
	client := &http.Client{Timeout: 1 * time.Second}

	for _, node := range nodes {
		// 즉시 실행되는 익명 함수. 함수 종료시 defer 처리된 리소스 해제
		func() {
			url := fmt.Sprintf("http://localhost:%d/status", node.port)
			resp, err := client.Get(url)
			color := colors[node.id]

			if err != nil {
				fmt.Printf("%s   노드 %d: 다운%s\n", color, node.id, colorReset)
				return
			}
			// err 발생시 resp는 nil. 따라서 err 체크 후 defer 처리가 일반적인 방법.
			defer resp.Body.Close()

			var data StatusResponse
			// resp.Body를 JSON -> StatusResponse 타입으로 변환; 그리고 err 발생시 처리
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				fmt.Printf("%s   노드 %d: (응답 파싱 실패)%s\n", color, node.id, colorReset)
				return
			}

			leaderMark := "  "
			if data.IsLeader {
				leaderMark = "👑"
			}

			leaderStr := "none"
			if data.CurrentLeaderID != nil {
				leaderStr = strconv.Itoa(*data.CurrentLeaderID)
			}
			promisedStr := "none"
			if data.PromisedProposal != nil {
				promisedStr = strconv.Itoa(*data.PromisedProposal)
			}
			acceptedStr := "none"
			if data.AcceptedProposal != nil {
				acceptedStr = strconv.Itoa(*data.AcceptedProposal)
			}

			// 노드 상태 출력
			fmt.Printf("%s%s 노드 %d: leader=%s, term=%d, isLeader=%v, promised=%s, accepted=%s%s\n",
				color, leaderMark, node.id, leaderStr, data.CurrentTerm, data.IsLeader,
				promisedStr, acceptedStr, colorReset)
		}()
	}
	fmt.Println()
}

func isValidNodeID(id int) bool {
	for _, n := range nodes {
		if n.id == id {
			return true
		}
	}
	return false
}

func findNode(id int) *nodeConfig {
	for _, n := range nodes {
		if n.id == id {
			return &n
		}
	}
	return nil
}

// === 대화형 CLI 구현 ===

func runCluster() {
	// 모든 노드 시작
	fmt.Print("\n🚀 Starting Paxos Leader Election Cluster...\n\n")

	for _, node := range nodes {
		startNode(node)
	}

	// 모든 노드가 시작될 때까지 대기 후 명령어 출력
	time.AfterFunc(1*time.Second, func() {
		fmt.Println()
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println("📋 사용 가능한 명령어:")
		fmt.Println("  kill <id>   - 노드 종료 (특정 노드에 장애가 발생하여 종료되는 걸 시뮬레이션)")
		fmt.Println("  start <id>  - 종료된 노드 재시작")
		fmt.Println("  status      - 클러스터 상태 출력")
		fmt.Println("  exit        - 모든 노드 종료 및 종료")
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println()
	})

	// goroutine 끼리 메세지를 주고받기 위한 채널 생성, 버퍼 크기를 1로 설정하여 수신자가 메세지를 받을 때까지 대기할 수 있도록 함
	sigChan := make(chan os.Signal, 1)
	// Ctrl+C 발생시 해당 시그널을 채널로 전달
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		// 채널에서 메세지(Ctrl+C)를 받을 때까지 대기
		<-sigChan
		fmt.Print("\n\n🛑 모든 노드 종료...\n\n")
		// 모든 노드 프로세스에 SIGTERM 시그널 전송
		processesMu.Lock()
		for _, cmd := range processes {
			if cmd.Process != nil {
				cmd.Process.Signal(syscall.SIGTERM)
			}
		}
		processesMu.Unlock()
		time.Sleep(500 * time.Millisecond)
		os.Exit(0)
	}()

	// 대화형 CLI 명령어 처리
	scanner := bufio.NewScanner(os.Stdin)
	// Stdin이 종료될 때까지 한 줄씩 읽어서 처리
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		cmd := parts[0]
		arg := ""
		if len(parts) > 1 {
			arg = parts[1]
		}

		// 명령어 처리
		switch cmd {
		case "kill":
			id, err := strconv.Atoi(arg)
			if err != nil || !isValidNodeID(id) {
				fmt.Printf("잘못된 노드 ID: %s\n", arg)
				continue
			}
			killNode(id)

		case "start":
			id, err := strconv.Atoi(arg)
			if err != nil || !isValidNodeID(id) {
				fmt.Printf("잘못된 노드 ID: %s\n", arg)
				continue
			}
			processesMu.Lock()
			_, running := processes[id]
			processesMu.Unlock()
			if running {
				fmt.Printf("\n노드 %d 이미 실행 중\n\n", id)
			} else {
				nc := findNode(id)
				if nc != nil {
					startNode(*nc)
				}
			}

		case "status":
			printStatus()

		case "exit":
			fmt.Print("\n🛑 모든 노드 종료...\n\n")
			processesMu.Lock()
			for _, cmd := range processes {
				if cmd.Process != nil {
					cmd.Process.Signal(syscall.SIGTERM)
				}
			}
			processesMu.Unlock()
			time.Sleep(500 * time.Millisecond)
			os.Exit(0)

		case "help":
			fmt.Println("\n📋 사용 가능한 명령어:")
			fmt.Println("  kill <id>   - 노드 종료")
			fmt.Println("  start <id>  - 종료된 노드 재시작")
			fmt.Println("  status      - 클러스터 상태 출력")
			fmt.Println("  exit        - 모든 노드 종료 및 종료")

		default:
			fmt.Printf("알 수 없는 명령어: %s. 'help' 명령어를 사용하세요.\n\n", cmd)
		}
	}
}
