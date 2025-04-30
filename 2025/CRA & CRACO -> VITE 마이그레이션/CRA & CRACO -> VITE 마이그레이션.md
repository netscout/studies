# CRA & CRACO -> VITE 마이그레이션

## CRA와 씨름하다

2022년 한 프로젝트를 처음 시작할 때 저는 리액트에 대해 아주 잘 알지는 못했습니다. 그렇기 때문에 공식 사이트에 나와있는 대로 CRA(Create React App)을 이용해 프로젝트를 생성했습니다. 그리고 몇 달 후, 새롭게 합류한 프론트 개발자가 그 프로젝트에 [CRACO](https://craco.js.org)를 추가했습니다. 그 당시에는 CRA의 여러 문제점으로 인해서 eject하여 커스텀하게 설정을 하던지, CRACO나 [React App Rewired](https://github.com/timarney/react-app-rewired)같은 방법을 사용했던 걸로 기억합니다.

그리고 프로젝트가 점점 커지자 몇 가지 문제가 발생했습니다.

- 개발 환경의 메모리 문제
  CRA가 Webpack + Babel을 사용하는데, 앱이 시작 될 때 src 폴더와 node_modules 폴더에서 임포트 되는 모든 파일을 찾아서 의존성 그래프를 생성합니다. 그리고 생성된 결과를 번들링해서 메모리에 저장하는데요, 이 과정에서 많은 메모리가 필요합니다. 프로젝트가 복잡해지고 커질수록 더 악화되겠죠! 그래서 `npm run dev` 명령에 `NODE_OPTIONS='--max-old-space-size=4096'` 옵션을 추가하여 메모리를 조금 더 사용할 수 있도록 했습니다.
- 빌드 시간 문제
  프로젝트가 커짐에 따라 시간이 점점 더 걸리더니 결국 빌드에 10분씩 걸리는 지경에 이르렀습니다. 사소한 실수 하나를 수정하기 위해 10분이나 기다려야 하는 건 너무 비효율적이었습니다. 그래서 빌드 시간을 단축시키기 위해서 esbuild를 craco에 엮어서 사용하는 방식을 통해 어느 정도 줄일 수 있었지만 여전히 4-5분 정도의 빌드 타임이 소요되었습니다.
- CRA의 지원 종료
  2023년 즈음해서 사실상 [React 팀에서도 CRA에 대한 지원을 조용히 그만두기로 결정](https://github.com/facebook/create-react-app/issues/13072)했습니다. 그래서 회사에서 추가로 진행하는 프로젝트에 더 이상 CRA를 사용할 수 없는 상황이었고, 이 때문에 대안을 찾아야 했습니다.

이때쯤 부터 `VITE`에 대한 이야기가 더 많아졌던 것 같습니다. 몇 년전에 `Vue.js 3`를 처음 스터디 해볼 때 `VITE`라는 게 있더라 정도는 봤던 기억이 있었는데, 구체적으로 이게 뭔지는 몰랐죠. 그래서 작년 말에 새로운 프로젝트를 시작 할 때 `VITE`를 사용해보기로 했습니다.

## VITE에 놀라다

VITE는 제 생각보다 더 훌륭했습니다. 왜 사람들이 VITE를 칭송하는지 알겠더군요!

- 개발 환경의 쾌적함
  VITE는 앱이 시작될 때 앱 전체를 번들링 하지 않습니다. 요청되는 페이지와 관련있는 임포트만 번들링하여 제공합니다. 그러니까 엄청 빠르고, 메모리 문제도 없습니다!
- 빌드도 빠르다!
  초기 프로젝트 생성과정에서 간단하게 SWC를 선택할 수 있고, SWC를 이용하면 빌드 시간이 매우 단축됩니다. CRA에는 없던 다른 문제가 발생하긴 했지만 그건 나중에 자세히 말씀드리죠.

왜 그동안 CRA를 가지고 씨름했었을까 하는 현타가 몰려왔지만, 어쩔 수 없었죠. VITE가 좋은지 몰랐으니까!

그래서 한동안 새 프로젝트를 작업하다보니 그 작업이 다시 돌아왔습니다. 네, 기존 프로젝트 유지보수 작업이죠. 즐거운 작업입니다. 근데 이제와서 다시 CRA를 사용하는 건 마음이 불편했습니다. 실제로도 불편했고요. 메모리 옵션을 줬음에도 불구하고 꺼지는 앱과 HMR 할 때 마다 걸리는 시간! 그래서 대충 각을 잡아보고 마이그레이션 작업을 시작했습니다.

## 인생은 실전이다! CRA -> VITE 마이그레이션

1. CRA와 craco 제거

우선 CRA와 craco를 제거합니다.

```bash
npm uninstall react-scripts @craco/craco
```

2. VITE와 필수 플러그인을 설치

다음 명령으로 VITE 및 필수 플러그인을 설치합니다.

```bash
npm install -D vite @vitejs/plugin-react-swc vite-plugin-svgr vite-tsconfig-paths @swc/core @swc/cli
```

vite외에 설치되는 패키지 목록은 다음과 같습니다.

- @vitejs/plugin-react-swc: VITE가 React를 지원할 수 있도록 하며, SWC를 사용하여 빌드 속도를 높입니다.
- vite-plugin-svgr: SVG 파일을 React 컴포넌트로 변환합니다.(추후에 왜 설치했는지 설명합니다.)
- vite-tsconfig-paths: 타입스크립트 경로 별칭을 사용할 수 있도록 합니다.
- @swc/core: SWC의 핵심 모듈입니다.
- @swc/cli: SWC를 명령줄에서 사용할 수 있도록 합니다.

3. VITE 설정 파일을 생성

프로젝트의 루트 폴더에 `vite.config.ts` 파일을 생성하고 다음과 같이 작성합니다.

```typescript
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import tsconfigPaths from "vite-tsconfig-paths";
iport svgr from "vite-plugin-svgr";

// https://vite.dev/config/
export default defineConfig({
  define: {
    global: "globalThis", // <- 패키지 중에 global 관련 에러가 발생할 경우 추가
  },
  plugins: [
    react(),
    tsconfigPaths(),
    svgr(),
  ],
  server: {
    port: 3000,
  },
});
```

4. tsconfig.json 파일의 경로 정보 수정

프로젝트의 tsconfig.json 파일을 수정합니다.
아래 내용은 예시이므로 프로젝트에 맞게 수정해야 합니다.

```json
{
  "compilerOptions": {
    "baseUrl": "./src",
    "paths": {
      "@/*": ["*"],
      "@store": ["./app/store"],
      "@pages/*": ["./pages/*"],
      "@components/*": ["./components/*"],
      ...
    },
    "target": "es2015",
    "lib": ["dom", "dom.iterable", "esnext"],
    "allowJs": true,
    "skipLibCheck": true,
    "esModuleInterop": true,
    "allowSyntheticDefaultImports": true,
    "strict": true,
    "forceConsistentCasingInFileNames": true,
    "noFallthroughCasesInSwitch": true,
    "module": "esnext",
    "moduleResolution": "node",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx",
    "noImplicitAny": false
  },
  "include": ["src"]
}
```

5. vite-env.d.ts 파일 생성

`src/vite-env.d.ts` 파일을 생성하고 다음과 같이 작성합니다.

```typescript
/// <reference types="vite/client" />

// svg 파일을 React 컴포넌트로 사용할 수 있도록 선언
declare module "*.svg?react" {
  import * as React from "react";
  const ReactComponent: React.FunctionComponent<
    React.SVGProps<SVGSVGElement> & { title?: string }
  >;
  export default ReactComponent;
}
```

6. index.html 파일 수정

`public/index.html` 파일을 프로젝트의 루트 폴더로 옮기고, 스크립트 링크를 다음과 같이 수정합니다.

```html
...
<body>
  <div id="root"></div>
  <script type="module" src="/src/main.tsx"></script>
</body>
...
```

7. 환경 변수 업데이트

프로젝트의 .env 파일에서 환경 변수를 `REACT_APP_` 대신 `VITE_` 접두사를 사용하도록 수정합니다.

```bash
// 이전
REACT_APP_API_URL=http://localhost:3000

// 다음과 같이 수정
VITE_API_URL=http://localhost:3000
```

8. package.json 파일의 스크립트 수정

프로젝트의 package.json 파일을 수정합니다.

```json
...
  "scripts": {
    "dev": "vite",
    "build": "NODE_ENV=production tsc -b && vite build",
  }
...
```

9. svg 임포트 수정

앞서 vite-plugin-svgr 플러그인을 설치했었는데요, VITE에서는 CRA 처럼 svg파일을 컴포넌트로 임포트 할 수 없습니다. 그래서 플러그인을 설치하고 임포트 부분을 다음과 같이 수정합니다.

```typescript
// 이전 방식
import { ReactComponent as Checklist } from "../assets/checklist.svg";

// 다음과 같이 수정
import Checklist from "../assets/checklist.svg?react";
```

앞서 작성한 `vite-env.d.ts` 파일을 다시 확인해보시면 `.svg?react`를 처리하기 위한 선언이 추가되어 있습니다.

10. ERR_REQUIRE_ESM 에러

`npm run dev` 명령으로 프로젝트를 실행했을 때, 다음과 같은 에러가 발생할 수 있는데요,

```bash
...
Error [ERR_REQUIRE_ESM]: require() of ES Module /path/to/vite-tsconfig-paths/dist/index.js from /path/to/vite.config.ts not supported.
...
```

그렇다면 프로젝트가 ESM 모듈을 사용하도록 타입을 변경해야 합니다.

```json
{
  "type": "module"
}
```

11. global is not defined 에러

몇몇 패키지는 global 변수를 참조하는데, VITE가 global 대신에 브라우저에서 사용하는 globalThis를 사용하도록 설정해주면 됩니다.

앞서 작성한 `vite.config.ts` 파일을 확인해보시면 다음과 같은 설정이 있었습니다.

```typescript
export default defineConfig({
  define: {
    global: 'globalThis', // ← 바로 이겁니다!
  },
  ...
});
```

12. process.env.NODE_ENV 를 찾을 수 없는 오류

VITE는 `process.env.NODE_ENV` 대신 `import.meta.env.MODE`를 사용합니다.

```typescript
// 이전 방식
const isProduction = process.env.NODE_ENV === "production";

// 다음과 같이 수정
const isProduction = import.meta.env.MODE === "production";
```

13. require를 import로 변경

VITE는 CommonJS 모듈 시스템을 사용하지 않습니다. 그래서 require를 import로 변경해야 합니다.

```typescript
// 이전 방식
import i18n from "i18next";
import { initReactI18next } from "react-i18next";

i18n.use(initReactI18next).init({
  fallbackLng: "en-US",
  lng: "en-US",
  resources: {
    en: {
      translations: require("./locales/en/translations.json"),
    },
    ko: {
      translations: require("./locales/ko/translations.json"),
    },
  },
  ...
});

// 다음과 같이 수정
import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import enTranslations from "./locales/en/translations.json";
import koTranslations from "./locales/ko/translations.json";

i18n.use(initReactI18next).init({
  fallbackLng: "en-US",
  lng: "en-US",
  resources: {
    en: {
      translations: enTranslations,
    },
    ko: {
      translations: koTranslations,
    },
  },
  ...
});
```

## 빌드에서 문제가 발생하다.

CRA는 개발환경에서는 문제가 많았지만, 빌드 속도를 제외하고 빌드에서 문제는 특별히 없었습니다. 그런데 VITE로 변경하니 로컬에서는 빌드가 잘 되는데 Github Actions에서는 메모리 이유로 빌드가 실패하는 문제가 발생했습니다.

원인을 찾다보니, 너무 거대한 번들 사이즈가 원인이었습니다. 현재 작업 중인 프로젝트는 좀 복잡한 요구사항으로 인해 무거운 패키지를 많이 사용할 수 밖에 없었는데요. 가장 큰 번들의 크기가 무려 16MB에 이를 정도 였습니다.

다음과 같이 manualChunks 옵션을 사용하여 번들의 크기를 분할하는 방법이 있었지만,

```typescript
build: {
  rollupOptions: {
    output: {
      manualChunks(id) {
        if (id.includes('node_modules')) {
          if (id.includes('react')) return 'vendor-react';
          if (id.includes('codemirror')) return 'vendor-codemirror';
          return 'vendor';
        }
      },
    },
  },
},
```

빌드되고 나서 실행과정에서 문제가 생겼기 때문에, 다른 방법을 찾아야 했습니다.

다른 방법으로는 lazy import를 사용하는 방법이었는데, 프로젝트가 워낙 복잡한 상황이었기에 한 번에 적용할 수는 없었고 일단은 Github Actions의 빌드 설정에 메모리 옵션을 추가하는 것으로 해결했습니다.

```yaml
- run: CI=false npm run build
  env:
    NODE_OPTIONS: --max-old-space-size=8192
```

일단은 한숨 돌리고 프로젝트를 유지보수 하면서 점진적으로 lazy import를 적용해 나가면서 상황을 개선 해봐야 겠습니다.
