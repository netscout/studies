import { hash } from "./hash.js";

/**
 * 해시 링 위의 노드 위치
 */
export type RingPoint = { 
    point: bigint; 
    /**
     * 노드가 소속된 서버
     */
    server: string 
};

/**
 * 해시 링(consistent hashing)을 만든다.
 * 서버 이름에 번호를 붙여 해시한 값이 링 위의 점이 되고, 서버 하나가 점을 vnodes개씩 가지도록 한다.
 * 서버 목록이 바뀔 때마다 링을 새로 만들어야 한다.
 * @param servers 서버 목록
 * @param vnodes 서버 하나가 가지는 지점(virtual node) 수. 기본값 100.
 * @returns point 오름차순으로 정렬된 링. ringHash에 넘겨 재사용한다.
 */
export function makeRing(servers: string[], vnodes = 100): RingPoint[] {
    const ring: RingPoint[] = [];

    for (const server of servers) {
        for (let i = 0; i < vnodes; i++) {
            // 가상노드의 개수만큼 해시를 추가
            ring.push({ point: hash(`${server}#${i}`), server });
        }
    }

    // 이진 탐색을 위해 해시 위치를 기준으로 오름차순 정렬
    return ring.sort((a, b) => (a.point < b.point ? -1 : a.point > b.point ? 1 : 0));
}

/**
 * 해시 링(consistent hashing) 분배 알고리즘
 * 키를 해시한 지점에서 시계 방향으로 처음 만나는 서버를 고른다.
 * 서버가 추가되거나 제거되면 그 서버가 차지한 구간의 키만 이동한다.
 * @param key 유저ID 등의 키
 * @param ring makeRing으로 만든 링
 * @returns 분배 대상이 되는 서버 이름
 */
export function ringHash(key: string, ring: RingPoint[]): string {
    // 키를 해시 링 위의 한 점으로 변환
    const keyPoint = hash(key);

    // 이진 탐색을 위한 양쪽 끝 경계. keyPoint보다 크거나 같은 첫 지점을 찾는다.
    let low = 0;
    let high = ring.length;

    // 탐색 구간이 비게 되면 탐색 종료
    while (low < high) {
        // 탐색 중간 지점 계산
        const mid = Math.floor((low + high) / 2)

        // 중간 지점의 값이 keyPoint보다 작으면, mid이전에는 찾는 값이 없으므로 탐색 범위를 mid+1~high로 재설정
        if (ring[mid].point < keyPoint) {
            low = mid + 1;
        } else { // 중간 지점의 값이 keyPoint보다 더 크면, mid이전에 찾는 값이 있을 수도 있으므로, 탐색 범위를 low~mid로 재설정
            high = mid;
        }
    }

    // 마지막 지점보다 큰 keyPoint의 경우 low가 최대 ring.length까지 갈 수 있다.
    // 그런데 ring[ring.length]는 존재하지 않으므로, % 연산으로 처음부터 순환하도록 한다.(wrap-around)
    return ring[low % ring.length].server;
}