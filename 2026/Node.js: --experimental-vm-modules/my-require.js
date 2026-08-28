const fs = require('fs');

function myRequire(filename) {
  // 1. 대상이 되는 파일을 텍스트로 읽어오기
  const yourCode = fs.readFileSync(filename, 'utf8');

  // 2. 앞, 뒤로 wrapper 코드 붙이기
  const wrapped =
    '(function (exports, require, module, __filename, __dirname) {' +
    yourCode +
    '\n})';

  // 3. 실행 가능한 함수로 변환
  const fn = eval(wrapped);

  // 4. 매개변수를 정의하고 호출하기
  const module = { exports: {} };
  // require 파라미터에 myRequire를 전달함!!!!!
  fn(module.exports, myRequire, module, filename, __dirname);

  // 5. module.exports 값을 리턴
  return module.exports;
}

console.log(myRequire('./greeting.js'));   // hello, world