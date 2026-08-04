process.env.TZ = "America/New_York";

// ---------------------------------------------------------------------------
console.log("1. 타임존은 '로컬'과 'UTC' 딱 두 개뿐이다");
// ---------------------------------------------------------------------------

const noon = new Date("2026-07-31T12:00:00Z");
console.log("로컬: getHours()", noon.getHours());
console.log("UTC: getUTCHours()", noon.getUTCHours());

const seoulHour = new Intl.DateTimeFormat("en-CA", {
  timeZone: "Asia/Seoul",
  hour: "2-digit",
  hourCycle: "h23",
}).format(noon);

// 결과값이 문자열이므로, 시간의 계산은 불가
console.log("서울: ", seoulHour);

console.log();

// ---------------------------------------------------------------------------
console.log("2. 파서(parser)를 신뢰할 수 없다");
// ---------------------------------------------------------------------------

for (const input of [
  '2026-07-31', //       날짜만 있는 ISO      → UTC로 해석
  '2026-07-31T09:00', // 시각이 붙은 ISO      → 로컬로 해석 (!)
  '2026-07-31 09:00', // T 대신 공백          → 명세 밖, 엔진 재량
  '07/31/2026', //       ISO가 아예 아님      → 명세 밖, 엔진 재량
  '2026-07-31T09:00Z', // 명시적 UTC          → UTC로 해석
]) {
  console.log(input, "-->", new Date(input).toISOString());
}

console.log();

// ---------------------------------------------------------------------------
console.log("3. Date는 변경 가능(mutable)하다");
// ---------------------------------------------------------------------------

function startOfMonth(date) {
  date.setDate(1); // 인자로 받은 객체를 직접 수정합니다
  return date;
}

const someDate = new Date('2026-07-31T12:00:00Z');
console.log("함수 호출 전", someDate.toISOString().slice(0, 10));
startOfMonth(someDate);
console.log("함수 호출 후", someDate.toISOString().slice(0, 10));

// ---------------------------------------------------------------------------
console.log("4. 서머타임(DST) 동작을 예측할 수 없다");
// ---------------------------------------------------------------------------

// 3월의 두번째 일요일 2am -> 3am으로 변경(-05:00 EST-> -04:00 EDT). 따라서 뉴욕기준 2026년 3월 8일 2am은 존재하지 않음.
// 따라서 2026년 3월 8일은 23시간
// 이후 11월 1일 2am에 다시 EDT->EST로 변경되며, 1am을 두번 통과하므로 11월 1일은 25시간
const beforeSpringForward = new Date('2026-03-08T00:00:00-05:00');
console.log("3월 8일 자정", beforeSpringForward);
// 다음 날(+ 86,400,000 ms) -> 2026-03-09T05:00:00.000Z 이므로, EDT 기준으로는 3월 9일 오전 1시.
const nextDay = new Date(beforeSpringForward.getTime() + 86400000);
console.log("다음날 자정이 아니다!", nextDay.toLocaleString('en-US', { timeZone: 'America/New_York' }));

console.log();

// ---------------------------------------------------------------------------
console.log("5. 계산 API가 조악하다");
// ---------------------------------------------------------------------------

// 1월 31일 + 1달 -> 2월 31일 -> 3월 3일
const monthEnd = new Date('2026-01-31T12:00:00Z');
monthEnd.setUTCMonth(monthEnd.getUTCMonth() + 1);
console.log("1월 31일 + 1달", monthEnd.toISOString().slice(0, 10));

// 두 날짜의 차이 구하기. 86,400,000 ms를 이용해서 나눠야 함. DST를 사용하는 경우 1년에 두 번 틀림
const rangeStart = new Date('2026-01-01T00:00:00Z');
const rangeEnd = new Date('2026-09-15T00:00:00Z');
console.log("두 날짜의 간격", (rangeEnd - rangeStart) / 86400000);

// 날짜의 각 요소 구하기, getMonth는 0부터 시작하고, getDay는 요일을 리턴 함.
const july31 = new Date("2026-07-31T12:00:00Z");
console.log("getMonth:", july31.getUTCMonth());
console.log("getDay:", july31.getUTCDay());
console.log("getFullYear:", july31.getUTCFullYear());

console.log();

// ---------------------------------------------------------------------------
console.log("6. 그레고리력 외의 달력을 지원하지 않는다");
// ---------------------------------------------------------------------------

// 표기는 가능하지만, 결과는 문자열 값이므로 1과 같은 이유로 날짜 계산 등은 불가능.
for (const calendar of ['hebrew', 'islamic-umalqura', 'japanese', 'buddhist']) {
  console.log(
    calendar,
    new Intl.DateTimeFormat(`en-u-ca-${calendar}`, {
      timeZone: 'UTC',
      dateStyle: 'long',
    }).format(july31)
  );
}