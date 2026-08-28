// 코드가 실행되기 전에 import로 인해 실행됩니다!
console.log('  [database.js] connecting to postgres://prod...');

export const db = {
  query(sql, id) {
    throw new Error('REAL DATABASE HIT');
  },
};