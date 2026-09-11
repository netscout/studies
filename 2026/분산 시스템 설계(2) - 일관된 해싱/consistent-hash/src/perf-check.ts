import { hash } from "./hash.js";
import { xxHash } from "./hrw.js";

const users = Array.from(
    {length: 100000},
    (_, i) => `user-${String(i + 1)}`,
);

const ROUNDS = 5;

/**
 * 해시 함수를 사용자 10만 명에 대해 실행하고 걸린 시간을 잰다.
 * 첫 번째 측정은 JIT 예열(warmup)이므로 측정에서 빼고, ROUNDS만큼 돌린 평균을 출력한다.
 */
function bench(name: string, fn: (input: string) => bigint) {
    for (const user of users) {
        fn(user);
    }

    const start = performance.now();

    for (let round = 0; round < ROUNDS; round++) {
        for (const user of users) {
            fn(user);
        }
    }

    const elapsed = (performance.now() - start) / ROUNDS;

    console.log(`${name}: ${elapsed.toFixed(1)}ms`);
}

bench("SHA256", (user) => hash(user, "sha256"));
bench("XXHash64", (user) => xxHash(user));
