package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// === 타입 정의 ===

type PrepareRequest struct {
	ProposalNumber int `json:"proposalNumber"`
}

type PrepareResponse struct {
	Promise          bool `json:"promise"`
	AcceptedProposal *int `json:"acceptedProposal"`
	AcceptedValue    *int `json:"acceptedValue"`
}

type AcceptRequest struct {
	ProposalNumber int `json:"proposalNumber"`
	Value          int `json:"value"`
}

type AcceptResponse struct {
	Accepted bool `json:"accepted"`
}

type HeartbeatMessage struct {
	LeaderID int `json:"leaderId"`
	Term     int `json:"term"`
}

type StatusResponse struct {
	NodeID           int  `json:"nodeId"`
	CurrentLeaderID  *int `json:"currentLeaderId"`
	CurrentTerm      int  `json:"currentTerm"`
	PromisedProposal *int `json:"promisedProposal"`
	AcceptedProposal *int `json:"acceptedProposal"`
	AcceptedValue    *int `json:"acceptedValue"`
	IsLeader         bool `json:"isLeader"`
}

// === Paxos 노드 구조체 ===

// PaxosNode는 Paxos 노드의 모든 상태를 관리합니다.
type PaxosNode struct {
	mu sync.Mutex

	// 컨텍스트 (graceful shutdown 용도)
	ctx    context.Context
	cancel context.CancelFunc

	// 설정
	nodeID   int      // 노드 아이디
	port     int      // 노드 포트
	peers    []string // 통신해야 할 다른 노드의 URL 목록
	allNodes []string // 자신을 포함한 모든 노드의 URL 목록
	majority int      // 다수 필요 개수 (전체 노드 수 / 2 + 1)

	// === Paxos Acceptor 상태 변수들 ===
	promisedProposal *int // 약속한 제안 번호 (nil = 아직 약속 없음)
	acceptedProposal *int // 수락된 제안 번호 (nil = 아직 수락 없음)
	acceptedValue    *int // 수락된 값(리더의 노드 아이디) (nil = 아직 수락 없음)

	// === Leader 상태와 Heartbeat ===
	currentLeaderID *int // 현재 리더 아이디 (nil = 리더 없음)
	currentTerm     int  // 현재 턴
	electionRound   int  // 선거 라운드

	// === 타이머 ===
	heartbeatStop      chan struct{} // 리더의 하트비트 전송 중지 시그널
	electionTimer      *time.Timer   // 하트비트 타임아웃 감시 타이머 (3~5초, 리더는 선출 후 중지)
	nextElectionTimer  *time.Timer   // 다음 선거 시작 타이머
	electionInProgress bool          // 선거 진행 중 플래그 (동시 선거 방지)

	// HTTP 클라이언트
	httpClient *http.Client
}

// newPaxosNode는 새로운 Paxos 노드를 생성합니다.
// cli 매개변수로 전달받은 노드 아이디, 포트, 통신해야 할 다른 노드의 URL 목록 등의 값을 변수에 저장
func newPaxosNode(ctx context.Context, nodeID, port int, peers []string) *PaxosNode {
	ctx, cancel := context.WithCancel(ctx)
	selfURL := fmt.Sprintf("http://localhost:%d", port)
	allNodes := append([]string{selfURL}, peers...)
	majority := len(allNodes)/2 + 1

	return &PaxosNode{
		ctx:        ctx,
		cancel:     cancel,
		nodeID:     nodeID,
		port:       port,
		peers:      peers,
		allNodes:   allNodes,
		majority:   majority,
		httpClient: &http.Client{Timeout: 1 * time.Second},
	}
}

// intPtr는 int를 *int로 변환하는 헬퍼 함수입니다.
func intPtr(v int) *int {
	return &v
}

// ceilDiv는 올림 나눗셈을 수행합니다.
func ceilDiv(a, b int) int {
	return (a + b - 1) / b
}

// === Paxos Proposer 로직 ===

