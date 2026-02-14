import express from 'express';

// === 타입 정의 ===

type NodeConfig = {
  id: number;
  port: number;
  peers: string[];
};

type PrepareRequest = {
  proposalNumber: number;
};

type PrepareResponse = {
  promise: boolean;
  acceptedProposal: number | null;
  acceptedValue: number | null;
};

type AcceptRequest = {
  proposalNumber: number;
  value: number;
};

type AcceptResponse = {
  accepted: boolean;
};

type HeartbeatMessage = {
  leaderId: number;
  turn: number;
};

// cli 매개변수로 전달받은 노드 아이디, 포트, 통신해야 할 다른 노드의 URL 목록 등의 값을 변수에 저장
const nodeId = parseInt(process.argv[2]);
const port = parseInt(process.argv[3]);
const peers = process.argv.slice(4);
const allNodes = [`http://localhost:${port}`, ...peers];
const majority = Math.floor(allNodes.length / 2) + 1;

console.log(`[노드 ${nodeId}] 시작 with 통신해야 할 다른 노드의 URL 목록: ${peers.join(', ')}`);
console.log(`[노드 ${nodeId}] 다수 필요: ${allNodes.length} 중 ${majority} 이상`);

// === Paxos Acceptor 상태 변수들 ===

let promisedProposal: number | null = null;
let acceptedProposal: number | null = null;
// 수락된 값(리더의 노드 아이디)
let acceptedValue: number | null = null;

// === Leader 상태와 Heartbeat ===

let currentLeaderId: number | null = null;
let currentTurn: number = 0;
// 리더의 하트비트 전송 타이머
let heartbeatTimer: NodeJS.Timeout | null = null;
// 리더 부재시 선출을 진행하기 위한 타이머
let electionTimer: NodeJS.Timeout | null = null;
let electionRound = 0;

// === Paxos Proposer 로직 ===

/**
 * 리더 선출 시작
 * @returns 리더 선출 성공 여부, 실패시 null 반환
 */
async function runElection(): Promise<boolean | null> {
  // 제안 번호가 이미 있는 것보다 큰지 확인
  if (promisedProposal !== null) {
    /**
     * 이전에 약속된 제안 번호가 있는 경우, 라운드 조정(다른 노드가 이미 제안한 제안 번호보다 높은 제안 번호를 사용하도록 함)
     * 
     * ex: Math.max(0, Math.ceil(6 / 3)) = Math.max(0, 2) = 2
     */
    electionRound = Math.max(electionRound, Math.ceil(promisedProposal / allNodes.length));
  }
  electionRound++;

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
  const proposalNumber = nodeId + (electionRound * allNodes.length);

  console.log(`\n[노드 ${nodeId}] 🗳️  리더 선출 시작 / 제안 번호 #${proposalNumber}`);

  // Phase 1: PREPARE 요청 전송
  console.log(`[노드 ${nodeId}] Phase 1: 모든 노드에 PREPARE 요청 전송...`);

  const preparePromises = allNodes.map(async (nodeUrl) => {
    // 자신을 포함한 모든 노드에 PREPARE 요청 전송
    try {
      // PREPARE 요청 전송
      const response = await fetch(`${nodeUrl}/paxos/prepare`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ proposalNumber } as PrepareRequest),
        signal: AbortSignal.timeout(1000),
      });

      // PREPARE 요청 실패 시 null 반환
      if (!response.ok) return null;
      return await response.json() as PrepareResponse;
    } catch (error) {
      // PREPARE 요청 실패 시 null 반환
      return null;
    }
  });

  // PREPARE 요청 결과 처리
  const prepareResults = await Promise.allSettled(preparePromises);
  // PREPARE 요청 결과 중 성공한 것들 중 약속을 받은 것들만 추출
  const promises = prepareResults
    .filter((r): r is PromiseFulfilledResult<PrepareResponse> =>
      r.status === 'fulfilled' && r.value !== null && !!r.value.promise
    )
    .map((r) => r.value);

  console.log(`[노드 ${nodeId}] Phase 1: 약속을 받은 노드 ${promises.length}/${allNodes.length} 개 (필요 ${majority} 개)`);

  if (promises.length < majority) {
    console.log(`[노드 ${nodeId}] ❌ Phase 1 실패: 약속을 받은 노드가 다수 필요 개수보다 적습니다. 리더 선출을 중단하고 다음 선거 시작.`);
    scheduleNextElection();
    return false;
  }

  // 수락된 값 중 가장 높은 값을 채택, 없으면 자신을 제안
  let proposedValue = nodeId;
  let highestAcceptedProposal = -1;

  for (const p of promises) {
    // 수락된 값 중 가장 높은 값을 채택
    if (p.acceptedProposal !== null && p.acceptedProposal > highestAcceptedProposal) {
      highestAcceptedProposal = p.acceptedProposal;
      proposedValue = p.acceptedValue!;
    }
  }

  if (highestAcceptedProposal > -1) {
    console.log(`[노드 ${nodeId}] Phase 1: 이전에 수락된 값 (노드 ${proposedValue}) 채택 / 제안 번호 #${highestAcceptedProposal}`);
  } else {
    console.log(`[노드 ${nodeId}] Phase 1: 이전에 수락된 값이 없으므로 자신을 제안 (노드 ${proposedValue})`);
  }

  // Phase 2: Accept
  console.log(`[노드 ${nodeId}] Phase 2: 모든 노드에 ACCEPT (값: 노드 ${proposedValue}) 요청 전송...`);

  const acceptPromises = allNodes.map(async (nodeUrl) => {
    // 자신을 포함한 모든 노드에 ACCEPT 요청 전송
    try {
      // ACCEPT 요청 전송
      const response = await fetch(`${nodeUrl}/paxos/accept`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ proposalNumber, value: proposedValue } as AcceptRequest),
        signal: AbortSignal.timeout(1000),
      });

      // ACCEPT 요청 실패 시 null 반환
      if (!response.ok) return null;
      return await response.json() as AcceptResponse;
    } catch (error) {
      // ACCEPT 요청 실패 시 null 반환
      return null;
    }
  });

  // ACCEPT 요청 결과 처리
  const acceptResults = await Promise.allSettled(acceptPromises);
  // ACCEPT 요청 결과 중 성공한 것들 중 수락된 것들만 추출
  const accepts = acceptResults
    .filter((r) => r.status === 'fulfilled' && r.value !== null && r.value.accepted)
    .length;

  console.log(`[노드 ${nodeId}] Phase 2: 수락된 노드 ${accepts}/${allNodes.length} 개 (필요 ${majority} 개)`);

  if (accepts >= majority) {
    console.log(`[노드 ${nodeId}] ✅ 리더 선출 성공! 노드 ${proposedValue} 가 리더가 되었습니다. (제안 번호 ${proposalNumber})`);

    if (proposedValue === nodeId) {
      console.log(`[노드 ${nodeId}] 👑 나는 리더!`);
      // 리더는 자신의 선출 타임아웃을 중지 (팔로워만 타임아웃을 감시해야 함)
      if (electionTimer) {
        clearTimeout(electionTimer);
        electionTimer = null;
      }
      startHeartbeat(proposalNumber);
    }

    return true;
  } else {
    console.log(`[노드 ${nodeId}] ❌ Phase 2 실패: 수락된 노드가 다수 필요 개수보다 적습니다. 리더 선출을 중단하고 다음 선거 시작.`);
    scheduleNextElection();
    return false;
  }
}

