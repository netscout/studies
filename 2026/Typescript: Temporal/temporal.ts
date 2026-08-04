// Temporal은 전역(global)이므로 import가 필요 없습니다.
// 타입은 tsconfig의 lib: ["esnext"] 에서 옵니다.
const meeting: Temporal.ZonedDateTime =
  Temporal.ZonedDateTime.from('2026-03-08T09:00[America/New_York]');

console.log(meeting.toString());
console.log(`이 날은 ${meeting.hoursInDay}시간`);          // 23 — 서머타임 시작
console.log('서울에서는:', meeting.withTimeZone('Asia/Seoul').toPlainDateTime().toString());