// runElection은 리더 선출 시작
func (node *PaxosNode) runElection() {
	node.mu.Lock()
	// 이미 리더 선출 중인 경우
	if node.electionInProgress {
		node.mu.Unlock()
		// 리더 선출 중단
		return
	}
	node.electionInProgress = true

	// 제안 번호가 이미 있는 것보다 큰지 확인
	if node.promisedProposal != nil {
		/**
		 * 이전에 약속된 제안 번호가 있는 경우, 라운드 조정(다른 노드가 이미 제안한 제안 번호보다 높은 제안 번호를 사용하도록 함)
		 *
		 * ex: max(0, ceilDiv(6, 3)) = max(0, 2) = 2
		 */
		node.electionRound = max(node.electionRound, ceilDiv(*node.promisedProposal, len(node.allNodes)))
	}
	node.electionRound++

	/**
	 * 제안 번호 = 노드 아이디 + (라운드 * 노드 개수) -> 각 라운드 마다 노드의 제안 번호의 고유성을 보장하기 위함
	 *
	 * 예를 들어, 노드 1, 2, 3이 있고, 라운드가 1인 경우,
	 * 노드 1, 라운드 1: 1 + (1 * 3) = 4
	 * 노드 2, 라운드 1: 2 + (1 * 3) = 5
	 * 노드 3, 라운드 1: 3 + (1 * 3) = 6
	 * 노드 1, 라운드 2: 1 + (2 * 3) = 7
	 * 노드 2, 라운드 2: 2 + (2 * 3) = 8
	 * 노드 3, 라운드 2: 3 + (2 * 3) = 9
	 * ...
	 */
	proposalNumber := node.nodeID + (node.electionRound * len(node.allNodes))
	allNodes := make([]string, len(node.allNodes))
	/**
	 * lock은 최대한 짧게 유지하는게 중요하므로 필요한 작업 이후에 해제해야 함.
	 * 따라서 이후 작업에 사용할 node.allNodes를 로컬 변수에 복사하여 사용.
	 */
	copy(allNodes, node.allNodes)
	majority := node.majority
	nodeID := node.nodeID
	// lock 해제
	node.mu.Unlock()

	// 함수 종료시 리더 선출 상태를 초기화하기 위해 defer 처리
	defer func() {
		node.mu.Lock()
		node.electionInProgress = false
		node.mu.Unlock()
	}()

	fmt.Printf("\n[노드 %d] 🗳️  리더 선출 시작 / 제안 번호 #%d\n", nodeID, proposalNumber)

	// === Phase 1: PREPARE 요청 전송 ===
	fmt.Printf("[노드 %d] Phase 1: 모든 노드에 PREPARE 요청 전송...\n", nodeID)

	type prepareResult struct {
		resp *PrepareResponse
		err  error
	}
	results := make([]prepareResult, len(allNodes))
	var wg sync.WaitGroup

	// 자신을 포함한 모든 노드에 PREPARE 요청 전송
	for i, nodeURL := range allNodes {
		wg.Add(1)
		go func(idx int, url string) {
			defer wg.Done()
			resp, err := node.sendPrepare(url, proposalNumber)
			results[idx] = prepareResult{resp: resp, err: err}
		}(i, nodeURL)
	}
	wg.Wait()

	// PREPARE 요청 결과 처리 - 약속을 받은 것들만 추출
	var promises []PrepareResponse
	for _, r := range results {
		if r.err == nil && r.resp != nil && r.resp.Promise {
			promises = append(promises, *r.resp)
		}
	}

	fmt.Printf("[노드 %d] Phase 1: 약속을 받은 노드 %d/%d 개 (필요 %d 개)\n",
		nodeID, len(promises), len(allNodes), majority)

	if len(promises) < majority {
		fmt.Printf("[노드 %d] ❌ Phase 1 실패: 약속을 받은 노드가 다수 필요 개수보다 적습니다. 리더 선출을 중단하고 다음 선거 시작.\n", nodeID)
		node.scheduleNextElection()
		return
	}

	// 수락된 값 중 가장 높은 값을 채택, 없으면 자신을 제안
	proposedValue := nodeID
	highestAcceptedProposal := -1

	for _, p := range promises {
		// 수락된 값 중 가장 높은 값을 채택
		if p.AcceptedProposal != nil && *p.AcceptedProposal > highestAcceptedProposal {
			highestAcceptedProposal = *p.AcceptedProposal
			proposedValue = *p.AcceptedValue
		}
	}

	if highestAcceptedProposal > -1 {
		fmt.Printf("[노드 %d] Phase 1: 이전에 수락된 값 (노드 %d) 채택 / 제안 번호 #%d\n",
			nodeID, proposedValue, highestAcceptedProposal)
	} else {
		fmt.Printf("[노드 %d] Phase 1: 이전에 수락된 값이 없으므로 자신을 제안 (노드 %d)\n", nodeID, proposedValue)
	}

	// === Phase 2: Accept ===
	fmt.Printf("[노드 %d] Phase 2: 모든 노드에 ACCEPT (값: 노드 %d) 요청 전송...\n", nodeID, proposedValue)

	type acceptResult struct {
		resp *AcceptResponse
		err  error
	}
	// ACCEPT 요청 결과를 저장할 슬라이스 생성
	acceptResults := make([]acceptResult, len(allNodes))
	// 모든 작업이 완료될 때까지 대기할 WaitGroup 생성
	wg = sync.WaitGroup{}

	// 자신을 포함한 모든 노드에 ACCEPT 요청 전송, 병렬 처리
	for i, nodeURL := range allNodes {
		// 대기 그룹에 작업 1개 추가
		wg.Add(1)
		// goroutine: ACCEPT 요청 전송
		go func(idx int, url string) {
			// 작업 완료시 대기 그룹에서 작업 1개 제거
			defer wg.Done()
			resp, err := node.sendAccept(url, proposalNumber, proposedValue)
			acceptResults[idx] = acceptResult{resp: resp, err: err}
		}(i, nodeURL)
	}
	// 모든 작업이 완료될 때까지 대기
	wg.Wait()

	// ACCEPT 요청 결과 처리 - 수락된 것들 카운트
	accepts := 0
	for _, r := range acceptResults {
		if r.err == nil && r.resp != nil && r.resp.Accepted {
			accepts++
		}
	}

	fmt.Printf("[노드 %d] Phase 2: 수락된 노드 %d/%d 개 (필요 %d 개)\n",
		nodeID, accepts, len(allNodes), majority)

	if accepts >= majority {
		fmt.Printf("[노드 %d] ✅ 리더 선출 성공! 노드 %d 가 리더가 되었습니다. (제안 번호 %d)\n",
			nodeID, proposedValue, proposalNumber)

		if proposedValue == nodeID {
			fmt.Printf("[노드 %d] 👑 나는 리더!\n", nodeID)
			// 리더는 자신의 선출 타임아웃을 중지 (팔로워만 타임아웃을 감시)
			node.mu.Lock()
			if node.electionTimer != nil {
				node.electionTimer.Stop()
			}
			node.mu.Unlock()
			node.startHeartbeat(proposalNumber)
		}
	} else {
		fmt.Printf("[노드 %d] ❌ Phase 2 실패: 수락된 노드가 다수 필요 개수보다 적습니다. 리더 선출을 중단하고 다음 선거 시작.\n", nodeID)
		node.scheduleNextElection()
	}
}

