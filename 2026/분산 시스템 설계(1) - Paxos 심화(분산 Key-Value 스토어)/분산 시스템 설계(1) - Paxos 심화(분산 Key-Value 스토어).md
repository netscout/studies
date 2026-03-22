# 분산 시스템 설계(1) - Paxos 심화(분산 Key-Value 스토어)

Paxos 알고리즘 관련 세번째 글입니다. 그러니까... 지난 글에서는 아주 간단하게 리더를 선출하는 예제를 만들어 봤었지만, 그게 뭐랄까 좀 예제를 위한 예제 같은 느낌이 들었습니다. Multi-Paxos를 제대로 구현하지도 못했고 말이죠. 그래서 Claude와 함께 고민을 하던 중에 Redis 같은 분산 Key-Value 스토어(이하 `dkv`)를 만들어 보는 게 어떨까 싶었습니다. 그래서 관련해서 튜토리얼을 생성해내고, 직접 쳐보기도 하면서 Claude Code와 예제를 완성하고 코드를 모두 한 줄씩 읽어가면서 동작을 이해하려 노력해봤습니다.

그 과정의 결과물을 [여기](https://github.com/netscout/dkv) 에 올려두었습니다.

## 간략한 소개

해당 리포지토리에 방문하시면 훨씬 더 상세한 README가 기다리고 있긴 합니다. 해당 내용은 초반에 틀을 제가 손으로 직접 잡고 나머지 부분을 Claude Code와 함께 완성 및 보강했습니다. 여기서는 어떻게 `dkv`가 동작하는지를 좀 설명드려보고자 합니다.

### Paxos Prepare-Accept를 통한 리더 선출

우선 3개의 노드를 실행하는 과정을 볼까요? 일단 첫 번째 노드부터 실행합니다.

```bash
❯ go run ./cmd/dkv -id=1
2026/03/17 14:16:53 gRPC listening on 127.0.0.1:9001
time=2026-03-17T14:16:53.495+09:00 level=INFO msg="wal: 재생 완료" node=1 lastApplied=1 totalRecords=42 acceptorStates=41 committedEntries=1
time=2026-03-17T14:16:53.495+09:00 level=INFO msg="paxos: 시작" node=1 peers="[1 2 3]" electionTimeout=500ms heartbeatInterval=100ms
time=2026-03-17T14:16:53.495+09:00 level=INFO msg="http: HTTP 서버 시작" node=1 addr=127.0.0.1:8001
Node 1 started — HTTP: 127.0.0.1:8001, gRPC: 127.0.0.1:9001
time=2026-03-17T14:16:54.253+09:00 level=INFO msg="election: 타이머 발동, 선거 시작" node=1 elapsed=2562047h47m16.854775807s threshold=500ms
time=2026-03-17T14:16:54.253+09:00 level=INFO msg="election: 🟢 prepare 시작" node=1 ballot=N1.1 prevNumber=0
time=2026-03-17T14:16:54.256+09:00 level=INFO msg="election: prepare 전송 실패" node=1 peer=2 ballot=N1.1 err="rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing: dial tcp 127.0.0.1:9002: connect: connection refused\""
time=2026-03-17T14:16:54.256+09:00 level=INFO msg="election: prepare 전송 실패" node=1 peer=3 ballot=N1.1 err="rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing: dial tcp 127.0.0.1:9003: connect: connection refused\""
time=2026-03-17T14:16:54.257+09:00 level=INFO msg="election: prepare 거부됨" node=1 peer=1 ballot=N1.1 promisedBallot=N30.1
time=2026-03-17T14:16:54.257+09:00 level=INFO msg="election: prepare 결과" node=1 promises=0 majority=2 ballot=N1.1 result=실패
time=2026-03-17T14:16:54.257+09:00 level=INFO msg="election: ballot fast-forward" node=1 from=1 to=30
time=2026-03-17T14:16:54.257+09:00 level=INFO msg="election: 실패" node=1 err="🔴election failed: 0/2 promises"
time=2026-03-17T14:16:54.988+09:00 level=INFO msg="election: 타이머 발동, 선거 시작" node=1 elapsed=2562047h47m16.854775807s threshold=500ms
time=2026-03-17T14:16:54.988+09:00 level=INFO msg="election: 🟢 prepare 시작" node=1 ballot=N31.1 prevNumber=30
time=2026-03-17T14:16:54.988+09:00 level=INFO msg="election: prepare 전송 실패" node=1 peer=2 ballot=N31.1 err="rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing: dial tcp 127.0.0.1:9002: connect: connection refused\""
time=2026-03-17T14:16:54.988+09:00 level=INFO msg="election: prepare 전송 실패" node=1 peer=3 ballot=N31.1 err="rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing: dial tcp 127.0.0.1:9003: connect: connection refused\""
time=2026-03-17T14:16:54.994+09:00 level=INFO msg="election: prepare 약속 받음" node=1 peer=1 ballot=N31.1
time=2026-03-17T14:16:54.994+09:00 level=INFO msg="election: prepare 결과" node=1 promises=1 majority=2 ballot=N31.1 result=실패
time=2026-03-17T14:16:54.994+09:00 level=INFO msg="election: 실패" node=1 err="🔴election failed: 1/2 promises"
...
```

뭔가 로그가 복잡해 보이지만, 리더 선출을 위한 선거(Prepare)를 진행하고 있습니다. 🟢과 🔴를 주목해서 보시면 선거가 시작됐지만 아직 자신 밖에 없으므로 선거는 실패하고 있습니다. 자 이제 2번 노드 투입합니다!

```bash
❯ go run ./cmd/dkv -id=2
2026/03/17 14:17:01 gRPC listening on 127.0.0.1:9002
time=2026-03-17T14:17:01.908+09:00 level=INFO msg="wal: 재생 완료" node=2 lastApplied=1 totalRecords=20 acceptorStates=19 committedEntries=1
time=2026-03-17T14:17:01.908+09:00 level=INFO msg="paxos: 시작" node=2 peers="[1 2 3]" electionTimeout=500ms heartbeatInterval=100ms
time=2026-03-17T14:17:01.908+09:00 level=INFO msg="http: HTTP 서버 시작" node=2 addr=127.0.0.1:8002
Node 2 started — HTTP: 127.0.0.1:8002, gRPC: 127.0.0.1:9002
time=2026-03-17T14:17:02.429+09:00 level=INFO msg="election: 타이머 발동, 선거 시작" node=2 elapsed=2562047h47m16.854775807s threshold=500ms
time=2026-03-17T14:17:02.429+09:00 level=INFO msg="election: 🟢 prepare 시작" node=2 ballot=N1.2 prevNumber=0
time=2026-03-17T14:17:02.431+09:00 level=INFO msg="election: prepare 전송 실패" node=2 peer=3 ballot=N1.2 err="rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing: dial tcp 127.0.0.1:9003: connect: connection refused\""
time=2026-03-17T14:17:02.431+09:00 level=INFO msg="election: prepare 거부됨" node=2 peer=1 ballot=N1.2 promisedBallot=N42.1
time=2026-03-17T14:17:02.432+09:00 level=INFO msg="election: prepare 거부됨" node=2 peer=2 ballot=N1.2 promisedBallot=N25.1
time=2026-03-17T14:17:02.432+09:00 level=INFO msg="election: prepare 결과" node=2 promises=0 majority=2 ballot=N1.2 result=실패
time=2026-03-17T14:17:02.432+09:00 level=INFO msg="election: ballot fast-forward" node=2 from=1 to=42
time=2026-03-17T14:17:02.432+09:00 level=INFO msg="election: 실패" node=2 err="🔴election failed: 0/2 promises"
time=2026-03-17T14:17:02.991+09:00 level=INFO msg="election: 타이머 발동, 선거 시작" node=2 elapsed=2562047h47m16.854775807s threshold=500ms
time=2026-03-17T14:17:02.991+09:00 level=INFO msg="election: 🟢 prepare 시작" node=2 ballot=N43.2 prevNumber=42
time=2026-03-17T14:17:02.991+09:00 level=INFO msg="election: prepare 전송 실패" node=2 peer=3 ballot=N43.2 err="rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing: dial tcp 127.0.0.1:9003: connect: connection refused\""
time=2026-03-17T14:17:02.996+09:00 level=INFO msg="election: prepare 약속 받음" node=2 peer=2 ballot=N43.2
time=2026-03-17T14:17:03.000+09:00 level=INFO msg="election: prepare 약속 받음" node=2 peer=1 ballot=N43.2
time=2026-03-17T14:17:03.000+09:00 level=INFO msg="election: prepare 결과" node=2 promises=2 majority=2 ballot=N43.2 result=성공
time=2026-03-17T14:17:03.000+09:00 level=INFO msg="election: 미커밋 엔트리 재제안" node=2 slot=1 ballot=N43.2 key=hahaha
time=2026-03-17T14:17:03.008+09:00 level=INFO msg="election: 🎉 리더 당선" node=2 ballot=N43.2 committedUpTo=1
```

마지막 줄을 보시면 2번 노드는 시작되자 마자 과반수의 지지를 받아 리더가 됩니다! 이때 1번 노드의 상황을 볼까요?

```bash
...
time=2026-03-17T14:17:02.997+09:00 level=INFO msg="election: 🟢 prepare 시작" node=1 ballot=N43.1 prevNumber=42
time=2026-03-17T14:17:02.997+09:00 level=INFO msg="election: prepare 전송 실패" node=1 peer=3 ballot=N43.1 err="rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing: dial tcp 127.0.0.1:9003: connect: connection refused\""
time=2026-03-17T14:17:02.997+09:00 level=INFO msg="election: prepare 전송 실패" node=1 peer=2 ballot=N43.1 err="rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing: dial tcp 127.0.0.1:9002: connect: connection refused\""
time=2026-03-17T14:17:03.000+09:00 level=INFO msg="election: prepare 거부됨" node=1 peer=1 ballot=N43.1 promisedBallot=N43.2
time=2026-03-17T14:17:03.000+09:00 level=INFO msg="election: prepare 결과" node=1 promises=0 majority=2 ballot=N43.1 result=실패
time=2026-03-17T14:17:03.000+09:00 level=INFO msg="election: 실패" node=1 err="🔴election failed: 0/2 promises"
time=2026-03-17T14:17:03.110+09:00 level=INFO msg="heartbeat: 새 리더 감지" node=1 leader=2 ballot=N43.2 prevLeader=0
...
```

자신이 리더가 되고자 하는 요청(Prepare)는 한 표도 받지 못했고, 2번 노드가 2표를 모두 약속받은 뒤 Accept까지 진행해버렸기 때문에 2번 노드가 새로운 리더임을 인정하게 됩니다.

이제 3번 노드가 또 합류합니다!

```bash
❯ go run ./cmd/dkv -id=3
2026/03/17 14:17:11 gRPC listening on 127.0.0.1:9003
time=2026-03-17T14:17:11.717+09:00 level=INFO msg="wal: 재생 완료" node=3 lastApplied=1 totalRecords=31 acceptorStates=30 committedEntries=1
time=2026-03-17T14:17:11.717+09:00 level=INFO msg="paxos: 시작" node=3 peers="[1 2 3]" electionTimeout=500ms heartbeatInterval=100ms
time=2026-03-17T14:17:11.717+09:00 level=INFO msg="http: HTTP 서버 시작" node=3 addr=127.0.0.1:8003
Node 3 started — HTTP: 127.0.0.1:8003, gRPC: 127.0.0.1:9003
time=2026-03-17T14:17:12.468+09:00 level=INFO msg="election: 타이머 발동, 선거 시작" node=3 elapsed=2562047h47m16.854775807s threshold=500ms
time=2026-03-17T14:17:12.469+09:00 level=INFO msg="election: 🟢 prepare 시작" node=3 ballot=N1.3 prevNumber=0
time=2026-03-17T14:17:12.471+09:00 level=INFO msg="election: prepare 거부됨" node=3 peer=2 ballot=N1.3 promisedBallot=N43.2
time=2026-03-17T14:17:12.471+09:00 level=INFO msg="election: prepare 거부됨" node=3 peer=3 ballot=N1.3 promisedBallot=N34.3
time=2026-03-17T14:17:12.471+09:00 level=INFO msg="election: prepare 거부됨" node=3 peer=1 ballot=N1.3 promisedBallot=N43.2
time=2026-03-17T14:17:12.471+09:00 level=INFO msg="election: prepare 결과" node=3 promises=0 majority=2 ballot=N1.3 result=실패
time=2026-03-17T14:17:12.471+09:00 level=INFO msg="election: ballot fast-forward" node=3 from=1 to=43
time=2026-03-17T14:17:12.471+09:00 level=INFO msg="election: 실패" node=3 err="🔴election failed: 0/2 promises"
time=2026-03-17T14:17:13.092+09:00 level=INFO msg="election: 타이머 발동, 선거 시작" node=3 elapsed=2562047h47m16.854775807s threshold=500ms
...
time=2026-03-17T14:17:16.468+09:00 level=INFO msg="election: 🟢 prepare 시작" node=3 ballot=N49.3 prevNumber=48
time=2026-03-17T14:17:16.469+09:00 level=INFO msg="election: prepare 거부됨" node=3 peer=2 ballot=N49.3 promisedBallot=N43.2
time=2026-03-17T14:17:16.469+09:00 level=INFO msg="election: prepare 거부됨" node=3 peer=1 ballot=N49.3 promisedBallot=N43.2
time=2026-03-17T14:17:16.472+09:00 level=INFO msg="election: prepare 약속 받음" node=3 peer=3 ballot=N49.3
time=2026-03-17T14:17:16.472+09:00 level=INFO msg="election: prepare 결과" node=3 promises=1 majority=2 ballot=N49.3 result=실패
time=2026-03-17T14:17:16.472+09:00 level=INFO msg="election: 실패" node=3 err="🔴election failed: 1/2 promises"
time=2026-03-17T14:17:17.109+09:00 level=INFO msg="heartbeat: 새 리더 감지" node=3 leader=2 ballot=N43.2 prevLeader=0
```

처음에는 자신을 리더로 제안(Propose)하지만, 결국 2번 노드가 리더임을 인정하게 됩니다.

### Paxos Accept를 통한 key-value 전파

3번 노드에 키를 `gundam`으로 하고 값을 `건담이말을건담`으로 넣어볼까요?

```bash
❯ curl -L -X PUT http://127.0.0.1:8003/kv/gundam -d "건담이말을건담"
OK (index=2)
```

참고로 `-L` 플래그는 요청이 리다이렉트 될 수 있도록 해줍니다. 현재 `dkv`는 리더가 아닌 노드에게 쓰기 요청을 하면 해당 요청을 리더에게 리다이렉트 하도록 구현되었거든요! 자 그럼 1번 노드에서 키를 조회해볼까요?

```bash
❯ curl http://127.0.0.1:8001/kv/gundam
건담이말을건담
```

3번 노드가 쓰기 요청을 리더인 2번 노드에게 전달하고, 리더는 다른 노드들에 Accept 요청을 전송해서 쓰기 요청이 다른 노드에 전파되도록 합니다. 그래서 1번 노드에서도 키로 값을 조회해올 수 있는 거죠.

> 리더 선출 과정에서 이미 Prepare를 거쳤기 때문에 바로 Accept단계로 넘어갑니다!

## 정리

이걸 구현하고 이해해보려고 노력하면서 왜 Paxos가 더 이상 쓰이지 않는지 이해했습니다. 이게 좀 많이 복잡하네요. 솔직히 아직도 좀 헷갈립니다. 다음 부터는 `dkv`를 만드는 과정에서 제가 새롭게 배운 내용들을 따로 한 번 정리해보려고 합니다. 그러고 나서 Raft를 추가로 구현해볼 생각입니다.