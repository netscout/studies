import xxhash_addon from "xxhash-addon";
const { XXHash64 } = xxhash_addon;

/**
 * 고성능 non-cryptographic xxHash를 통한 해시를 진행한다.
 * @param input 입력값
 * @returns 해시결과
 */
export function xxHash(input: string): bigint {
    const digest = XXHash64.hash(Buffer.from(input));

    // 해시 결과(8바이트 Buffer)를 BigInt로 변환. number는 double 형식이고 표현가능한 정수의 범위는 -(2⁵³ - 1) ~ (2⁵³ - 1) 이므로 64비트를 표현 불가.
    // 16진수 문자열을 거치지 않고 Buffer에서 바로 읽는다. 문자열 변환과 파싱 비용을 줄이기 위해서다.
    return digest.readBigUInt64BE(0);
}

/**
 * 랑데뷰(HRW, Highest Random Weight) 분배 알고리즘
 * @param key 유저ID 등의 키
 * @param servers 서버 목록
 * @returns 분배 대상이 되는 서버 이름
 */
export function rendezvousHash(key: string, servers: string[]): string {
    let selected = "";
    let highestScore = -1n;

    for (const server of servers) {
        // 요청에 대한 각 서버별 해시 계산
        const score = xxHash(`${key}:${server}`);

        // 제일 높은 값의 서버를 선택
        if (score > highestScore) {
            highestScore = score;
            selected = server;
        }
    }

    return selected;
}