func (node *PaxosNode) sendPrepare(url string, proposalNumber int) (*PrepareResponse, error) {
	body, err := json.Marshal(PrepareRequest{ProposalNumber: proposalNumber})
	if err != nil {
		return nil, fmt.Errorf("marshal prepare request: %w", err)
	}

	req, err := http.NewRequestWithContext(node.ctx, "POST", url+"/paxos/prepare", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create prepare request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := node.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send prepare to %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prepare request to %s returned status %d", url, resp.StatusCode)
	}

	var result PrepareResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode prepare response from %s: %w", url, err)
	}

	return &result, nil
}

func (node *PaxosNode) sendAccept(url string, proposalNumber, value int) (*AcceptResponse, error) {
	body, err := json.Marshal(AcceptRequest{ProposalNumber: proposalNumber, Value: value})
	if err != nil {
		return nil, fmt.Errorf("marshal accept request: %w", err)
	}

	req, err := http.NewRequestWithContext(node.ctx, "POST", url+"/paxos/accept", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create accept request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := node.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send accept to %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("accept request to %s returned status %d", url, resp.StatusCode)
	}

	var result AcceptResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode accept response from %s: %w", url, err)
	}

	return &result, nil
}

func (node *PaxosNode) sendHeartbeat(url string, leaderID, term int) error {
	body, err := json.Marshal(HeartbeatMessage{LeaderID: leaderID, Term: term})
	if err != nil {
		return fmt.Errorf("marshal heartbeat: %w", err)
	}

	req, err := http.NewRequestWithContext(node.ctx, "POST", url+"/leader/heartbeat", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create heartbeat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := node.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send heartbeat to %s: %w", url, err)
	}
	defer resp.Body.Close()

	return nil
}

