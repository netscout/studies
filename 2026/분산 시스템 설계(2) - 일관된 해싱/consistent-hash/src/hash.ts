import crypto from "node:crypto";

/**
 * 나머지 연산(mod) 분배 알고리즘
 * 서버 수가 바뀌면 거의 모든 키의 나머지 값이 바뀐다. 그래서 대부분의 키가 이동한다.
 * @param key 유저ID 등의 키
 * @param servers 서버 목록
 * @returns 분배 대상이 되는 서버 이름
 */
export function modHash(key: string, servers: string[]) {
    const index = Number(hash(key) % BigInt(servers.length));

    return servers[index];
}

/**
 * 노드의 기본 crypto 모듈을 사용하여 입력 값을 해시한다.
 * @param input 입력 값
 * @param algorithm 해시 알고리즘
 * @returns 해시 결과
 */
export function hash(input: string, algorithm = "sha256"): bigint {
    // 256비트 해시 생성 -> 64개의 16진수 문자열(64 * 4 = 256)
    const digest = crypto.createHash(algorithm).update(input).digest("hex");

    // 결과로 나온 16진수 문자열을 256비트 숫자로 변환
    // number는 double 형식이므로, 정수의 범위는 -(2⁵³ - 1) ~ (2⁵³ - 1) 이므로 256bit 정수를 정확하게 표현불가.
    // 그래서 BigInt를 사용함. BigInt의 경우 자바스크립트 엔진에서 단일 값에 허용하는 최대 비트 수까지 할당가능.
    // 정수 비교를 위해 BigInt를 썼지만, sha256 해시의 결과는 항상 같은 길이의 문자열이므로 문자열을 그대로 비교해도 같은 결과를 얻을 수 있다!
    return BigInt(`0x${digest}`);
}

/**
 * 각 키를 대응하는 서버에 라우팅하여 할당
 * @param keys 유저ID 등의 키 목록
 * @param servers 서버 목록
 * @param hashFn 해시 알고리즘(링, 랑데뷰)
 * @returns 
 */
export function assign(
    keys: string[], 
    servers: string[], 
    hashFn: (key: string, servers: string[]) => string): Record<string, string> {
    const result: Record<string, string> = {};

    for (const key of keys) {
        result[key] = hashFn(key, servers);
    }

    return result;
}

/**
 * 서버 목록에 변동이 발생하여 사용자를 재분배한 경우, 재분배로 인해 발생한 이동을 확인한다.
 * @param previous 이전 분배 결과
 * @param current 재분배 결과
 * @returns 재분배 결과 발생한 이동
 */
export function compare(previous: Record<string, string>, current: Record<string, string>) {
    let moved = 0;

    for (const key of Object.keys(previous)) {
        if (previous[key] !== current[key]) {
            moved++;
        }
    }

    return moved;
}

/**
 * 각 서버별 분배 결과를 테이블로 출력.
 */
export function printDistribution(result: Record<string, string>) {
    const distribution: Record<string, number> = {};

    for (const server of Object.values(result)) {
        distribution[server] = (distribution[server] ?? 0) + 1;
    }

    console.table(distribution);
}