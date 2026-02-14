import { spawn, ChildProcess } from 'child_process';
import * as readline from 'readline';
import { fileURLToPath } from 'url';
import { dirname } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

type NodeConfig = {
  id: number;
  port: number;
};

// 각 노드 아이디와 포트 번호 정의
const NODES: NodeConfig[] = [
  { id: 1, port: 3001 },
  { id: 2, port: 3002 },
  { id: 3, port: 3003 }
];

// 노드 프로세스 관리를 위한 맵 정의
const processes = new Map<number, ChildProcess>();

// 가독성을 위해 각 노드의 아이디별 출력 색상 정의
const colors = {
  1: '\x1b[36m', // 파란색
  2: '\x1b[33m', // 노란색
  3: '\x1b[35m', // 핑크색
  reset: '\x1b[0m', // 기본 색상
};

/**
 * 노드 프로세스를 시작하는 함수
 * @param nodeConfig - 노드 아이디와 포트 번호 정의
 */
function startNode(nodeConfig: NodeConfig) {
  // 현재 노드가 통신해야 할 다른 노드의 URL 목록 생성
  const peers: string[] = NODES
    .filter((n: NodeConfig) => n.id !== nodeConfig.id)
    .map((n: NodeConfig) => `http://localhost:${n.port}`);

  console.log(`${colors[nodeConfig.id as keyof typeof colors]}노드 ${nodeConfig.id} 시작...${colors.reset}`);

  // 노드 프로세스 생성
  const child = spawn(
    'npx',
    ['tsx', 'node.ts', String(nodeConfig.id), String(nodeConfig.port), ...peers],
    {
      cwd: __dirname,
      stdio: ['ignore', 'pipe', 'pipe'],
    }
  );

  // 노드 프로세스의 출력에 색상을 적용 하여 출력
  child.stdout?.on('data', (data: Buffer) => {
    const color = colors[nodeConfig.id as keyof typeof colors] || colors.reset;
    process.stdout.write(`${color}${data.toString()}${colors.reset}`);
  });

  // 노드 프로세스의 에러 출력에 색상을 적용 하여 출력
  child.stderr?.on('data', (data: Buffer) => {
    const color = colors[nodeConfig.id as keyof typeof colors] || colors.reset;
    process.stderr.write(`${color}[노드 ${nodeConfig.id} 오류] ${data.toString()}${colors.reset}`);
  });

  // 노드 프로세스가 종료될 때 처리
  child.on('exit', (code, signal) => {
    console.log(`${colors[nodeConfig.id as keyof typeof colors]}[노드 ${nodeConfig.id}] 프로세스 종료 (code: ${code}, signal: ${signal})${colors.reset}`);
    processes.delete(nodeConfig.id);
  });

  // 노드 프로세스를 관리하는 맵에 추가
  processes.set(nodeConfig.id, child);
}

/**
 * 노드 프로세스 종료
 * @param id - 종료할 노드의 아이디
 */
function killNode(id: number) {
  const child = processes.get(id);

  // 노드 프로세스가 존재하면 종료
  if (child) {
    child.kill('SIGTERM');
    console.log(`\n${colors[id as keyof typeof colors]}[클러스터] 노드 ${id} 종료${colors.reset}\n`);
  } else {
    console.log(`\n[클러스터] 노드 ${id} 실행 중이 아님\n`);
  }
}

/**
 * 클러스터 상태 출력
 */
async function printStatus() {
  console.log('\n=== 클러스터 상태 ===');

  for (const node of NODES) {
    try {
      const res = await fetch(`http://localhost:${node.port}/status`, {
        signal: AbortSignal.timeout(1000),
      });
      const data = await res.json();
      const color = colors[node.id as keyof typeof colors] || colors.reset;
      const leaderMark = data.isLeader ? '👑' : '  ';

      // 노드 상태 출력
      console.log(
        `${color}${leaderMark} 노드 ${node.id}: 리더=${data.currentLeaderId ?? '없음'}, ` +
        `턴=${data.currentTurn}, 리더=${data.isLeader}, ` +
        `약속한 제안 번호=${data.promisedProposal ?? '없음'}, 수락한 제안 번호=${data.acceptedProposal ?? '없음'}${colors.reset}`
      );
    } catch {
      const color = colors[node.id as keyof typeof colors] || colors.reset;
      console.log(`${color}   노드 ${node.id}: ❌ 다운${colors.reset}`);
    }
  }

  console.log('');
}

/**
 * 대화형 CLI 구현
 */

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
});

// 모든 노드 시작
console.log('\n🚀 Starting Paxos Leader Election Cluster...\n');
for (const node of NODES) {
  startNode(node);
}

// 모든 노드가 시작될 때까지 대기 후 명령어 출력
setTimeout(() => {
  console.log('\n' + '='.repeat(60));
  console.log('📋 사용 가능한 명령어:');
  console.log('  kill <id>   - 노드 종료 (특정 노드에 장애가 발생하여 종료되는 걸 시뮬레이션)');
  console.log('  start <id>  - 종료된 노드 재시작');
  console.log('  status      - 클러스터 상태 출력');
  console.log('  exit        - 모든 노드 종료 및 종료');
  console.log('='.repeat(60) + '\n');
}, 1000);

// 대화형 CLI 명령어 처리
rl.on('line', async (line: string) => {
  const [cmd, arg] = line.trim().split(' ');

  // 명령어 처리
  switch (cmd) {
    case 'kill':
      const killId = parseInt(arg);
      if (NODES.find((n) => n.id === killId)) {
        killNode(killId);
      } else {
        console.log(`잘못된 노드 ID: ${arg}\n`);
      }
      break;

    case 'start':
      const startId = parseInt(arg);
      const node = NODES.find((n) => n.id === startId);
      if (node) {
        if (processes.has(startId)) {
          console.log(`\n노드 ${startId}: 이미 실행 중\n`);
        } else {
          startNode(node);
        }
      } else {
        console.log(`잘못된 노드 ID: ${arg}\n`);
      }
      break;

    case 'status':
      await printStatus();
      break;

    case 'exit':
      console.log('\n🛑 모든 노드 종료...\n');
      processes.forEach((p) => p.kill('SIGTERM'));
      setTimeout(() => process.exit(0), 500);
      break;

    case 'help':
      console.log('\n📋 사용 가능한 명령어:');
      console.log('  kill <id>   - 노드 종료');
      console.log('  start <id>  - 종료된 노드 재시작');
      console.log('  status      - 클러스터 상태 출력');
      console.log('  exit        - 모든 노드 종료 및 종료\n');
      break;

    default:
      if (line.trim()) {
        console.log(`알 수 없는 명령어: ${cmd}. 'help' 명령어를 사용하여 사용 가능한 명령어를 확인하세요.\n`);
      }
  }
});

// Ctrl+C 처리
process.on('SIGINT', () => {
  console.log('\n\n🛑 모든 노드 종료...\n');
  processes.forEach((p) => p.kill('SIGTERM'));
  setTimeout(() => process.exit(0), 500);
});