// 리더가 자신을 생존을 알리는 하트비트 전송을 시작하는 함수 (리더만 실행)
func (node *PaxosNode) startHeartbeat(term int) {
	node.mu.Lock()
	// 기존 하트비트 중지 (TOCTOU 경합 방지를 위해 인라인 처리)
	if node.heartbeatStop != nil {
		// 기존 하트비트 채널 종료
		close(node.heartbeatStop)
	}
	// 데이터는 없이 신호만 수신할 채널 생성
	stopCh := make(chan struct{})
	node.heartbeatStop = stopCh
	node.mu.Unlock()

	// goroutine: 각 채널의 메세지 수신 및 처리
	go func() {
		// 1초 간격 타이머 채널 생성
		ticker := time.NewTicker(1 * time.Second)
		// goroutine 종료시 타이머 종료
		defer ticker.Stop()

		for { // 종료될 때까지 무한 반복하는 이벤트 루프
			// 각 채널에서 수신되는 신호를 대기하며, 수신된 신호에 따라 동작 결정(select)
			select {
			case <-stopCh: // 하트비트 중지 신호
				return
			case <-node.ctx.Done(): // context 취소 신호
				return
			case <-ticker.C: // 1초 간격 타이머 채널 신호
				fmt.Printf("[노드 %d] 💓 하트비트 전송 (턴 %d)\n", node.nodeID, term)

				node.mu.Lock()
				// 이후 작업에 사용할 node.peers를 로컬 변수에 복사하여 사용
				peers := make([]string, len(node.peers))
				copy(peers, node.peers)
				// lock 해제
				node.mu.Unlock()

				// 자신(리더)를 제외한 나머지 노드에 하트비트 전송
				for _, peerURL := range peers {
					// goroutine: 하트비트 전송
					go func(url string) {
						_ = node.sendHeartbeat(url, node.nodeID, term)
					}(peerURL)
				}
			}
		}
	}()
}

// 하트비트 중지 처리 함수
func (node *PaxosNode) stopHeartbeat() {
	node.mu.Lock()
	if node.heartbeatStop != nil {
		close(node.heartbeatStop)
		node.heartbeatStop = nil
	}
	node.mu.Unlock()
}

// 모든 타이머와 하트비트를 정리하고 context를 취소하는 함수
func (node *PaxosNode) shutdown() {
	node.mu.Lock()
	if node.electionTimer != nil {
		node.electionTimer.Stop()
	}
	if node.nextElectionTimer != nil {
		node.nextElectionTimer.Stop()
	}
	node.mu.Unlock()
	node.stopHeartbeat()
	node.cancel()
}

// 리더 선출 타임아웃 리셋 처리 함수
//
// 타임아웃 이후, 리더가 없으면 리더 선출 시작
func (node *PaxosNode) resetElectionTimeout() {
	node.mu.Lock()
	defer node.mu.Unlock()

	if node.electionTimer != nil {
		// 기존 리더 선출 타이머 중지
		node.electionTimer.Stop()
	}

	// 리더 선출 타임아웃 랜덤 적용 (3-5초)
	timeout := 3*time.Second + time.Duration(rand.Float64()*2000)*time.Millisecond

	// goroutine: 리더 선출 타임아웃 타이머 설정
	node.electionTimer = time.AfterFunc(timeout, func() {
		fmt.Printf("[노드 %d] ⏰ 하트비트 타임아웃 - 리더 없음 감지, 리더 선출 시작...\n", node.nodeID)

		node.mu.Lock()
		node.currentLeaderID = nil
		// 현재 리더로 선출된 노드가 서비스 다운되는 경우에 해당 노드는 리더 선출에서 제외되어야 함
		// 따라서 수락된 상태를 리셋하여 진행하며, 각 리더 선출은 사실상 새로운 Paxos 인스턴스가 되도록 함
		node.acceptedProposal = nil
		node.acceptedValue = nil
		node.mu.Unlock()

		node.runElection()
	})
}