/**
 * 다음 리더 선출 예약
 */
function scheduleNextElection() {
  // 다음 선거 시작 전 랜덤 딜레이 적용
  const delay = 2000 + Math.random() * 3000;
  setTimeout(() => {
    // 현재 리더가 없으면 다음 선거 시작
    if (currentLeaderId === null) {
      runElection();
    }
  }, delay);
}

/**
 * 하트비트 시작
 * @param turn 현재 턴
 */
function startHeartbeat(turn: number) {
  // 하트비트 중지
  stopHeartbeat();

  heartbeatTimer = setInterval(() => {
    console.log(`[노드 ${nodeId}] 💓 하트비트 전송 (턴 ${turn})`);

    // 자신(리더)를 제외한 나머지 노드에 하트비트 전송
    peers.forEach(async (peerUrl) => {
      try {
        await fetch(`${peerUrl}/leader/heartbeat`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ leaderId: nodeId, turn } as HeartbeatMessage),
          signal: AbortSignal.timeout(1000),
        });
      } catch (error) {
        // 노드 다운 시 무시
      }
    });
  }, 1000);
}

/**
 * 하트비트 중지
 */
function stopHeartbeat() {
  if (heartbeatTimer) {
    clearInterval(heartbeatTimer);
    heartbeatTimer = null;
  }
}

/**
 * 리더 선출 타임아웃 리셋
 * 
 * 타임아웃 이후, 리더가 없으면 리더 선출 시작
 */
function resetElectionTimeout() {
  if (electionTimer) {
    clearTimeout(electionTimer);
  }

  // 리더 선출 타임아웃 랜덤 적용 (3-5초)
  const timeout = 3000 + Math.random() * 2000;

  electionTimer = setTimeout(() => {
    console.log(`[노드 ${nodeId}] ⏰ 하트비트 타임아웃 - 리더 없음 감지, 리더 선출 시작...`);
    currentLeaderId = null;
    // 현재 리더로 선출된 노드가 서비스 다운되는 경우에 해당 노드는 리더 선출에서 제외되어야 함
    // 따라서 수락된 상태를 리셋하여 진행하며, 각 리더 선출은 사실상 새로운 Paxos 인스턴스가 되도록 함
    acceptedProposal = null;
    acceptedValue = null;
    runElection();
  }, timeout);
}

