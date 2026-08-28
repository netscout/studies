import vm from "node:vm";
import fs from "node:fs";
import path from "node:path";

// entryPath 모듈을 vm context 안에서 로드하고 실행합니다.
// mocks에 등록한 specifier는 파일 대신 가짜 모듈(SyntheticModule)로 바꿔치기
async function loadWithMocks(entryPath, mocks = {}) {
  const context = vm.createContext({ console });
  const cache = new Map(); // 같은 모듈을 두 번 만들지 않기 위한 캐시

  // specifier 하나를 모듈 객체로 변환하고
  // mock이면 SyntheticModule을 만들고, 아니면 파일을 읽어 SourceTextModule을 만듭니다.
  function loadModule(specifier, referencingPath) {
    if (Object.hasOwn(mocks, specifier)) {
      const key = `mock:${specifier}`;
      if (!cache.has(key)) {
        const exports = mocks[specifier];
        const mockModule = new vm.SyntheticModule(
          Object.keys(exports),
          // evaluate 시점에 mock 값을 export로 채워 넣습니다.
          function fillExports() {
            for (const [name, value] of Object.entries(exports)) {
              this.setExport(name, value);
            }
          },
          { context, identifier: key },
        );
        cache.set(key, mockModule);
      }
      return cache.get(key);
    }

    // mock이 아니면 진짜 파일이기 때문에, import 문을 적은 파일 위치 기준으로 경로를 설정합니다.
    const filePath = path.resolve(path.dirname(referencingPath), specifier);
    if (!cache.has(filePath)) {
      const source = fs.readFileSync(filePath, "utf8");
      // identifier에 절대 경로를 넣어두고,
      // 이 모듈이 또 다른 import를 만나면 이 값이 referencingPath가 됩니다.
      cache.set(filePath, new vm.SourceTextModule(source, { context, identifier: filePath }));
    }
    return cache.get(filePath);
  }

  // 의존성 그래프를 순회합니다.
  const visited = new Set(); // 순환 import를 만나도 무한 재귀에 빠지지 않게 방문한 모듈을 기록합니다.
  function resolveGraph(module) {
    if (visited.has(module)) return;
    visited.add(module);

    // SyntheticModule은 처음부터 'linked' 상태이고 moduleRequests도 없습니다.
    if (module.status !== "unlinked") return;

    const dependencies = module.moduleRequests.map(
      (request) => loadModule(request.specifier, module.identifier),
    );

    // linkRequests는 "이 모듈의 n번째 import는 이 모듈 객체다"라는 짝만 기록합니다.
    // status는 'unlinked' 상태이고, 아래 instantiate() 단계에서 'linked'로 바귑니다.
    module.linkRequests(dependencies);
    for (const dependency of dependencies) resolveGraph(dependency);
  }

  // entry 파일 자신을 referencingPath로 넘깁니다.
  // dirname(entry) + basename(entry) = entry
  const absoluteEntryPath = path.resolve(entryPath);
  const entry = loadModule(path.basename(absoluteEntryPath), absoluteEntryPath);

  resolveGraph(entry); //     1. 그래프의 모든 모듈을 만들고 서로 연결힙니다.
  entry.instantiate(); //     2. import/export 바인딩을 실제 메모리에 생성합니다.
  await entry.evaluate(); //  3. 코드를 실행합니다.
  return entry.namespace;
}

// 진짜 database.js 대신 쓸 가짜 db
const fakeDb = {
  query(_sql, id) {
    return { id, name: "홍길동" };
  },
};

const mocked = await loadWithMocks("./user-service.js", {
  "./database.js": { db: fakeDb },
});
console.log(mocked.getUser(42)); // { id: 42, name: '홍길동' }