// 다음 리더 선출 예약 처리 함수
func (node *PaxosNode) scheduleNextElection() {
	node.mu.Lock()
	defer node.mu.Unlock()

	if node.nextElectionTimer != nil {
		// 기존 다음 리더 선출 타이머 중지
		node.nextElectionTimer.Stop()
	}

	// 다음 선거 시작 전 랜덤 딜레이 적용
	delay := 2*time.Second + time.Duration(rand.Float64()*3000)*time.Millisecond

	// goroutine: 다음 리더 선출 타이머 설정
	node.nextElectionTimer = time.AfterFunc(delay, func() {
		node.mu.Lock()
		hasLeader := node.currentLeaderID != nil
		node.mu.Unlock()

		// 현재 리더가 없으면 다음 선거 시작
		if !hasLeader {
			node.runElection()
		}
	})
}

// === HTTP 핸들러들 ===

// Paxos Phase 1: Prepare 요청 처리 함수
func (node *PaxosNode) handlePrepare(w http.ResponseWriter, r *http.Request) {
	// 요청 본문을 JSON -> PrepareRequest 타입으로 변환; 그리고 err 발생시 처리
	var req PrepareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	node.mu.Lock()
	fmt.Printf("[노드 %d] PREPARE 요청 수신 / 제안 번호 #%d\n", node.nodeID, req.ProposalNumber)

	// 함수 내부에서만 쓰이므로 포인터(*) 없이 사용
	var resp PrepareResponse

	if node.promisedProposal == nil || req.ProposalNumber > *node.promisedProposal {
		node.promisedProposal = intPtr(req.ProposalNumber)
		// 수락된 상태를 리셋하여 진행하며, 각 리더 선출은 사실상 새로운 Paxos 인스턴스가 되도록 함(죽은 리더 재선출 방지)
		node.acceptedProposal = nil
		node.acceptedValue = nil

		fmt.Printf("[노드 %d] → 제안 번호 #%d 약속\n", node.nodeID, req.ProposalNumber)
		resp = PrepareResponse{Promise: true, AcceptedProposal: nil, AcceptedValue: nil}
	} else {
		fmt.Printf("[노드 %d] → 이미 약속한 제안 번호 #%d 거절\n", node.nodeID, *node.promisedProposal)
		resp = PrepareResponse{Promise: false, AcceptedProposal: nil, AcceptedValue: nil}
	}
	node.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		fmt.Printf("[노드 %d] 응답 쓰기 실패: %v\n", node.nodeID, err)
	}
}

// Paxos Phase 2: ACCEPT 요청 처리 함수
func (node *PaxosNode) handleAccept(w http.ResponseWriter, r *http.Request) {
	var req AcceptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	node.mu.Lock()
	fmt.Printf("[노드 %d] ACCEPT 요청 수신 / 제안 번호 #%d, 값: 노드 %d\n",
		node.nodeID, req.ProposalNumber, req.Value)

	var resp AcceptResponse

	// 약속된 제안 번호가 없거나, 현재 제안 번호가 약속된 제안 번호보다 크거나 같으면 수락
	if node.promisedProposal == nil || req.ProposalNumber >= *node.promisedProposal {
		// 약속된 제안 번호 갱신
		node.promisedProposal = intPtr(req.ProposalNumber)
		// 수락된 제안 번호 갱신
		node.acceptedProposal = intPtr(req.ProposalNumber)
		// 수락된 값 갱신(리더의 노드 아이디)
		node.acceptedValue = intPtr(req.Value)
		// 현재 리더 아이디 갱신
		node.currentLeaderID = intPtr(req.Value)
		// 현재 턴 갱신
		node.currentTerm = req.ProposalNumber

		fmt.Printf("[노드 %d] → 수락! 새로운 리더: 노드 %d (제안 번호 %d)\n",
			node.nodeID, req.Value, req.ProposalNumber)

		// 리더가 아니라면 하트비트 중지
		shouldStopHeartbeat := req.Value != node.nodeID
		node.mu.Unlock()

		if shouldStopHeartbeat {
			node.stopHeartbeat()
		}

		// 리더 선출 타임아웃 리셋
		node.resetElectionTimeout()

		resp = AcceptResponse{Accepted: true}
	} else {
		fmt.Printf("[노드 %d] → 이미 약속한 제안 번호 #%d 거절\n",
			node.nodeID, *node.promisedProposal)
		node.mu.Unlock()
		resp = AcceptResponse{Accepted: false}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		fmt.Printf("[노드 %d] 응답 쓰기 실패: %v\n", node.nodeID, err)
	}
}

