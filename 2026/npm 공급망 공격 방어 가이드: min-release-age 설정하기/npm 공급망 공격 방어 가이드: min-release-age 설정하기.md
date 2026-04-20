# npm 공급망 공격 방어 가이드: `min-release-age` 설정하기

> 이 문서는 CLAUDE를 통해 작성되었습니다.

> **목적**: 이 가이드는 npm 생태계에 대한 **공급망 공격(Supply Chain Attack)** 으로부터 개발자와 조직을 보호하기 위한 실무 설정 가이드입니다. macOS, Windows 모두에서 동일하게 적용할 수 있도록 작성되었습니다.

---

## 배경: 왜 이 설정이 필요한가 — axios 사태 (2026년 3월 31일)

2026년 3월 31일, 주당 약 **1억 회** 다운로드되는 인기 HTTP 클라이언트 라이브러리 **axios** 의 메인테이너 npm 계정이 탈취되어, 악성 버전 `axios@1.14.1`과 `axios@0.30.4`가 npm 레지스트리에 공개되었습니다.

**공격 요약**

- 공격자는 사회공학 기법(위장된 회사, 가짜 Slack/Teams 미팅)으로 메인테이너 계정을 탈취했습니다.
- 악성 버전은 `plain-crypto-js@4.2.1`이라는 별도의 악성 패키지를 의존성으로 주입했습니다.
- 해당 패키지의 `postinstall` 훅은 `npm install` 실행 즉시 macOS / Windows / Linux용 RAT(원격 제어 트로이목마)을 자동으로 다운로드하여 실행했습니다.
- 코드를 한 줄도 import 하지 않고, **단지 설치만 해도** 감염되었습니다.
- 악성 버전은 공개 후 **약 3시간** 뒤에 npm 레지스트리에서 제거되었습니다.

**핵심 교훈**

> 만약 새 패키지 버전의 설치를 **단 하루만** 지연시켰다면, 대부분의 개발자는 이 공격에서 안전했을 것입니다.

대부분의 악성 npm 패키지는 보안 연구자들에 의해 수 시간 내에 탐지·제거됩니다. 따라서 갓 출시된 버전의 설치를 일정 기간 지연시키는 "쿨다운(cooldown)" 전략이 가장 단순하고 강력한 공급망 공격 방어 수단 중 하나입니다.

npm CLI는 **11.10.0 버전부터** 이 기능을 `min-release-age` 설정으로 제공합니다.

---

## 전체 절차 개요

1. Node.js 및 npm 버전 확인
2. Node.js를 24.x로 업그레이드 (필요 시)
3. npm을 11.10.0 이상으로 업그레이드
4. `min-release-age` 설정
5. 설정이 실제로 적용되는지 검증
6. 추가 보안 설정 (선택)

---

## 1단계: 현재 버전 확인

터미널(macOS: Terminal / Windows: PowerShell 또는 CMD)을 열고 다음을 실행합니다.

```bash
node -v
npm -v
```

**요구 버전**

- Node.js: **v24.x 이상** (npm 11은 공식적으로 Node 24와 함께 제공됨)
- npm: **11.10.0 이상**

