// 실행 방법:  node --test
// node:test는 vm 모듈을 사용하지 않기 때문에 --experimental-vm-modules 플래그가 필요 없습니다!

import { test } from 'node:test';
import assert from 'node:assert';

// fake DB를 만드는 함수
function createFakeDb(rows = {}) {
  const calls = [];
  return {
    calls,
    query(sql, id) {
      calls.push({ sql, id });
      return rows[id] ?? null;
    },
  };
}

test('fake DB로 서비스를 생성하기', async () => {
  const { createUserService } = await import('./user-service.js');
  const fakeDb = createFakeDb({ 42: { id: 42, name: '홍길동' } });

  const service = createUserService({ db: fakeDb });
  const user = service.getUser(42);

  assert.deepStrictEqual(user, { id: 42, name: '홍길동' });
  assert.deepStrictEqual(fakeDb.calls, [
    { sql: 'SELECT * FROM users WHERE id = ?', id: 42 },
  ]);
});

test('찾을 수 없는 데이터 테스트', async () => {
  const { createUserService } = await import('./user-service.js');
  const service = createUserService({ db: createFakeDb() });

  assert.strictEqual(service.getUser(999), null);
});

test('db 오류가 전파되는지 확인하는 테스트', async () => {
  const { createUserService } = await import('./user-service.js');
  const service = createUserService({
    db: { query() { throw new Error('connection reset'); } },
  });

  assert.throws(() => service.getUser(42), /connection reset/);
});