// 리더가 전송하는 하트비트 수신 처리 함수
func (node *PaxosNode) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	var msg HeartbeatMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	node.mu.Lock()
	// 새로운 턴이 시작되었다면,
	shouldReset := msg.Term >= node.currentTerm
	if shouldReset {
		// 새로운 리더가 등장했다면,
		if node.currentLeaderID == nil || *node.currentLeaderID != msg.LeaderID {
			fmt.Printf("[노드 %d] 💓 새로운 리더로부터 하트비트 수신 / 리더: 노드 %d (턴 %d)\n",
				node.nodeID, msg.LeaderID, msg.Term)
		}

		// 현재 리더 아이디 갱신
		node.currentLeaderID = intPtr(msg.LeaderID)
		// 현재 턴 갱신
		node.currentTerm = msg.Term
	}
	node.mu.Unlock()

	/**
	 * 리더 선출 타임아웃 리셋(타임아웃 이후, 리더가 없으면 리더 선출 시작)
	 *
	 * 리더의 하트비트가 안정적으로 수신된다면, 리더 선출의 타임아웃은 계속 리셋된다. -> 현재 상태 유지.
	 * 현재 턴 이상의 하트비트만 타임아웃 리셋 (오래된 리더의 하트비트는 무시)
	 */
	if shouldReset {
		node.resetElectionTimeout()
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"ok": true}); err != nil {
		fmt.Printf("[노드 %d] 응답 쓰기 실패: %v\n", node.nodeID, err)
	}
}

// 상태 조회 엔드포인트 처리 함수
func (node *PaxosNode) handleStatus(w http.ResponseWriter, r *http.Request) {
	node.mu.Lock()
	isLeader := node.currentLeaderID != nil && *node.currentLeaderID == node.nodeID
	resp := StatusResponse{
		NodeID:           node.nodeID,
		CurrentLeaderID:  node.currentLeaderID,
		CurrentTerm:      node.currentTerm,
		PromisedProposal: node.promisedProposal,
		AcceptedProposal: node.acceptedProposal,
		AcceptedValue:    node.acceptedValue,
		IsLeader:         isLeader,
	}
	node.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		fmt.Printf("[노드 %d] 응답 쓰기 실패: %v\n", node.nodeID, err)
	}
}

// === 메인 함수 ===

func runNode() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run . node <id> <port> <peers...>")
		os.Exit(1)
	}

	nodeID, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid node ID %q: %v\n", os.Args[2], err)
		os.Exit(1)
	}

	port, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid port %q: %v\n", os.Args[3], err)
		os.Exit(1)
	}

	peers := os.Args[4:]

	ctx := context.Background()
	node := newPaxosNode(ctx, nodeID, port, peers)

	fmt.Printf("[노드 %d] 시작 with 통신해야 할 다른 노드의 URL 목록: %s\n",
		nodeID, strings.Join(peers, ", "))
	fmt.Printf("[노드 %d] 다수 필요: %d 중 %d 이상\n",
		nodeID, len(node.allNodes), node.majority)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /paxos/prepare", node.handlePrepare)
	mux.HandleFunc("POST /paxos/accept", node.handleAccept)
	mux.HandleFunc("POST /leader/heartbeat", node.handleHeartbeat)
	mux.HandleFunc("GET /status", node.handleStatus)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		fmt.Printf("[노드 %d] 🚀 포트 %d에서 수신 중...\n", nodeID, port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("[노드 %d] 서버 에러: %v\n", nodeID, err)
			os.Exit(1)
		}
	}()

	// 리더 선출 예약(랜덤 딜레이 적용)
	delay := 1*time.Second + time.Duration(rand.Float64()*2000)*time.Millisecond
	fmt.Printf("[노드 %d] 리더 조회 전 %dms 대기...\n", nodeID, delay.Milliseconds())

	// 랜덤 딜레이 이후, 리더가 없으면 리더 선출 시작
	time.AfterFunc(delay, func() {
		node.mu.Lock()
		hasLeader := node.currentLeaderID != nil
		node.mu.Unlock()

		if !hasLeader {
			fmt.Printf("[노드 %d] 리더 없음, 리더 선출 시작...\n", nodeID)
			node.runElection()
		}
	})

	// 리더 선출 타임아웃 리셋(타임아웃 이후, 리더가 없으면 리더 선출 시작)
	node.resetElectionTimeout()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Printf("[노드 %d] 종료 중...\n", nodeID)
	node.shutdown()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownCancel()
	_ = server.Shutdown(shutdownCtx)
}