두 버전 모두 충족한다면 [3단계](#3단계-min-release-age-설정)로 건너뛸 수 있습니다.

---

## 2단계: Node.js 설치 / 업그레이드

Node 버전을 유연하게 관리하기 위해 **nvm(Node Version Manager)** 사용을 권장합니다.

### macOS / Linux — nvm

**nvm 설치** (이미 설치되어 있다면 생략):

```bash
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.1/install.sh | bash
```

설치 후 **새 터미널을 엽니다**. 그 다음:

```bash
# 최신 Node 24.x 설치 및 기본값으로 지정
nvm install 24
nvm alias default 24
```

**이미 Node 24.x가 설치된 경우, 최신 패치 버전으로 업데이트**:

```bash
nvm install 24 --reinstall-packages-from=current
nvm alias default 24
```

`--reinstall-packages-from=current`는 기존에 전역 설치한 패키지들을 새 버전으로 자동 이전합니다.

### Windows — nvm-windows

Windows는 별도의 프로젝트인 **nvm-windows**를 사용합니다.

1. <https://github.com/coreybutler/nvm-windows/releases> 에서 `nvm-setup.exe` 다운로드 후 설치
2. **관리자 권한으로** PowerShell 또는 CMD 실행

```powershell
# 최신 Node 24.x 설치
nvm install 24

# 설치된 24.x 중 하나를 활성화
nvm use 24
```

**Windows에서 최신 패치 버전으로 업데이트**:

```powershell
nvm install 24
nvm use 24
# (선택) 이전 버전 제거
nvm uninstall <이전버전>
```

### 설치 확인

```bash
node -v   # v24.x.x 가 출력되어야 함
```

---

## 3단계: npm 11.10.0 이상으로 업그레이드

Node 24를 설치해도 번들된 npm이 11.10.0보다 낮을 수 있습니다.

```bash
npm -v
```

11.10.0 미만이라면 업그레이드:

```bash
npm install -g npm@latest
```

> **nvm 사용자 참고**: 이 명령은 `-g`로 설치하지만, nvm이 각 Node 버전마다 독립된 전역 디렉터리를 두기 때문에 **sudo / 관리자 권한 없이** 실행할 수 있습니다. 단, 이렇게 업그레이드한 npm은 **현재 활성화된 Node 버전에만** 적용됩니다. 다른 Node 버전으로 전환(`nvm use`)하면 해당 버전에 번들된 npm으로 되돌아갑니다.

확인:

```bash
npm -v   # 11.10.0 이상이어야 함
```

---

## 4단계: `min-release-age` 설정

사용자 수준 `.npmrc` 파일에 설정을 추가합니다. 이 파일의 위치는 OS별로 다릅니다.

- **macOS / Linux**: `~/.npmrc`
- **Windows**: `C:\Users\<사용자명>\.npmrc`

**OS 공통 명령**으로 설정:

```bash
npm config set min-release-age 7
```

`7`은 **일(day)** 단위입니다 (pnpm은 분 단위이니 혼동 주의). 권장값:

| 값 | 설명 | 적합한 환경 |
|----|------|-------------|
| `1` | 최소 방어. 대부분의 악성 패키지는 수 시간 내 제거됨 | 빠른 업데이트가 필요한 개인 프로젝트 |
| `3` | 균형 잡힌 선택. Renovate Bot의 권장 기본값 | 일반 팀/프로젝트 |
| `7` | 견고한 방어. 일반적으로 추천 | 대부분의 팀에 적합 |
| `14`+ | 매우 보수적 | 엔터프라이즈 / 금융권 |

### 파일 내용 확인

**macOS / Linux:**
```bash
cat ~/.npmrc
```

**Windows (PowerShell):**
```powershell
Get-Content $HOME\.npmrc
```

출력에 `min-release-age=7`이 포함되어 있으면 정상입니다.

---

## 5단계: 설정이 실제로 작동하는지 검증

> ⚠️ **알려진 버그 주의**: 현재 npm 버전(~11.12.x)에서는 `npm config get min-release-age`가 `null`을 반환하는 표시 버그([npm/cli#9199](https://github.com/npm/cli/issues/9199))가 있습니다. **실제 기능은 정상 동작**합니다. 따라서 `config get` 대신 아래의 **실제 설치 시험**으로 검증하세요.

### 최근 7일 이내에 배포된 버전을 찾아 설치 시도

```bash
# 최근에 자주 배포되는 패키지의 버전 히스토리 확인
npm view @types/node time --json
```

출력에서 **최근 7일 이내 날짜**의 버전을 하나 고릅니다. 그 다음 임시 폴더에서:

**macOS / Linux:**
```bash
mkdir -p /tmp/npm-test && cd /tmp/npm-test
npm install @types/node@<최근버전> --dry-run
```

**Windows (PowerShell):**
```powershell
mkdir $env:TEMP\npm-test; cd $env:TEMP\npm-test
npm install @types/node@<최근버전> --dry-run
```

### 성공 시 출력 예시

```
npm error code ETARGET
npm error notarget No matching version found for @types/node@X.X.X
  with a date before 2026/M/D, HH:MM:SS.
```

"`with a date before ...`" 의 날짜가 **오늘로부터 설정한 일수 이전**이면 정상 동작 중입니다. 설정된 기간보다 최근에 배포된 버전은 설치가 거부됩니다.

---

## 긴급 상황 대응: 일회성으로 제한 우회

긴급 CVE 패치가 출시되어 최신 버전을 즉시 설치해야 하는 경우, 한 번만 제한을 우회할 수 있습니다:

```bash
npm install <패키지명> --min-release-age=0
```

전역 `.npmrc` 설정은 유지되며, 이 명령에서만 제한이 무시됩니다.

---

## 주의사항 및 한계

1. **`npm ci`는 영향을 받지 않습니다.**
   `min-release-age`는 `npm install`의 의존성 해결 과정에 적용됩니다. `package-lock.json`이 이미 있고 `npm ci`로 설치하는 경우(주로 CI 환경), 락파일에 고정된 버전을 설치하므로 이 설정과 무관합니다. → **개발자의 로컬 머신에서 락파일에 악성 버전이 고정되는 것을 막는 것이 핵심 방어선**입니다.

2. **`~` 버전 범위 관련 버그.**
   의존성에 `~1.2.3`처럼 `~` 범위 지정자가 포함된 경우 min-release-age가 에러를 일으킬 수 있습니다 ([npm/cli#9005](https://github.com/npm/cli/issues/9005)). 수정되기 전까지는 `^`를 사용하는 것을 권장합니다.

3. **숫자만 허용.**
   `7d`, `7days` 같은 문자 접미사 형식은 지원하지 않습니다. 오직 **정수(일수)** 만 유효한 값입니다.

4. **예외 목록 미지원.**
   pnpm과 달리 현재 npm은 "이 패키지만 예외" 같은 화이트리스트 기능이 없습니다([npm/cli#8979](https://github.com/npm/cli/issues/8979)). 긴급한 경우 위의 `--min-release-age=0` 플래그로 우회하세요.

---

## 추가 보안 권장 사항 (axios 공격에서 배운 교훈)

`min-release-age`는 공급망 공격 방어의 **첫 번째 방어선**일 뿐입니다. 다음 조치들을 함께 적용하면 훨씬 안전합니다.

### (1) `package-lock.json`을 반드시 Git에 커밋

락파일이 있어야 `npm ci`로 재현 가능한 설치가 가능하고, 의도치 않은 버전 업그레이드를 막을 수 있습니다.

### (2) CI에서는 `npm install` 대신 `npm ci` 사용

```bash
npm ci
```

락파일에 고정된 버전만 설치하므로, 악성 패키지가 떠도 CI가 자동으로 업그레이드하지 않습니다.

### (3) 프로덕션 빌드에서 lifecycle script 비활성화 고려

```bash
npm config set ignore-scripts true
```

`postinstall` 같은 훅이 실행되지 않아 axios 사례와 같은 공격 대부분을 원천 차단합니다. 단, 일부 정상 패키지(네이티브 바인딩 포함)는 빌드에 실패할 수 있으므로 **CI에서 먼저 검증** 후 적용하세요.

### (4) axios 감염 여부 직접 점검

현재 프로젝트에 악성 버전이 혹시 들어와 있지 않은지 확인:

```bash
npm ls axios
npm ls plain-crypto-js
```

다음 중 하나라도 발견되면 **즉시 감염으로 간주**하고 조치하세요:

- `axios@1.14.1`
- `axios@0.30.4`
- `plain-crypto-js`(모든 버전)

**조치 절차**

1. 해당 머신의 **모든 자격증명 즉시 회전(rotate)**:
   API 키, SSH 키, GitHub/npm 토큰, 클라우드 크리덴셜 등
2. 캐시 삭제:
   ```bash
   npm cache clean --force
   ```
3. 안전한 버전으로 다운그레이드 (`axios@1.14.0` 또는 `axios@0.30.3`)
4. CI/CD 로그에서 2026년 3월 31일 00:21 UTC ~ 03:15 UTC 사이 `npm install` 실행 이력 검토

---

## 참고 자료

- [npm 공식 문서: min-release-age](https://docs.npmjs.com/cli/v11/using-npm/config#min-release-age)
- [axios 공격 Post Mortem (GitHub Issue #10636)](https://github.com/axios/axios/issues/10636)
- [Microsoft Security: Mitigating the Axios npm supply chain compromise](https://www.microsoft.com/en-us/security/blog/2026/04/01/mitigating-the-axios-npm-supply-chain-compromise/)
- [Socket Blog: npm Introduces minimumReleaseAge](https://socket.dev/blog/npm-introduces-minimumreleaseage-and-bulk-oidc-configuration)
- [nvm (macOS / Linux)](https://github.com/nvm-sh/nvm)
- [nvm-windows](https://github.com/coreybutler/nvm-windows)

---

*최종 업데이트: 2026년 4월 16일*