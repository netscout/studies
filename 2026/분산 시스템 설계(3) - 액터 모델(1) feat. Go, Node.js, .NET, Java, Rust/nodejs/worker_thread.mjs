import { Worker, isMainThread, parentPort, workerData } from 'node:worker_threads';

if (isMainThread) { // 메인 스레드에서 실행 할 코드
  const w = new Worker(new URL(import.meta.url), { workerData: 10000000000 }); // 워커 스레드 실행
  w.on('message', (sum) => console.log(sum));
} else { // 워커 스레드에서 실행할 코드
  let sum = 0;
  for (let i = 0; i < workerData; i++) {
    sum += i;
  }
  parentPort.postMessage(sum); // 메인 스레드에 결과 전달
}