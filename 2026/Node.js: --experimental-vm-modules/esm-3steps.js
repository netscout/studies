import vm from 'node:vm';

// vm.createContext는 독립적인 코드 실행공간(샌드박스)을 생성합니다.
// 새로 생성된 context에는 console 조차도 없기 때문에 주입해줘야 합니다.
const context = vm.createContext({ console });

// 1단계 - 분석. 코드를 컴파일하면서 import와 export를 추출합니다.
// 아직 코드는 실행되지 않습니다.
const module = new vm.SourceTextModule(
  `export const greeting = 'hello world';`,
  { context, identifier: 'hello.mjs' },
);
console.log(module.status);            // 'unlinked'

// 2단계 - 연결. 모든 import, export에서 로드할 모듈을 결정합니다.
// hello.mjs는 다른 모듈을 import하지 않기 때문에, linker는 아무 일도 하지 않습니다.
await module.link(() => { throw new Error('no imports'); });
console.log(module.status);            // 'linked'

// 3단계 - 실행. 이제 코드가 실행됩니다.
await module.evaluate();
console.log(module.status);            // 'evaluated'

console.log(module.namespace.greeting); // 'hello world'