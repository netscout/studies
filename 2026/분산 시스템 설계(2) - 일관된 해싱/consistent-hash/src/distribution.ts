import { assign, compare, modHash, printDistribution } from "./hash.js";
import { rendezvousHash } from "./hrw.js";
import { makeRing, ringHash } from "./ring.js";

const users = Array.from(
    {length: 1000},
    (_, i) => `user-${String(i + 1)}`,
);

// 해시 알고리즘과 각 해시 알고리즘 실행을 위한 hashFn 구성
const algorithms: Record<string, (servers: string[]) => (key: string, servers: string[]) => string> = {
    "mod": () => modHash,
    "ring hash": (servers) => {
        const ring = makeRing(servers); // 링을 재사용하기 위해서 클로저로 캡쳐한다.

        return (key) => ringHash(key, ring);
    },
    "랑데뷰(HRW)": () => rendezvousHash,
};

// 각 해시 알고리즘 별로 실행하여 결과 확인
for (const [name, makeHashFn] of Object.entries(algorithms)) {
    console.log(`\n===== ${name} =====`);

    console.log("--- 서버 3대 ---");

    const servers = ["server-a", "server-b", "server-c"];

    const before = assign(users, servers, makeHashFn(servers));

    printDistribution(before);

    console.log("--- server-d 추가 ---");

    servers.push("server-d");

    const after = assign(users, servers, makeHashFn(servers));

    printDistribution(after);

    console.log(`이동된 사용자 수: ${compare(before, after)} / ${users.length}`);
}