/**
 * date-issues-fixed.mjs
 *
 * date-issues.js의 6개 섹션을 그대로 따라가면서, 각 문제를 Temporal로
 * 어떻게 해결하는지 보여줍니다.
 *
 * 실행:
 *   npm i @js-temporal/polyfill
 *   node date-issues-fixed.mjs
 *
 * 확장자가 .mjs인 이유: package.json에 "type": "module"이 없어도 Node가 ESM으로 인식하게 하기 위해서입니다.
 */

// 폴리필을 우선 사용하고, 없으면 런타임 내장 Temporal로 넘어갑니다.
let Temporal;
try {
  ({ Temporal } = await import("@js-temporal/polyfill"));
} catch {
  Temporal = globalThis.Temporal;
  if (!Temporal) {
    console.error("npm i @js-temporal/polyfill 을 먼저 실행해 주세요.");
    process.exit(1);
  }
}

// Temporal은 프로세스 타임존(TZ)에 의존하지 않습니다.
// 아래 process.env.TZ는 Date 쪽 비교 코드를 위해서만 남겨둡니다.
process.env.TZ = "America/New_York";

// ---------------------------------------------------------------------------
console.log("1. 타임존은 '로컬'과 'UTC' 딱 두 개뿐이다");
// ---------------------------------------------------------------------------
//
// Temporal의 경우 타임존이 값에 붙어 다닙니다.

// 특정 순간을 의미하는 Instant를 생성(Date는 항상 Instant 입니다)
const noon = Temporal.Instant.from("2026-07-31T12:00:00Z");

// 각 타임존으로 noon을 변환
for (const tz of ["America/New_York", "UTC", "Asia/Seoul"]) {
  const z = noon.toZonedDateTimeISO(tz);
  console.log(`${tz.padEnd(20)} ${z.hour}시  (offset ${z.offset})`);
}

// 문자열이 아니라 값이므로 그대로 계산에 쓸 수 있습니다.
const seoul = noon.toZonedDateTimeISO("Asia/Seoul");
console.log("서울 기준 3시간 뒤:", seoul.add({ hours: 3 }).hour + "시");

console.log();

// ---------------------------------------------------------------------------
console.log("2. 파서(parser)를 신뢰할 수 없다");
// ---------------------------------------------------------------------------
//
// 입력 형식이 강제되며, 해석이 불가능한 경우 RangeError를 던집니다. 
// "로컬로 해석할까 UTC로 해석할까"라는 문제 자체가 생기지 않습니다.

// 날짜만 있는 문자열 → PlainDate. 타임존이 아예 개입하지 않습니다.
console.log(`PlainDate.from("2026-07-31")`, Temporal.PlainDate.from("2026-07-31").toString());

// 시각까지 있지만 타임존이 없는 문자열 → PlainDateTime. 역시 타임존은 포함되지 않습니다.
console.log(`PlainDateTime.from("...T09:00")`, Temporal.PlainDateTime.from("2026-07-31T09:00").toString());

// 명확한 시점을 원하면 타임존을 명시해야 합니다. 생략하면 통과하지 않습니다.
console.log(`ZonedDateTime.from("...[NY]")`, Temporal.ZonedDateTime.from("2026-07-31T09:00[America/New_York]").toString());

// 추측을 요구하는 입력은 모두 거부됩니다.
for (const [desc, action] of [
  ["타임존 없이 Instant 요구", () => Temporal.Instant.from("2026-07-31T09:00")],
  ["타임존 없이 Zoned 요구", () => Temporal.ZonedDateTime.from("2026-07-31T09:00")],
  ["존재하지 않는 2월 30일", () => Temporal.PlainDate.from("2026-02-30")],
  ["ISO가 아닌 07/31/2026", () => Temporal.PlainDate.from("07/31/2026")],
  ["banana", () => Temporal.PlainDate.from("banana")],
]) {
  try {
    console.log(desc, "→", action().toString(), "(통과)");
  } catch (e) {
    console.log(desc, "→", e.constructor.name, "즉시 예외");
  }
}

// Date와의 결정적 차이: 파싱 실패가 "Invalid Date"라는 정상 처리 흐름으로 넘어가는 대신에 파싱한 그 자리에서 에러를 냅니다.

console.log();

// ---------------------------------------------------------------------------
console.log("3. Date는 변경 가능(mutable)하다");
// ---------------------------------------------------------------------------
//
// 모든 Temporal 객체는 불변(immutable)입니다. 값을 변경할 수 있는 setter가 없고, with(), add() 같은 메서드는 항상 새 객체를 반환합니다.

function startOfMonth(date) {
  return date.with({ day: 1 }); // 원본을 건드리지 않고 새 값을 반환
}

const someDate = Temporal.PlainDate.from("2026-07-31");
console.log("함수 호출 전", someDate.toString());
const firstDay = startOfMonth(someDate);
console.log("함수 호출 후", someDate.toString(), "  ← 원본 그대로");
console.log("반환값", firstDay.toString());

