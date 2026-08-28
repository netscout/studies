// 팩토리 패턴을 사용해서 import를 하나도 선언하지 않습니다.
// db 파라미터에는 실제 DB나 fake DB가 사용될 수 있죠.
export function createUserService({ db }) {
  return {
    getUser(id) {
      return db.query('SELECT * FROM users WHERE id = ?', id);
    },
  };
}