// === Express Routes ===

const app = express();
app.use(express.json());

// Paxos Phase 1: Prepare
app.post('/paxos/prepare', (req, res) => {
  const { proposalNumber } = req.body as PrepareRequest;

  console.log(`[노드 ${nodeId}] PREPARE 요청 수신 / 제안 번호 #${proposalNumber}`);

  if (promisedProposal === null || proposalNumber > promisedProposal) {
    promisedProposal = proposalNumber;
    // 수락된 상태를 리셋하여 진행하며, 각 리더 선출은 사실상 새로운 Paxos 인스턴스가 되도록 함(죽은 리더 재선출 방지)
    acceptedProposal = null;
    acceptedValue = null;
    console.log(`[노드 ${nodeId}] → 제안 번호 #${proposalNumber} 약속`);

    res.json({
      promise: true,
      acceptedProposal: null,
      acceptedValue: null,
    } as PrepareResponse);
  } else {
    console.log(`[노드 ${nodeId}] → 이미 약속한 제안 번호 #${promisedProposal} 거절`);
    res.json({
      promise: false,
      acceptedProposal: null,
      acceptedValue: null,
    } as PrepareResponse);
  }
});

// Paxos Phase 2: ACCEPT 요청 처리
app.post('/paxos/accept', (req, res) => {
  const { proposalNumber, value } = req.body as AcceptRequest;

  console.log(`[노드 ${nodeId}] ACCEPT 요청 수신 / 제안 번호 #${proposalNumber}, 값: 노드 ${value}`);

  // 약속된 제안 번호가 없거나, 현재 제안 번호가 약속된 제안 번호보다 크거나 같으면 수락
  if (promisedProposal === null || proposalNumber >= promisedProposal) {
    // 약속된 제안 번호 갱신
    promisedProposal = proposalNumber;
    // 수락된 제안 번호 갱신
    acceptedProposal = proposalNumber;
    // 수락된 값 갱신(리더의 노드 아이디)
    acceptedValue = value;
    // 현재 리더 아이디 갱신
    currentLeaderId = value;
    // 현재 턴 갱신
    currentTurn = proposalNumber;

    console.log(`[노드 ${nodeId}] → 수락! 새로운 리더: 노드 ${value} (제안 번호 ${proposalNumber})`);

    // 리더가 아니라면 하트비트 중지
    if (value !== nodeId) {
      stopHeartbeat();
    }

    // 리더 선출 타임아웃 리셋
    resetElectionTimeout();

    res.json({ accepted: true } as AcceptResponse);
  } else {
    console.log(`[노드 ${nodeId}] → 이미 약속한 제안 번호 #${promisedProposal} 거절`);
    res.json({ accepted: false } as AcceptResponse);
  }
});

// 리더가 전송하는 하트비트 수신
app.post('/leader/heartbeat', (req, res) => {
  const { leaderId, turn } = req.body as HeartbeatMessage;

  // 새로운 턴이 시작되었다면,
  if (turn >= currentTurn) {
    // 새로운 리더가 등장했다면,
    if (leaderId !== currentLeaderId) {
      console.log(`[노드 ${nodeId}] 💓 새로운 리더로부터 하트비트 수신 / 리더: 노드 ${leaderId} (턴 ${turn})`);
    }

    // 현재 리더 아이디 갱신
    currentLeaderId = leaderId;
    // 현재 턴 갱신
    currentTurn = turn;
    
    /**
     * 리더 선출 타임아웃 리셋(타임아웃 이후, 리더가 없으면 리더 선출 시작)
     * 
     * 리더의 하트비트가 안정적으로 수신된다면, 리더 선출의 타임아웃은 계속 리셋된다. -> 현재 상태 유지.
     */
    resetElectionTimeout();
  }

  res.json({ ok: true });
});

// 상태 조회 엔드포인트
app.get('/status', (req, res) => {
  res.json({
    nodeId,
    currentLeaderId,
    currentTurn,
    promisedProposal,
    acceptedProposal,
    acceptedValue,
    isLeader: currentLeaderId === nodeId,
  });
});

// === 메인 함수 ===

app.listen(port, () => {
  console.log(`[노드 ${nodeId}] 🚀 포트 ${port}에서 수신 중...`);

  // 리더 선출 예약(랜덤 딜레이 적용)
  const delay = 1000 + Math.random() * 2000;
  console.log(`[노드 ${nodeId}] 리더 조회 전 ${Math.round(delay)}ms 대기...`);

  // 랜덤 딜레이 이후, 리더가 없으면 리더 선출 시작
  setTimeout(() => {
    if (currentLeaderId === null) {
      console.log(`[노드 ${nodeId}] 리더 없음, 리더 선출 시작...`);
      runElection();
    }
  }, delay);

  // 리더 선출 타임아웃 리셋(타임아웃 이후, 리더가 없으면 리더 선출 시작)
  resetElectionTimeout();
});