console.log();

// ---------------------------------------------------------------------------
console.log("4. 서머타임(DST) 동작을 예측할 수 없다");
// ---------------------------------------------------------------------------
//
// days/weeks/months → DST를 적용하여 계산
// hours 이하 → Date와 동일하게 실제 시간으로 계산

const midnight = Temporal.ZonedDateTime.from("2026-03-08T00:00[America/New_York]");

console.log("3월 8일 자정", midnight.toString());
console.log("  .add({ days: 1 })  ", midnight.add({ days: 1 }).toString(), " ← 진짜 다음 날 자정");
console.log("  .add({ hours: 24 })", midnight.add({ hours: 24 }).toString(), " ← 24시간 뒤 (오전 1시)");

// 하루가 몇 시간인지 직접 물어볼 수 있습니다.
for (const d of ["2026-03-07", "2026-03-08", "2026-11-01"]) {
  const z = Temporal.PlainDate.from(d).toZonedDateTime({ timeZone: "America/New_York" });
  console.log(`  ${d} → ${z.hoursInDay}시간`);
}

// 존재하지 않는 시각을 예약하려 하면 조용히 옮기지 않고 거부할 수 있습니다.
for (const option of ["compatible", "reject"]) {
  try {
    const r = Temporal.PlainDateTime.from("2026-03-08T02:30")
      .toZonedDateTime("America/New_York", { disambiguation: option });
    console.log(`  02:30 (${option})`.padEnd(24), r.toPlainTime().toString(), "← 조용히 이동함");
  } catch (e) {
    console.log(`  02:30 (${option})`.padEnd(24), e.constructor.name, "← 없는 시각이라고 알려줌");
  }
}

// 11월 1일 01:30은 두 번 존재합니다. 어느 쪽인지 고를 수 있습니다.
for (const option of ["earlier", "later"]) {
  const r = Temporal.PlainDateTime.from("2026-11-01T01:30")
    .toZonedDateTime("America/New_York", { disambiguation: option });
  console.log(`  01:30 (${option})`.padEnd(24), r.toInstant().toString());
}

console.log();

// ---------------------------------------------------------------------------
console.log("5. 계산 API가 조악하다");
// ---------------------------------------------------------------------------
//
// add/subtract/until/since가 있고, 필드는 1부터 시작하며, 이름이 우리가 생각하는 개념과 일치합니다.

// 1월 31일 + 1달 → 2월 28일로 잘라냅니다(clamp). 3월로 넘어가지 않습니다.
const endMonth = Temporal.PlainDate.from("2026-01-31");
console.log("1월 31일 + 1달      ", endMonth.add({ months: 1 }).toString());
console.log(`  overflow:"reject" `, (() => {
  try { return endMonth.add({ months: 1 }, { overflow: "reject" }).toString(); }
  catch (e) { return e.constructor.name + " (넘침을 알려달라고 할 수도 있음)"; }
})());

// 주의: 한 달씩 두 번 더하는 것과 두 달을 한 번에 더하는 것은 다릅니다.
// ex: 매월 구독 결제일은 최초 결제일에서 누적된 달만큼 한 번에 계산해야 합니다.
console.log("  1달+1달           ", endMonth.add({ months: 1 }).add({ months: 1 }).toString());
console.log("  2달 한 번에       ", endMonth.add({ months: 2 }).toString());

// 두 날짜의 간격. 86,400,000 같은 상수가 필요 없습니다.
const start = Temporal.PlainDate.from("2026-01-01");
const end = Temporal.PlainDate.from("2026-09-15");
console.log("두 날짜의 간격(일)  ", start.until(end).total({ unit: "day" }));
console.log("  ISO 8601 표기     ", start.until(end, { largestUnit: "month" }).toString());

// 필드 접근. 메서드가 아니라 속성이고, 월은 1부터 시작합니다.
const july31 = Temporal.PlainDate.from("2026-07-31");
console.log("month (1부터)       ", july31.month);
console.log("day   (일)          ", july31.day);
console.log("dayOfWeek (요일)    ", july31.dayOfWeek, "(1=월요일)");
console.log("year                ", july31.year);
console.log("daysInMonth         ", july31.daysInMonth);

console.log();

// ---------------------------------------------------------------------------
console.log("6. 그레고리력 외의 달력을 지원하지 않는다");
// ---------------------------------------------------------------------------
//
// 문자열이 아닌 날짜 값이기 때문에 표기 뿐만 아니라 계산이 가능합니다.

for (const calendar of ['hebrew', 'islamic-umalqura', 'japanese', 'buddhist']) {
  const d = july31.withCalendar(calendar);
  const year = d.era ? `${d.era} ${d.eraYear}년 (연속 ${d.year})` : `${d.year}년`;
  console.log(calendar, year);
}

// 문자열이 아니기 때문에 계산이 가능합니다.
const japaneseJuly31 = july31.withCalendar("japanese");
console.log(japaneseJuly31.add({ days: 2}).toString());

console.log();