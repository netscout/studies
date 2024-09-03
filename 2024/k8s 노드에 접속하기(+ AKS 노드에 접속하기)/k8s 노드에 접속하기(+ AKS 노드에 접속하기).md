# k8s 노드에 접속하기(+ AKS 노드에 접속하기)

k8s에 대해서 스터디하거나 서비스를 배포하다 보면 가끔 워커 노드에 접속해야 할 때가 있습니다. 예를 들면 pod에서 DB나 Redis 등에 제대로 접속하지 못할 때 혹시 네트워크에 문제 있는지 확인하거나 워커 노드의 로그 등을 수집할 때 해당 로그의 위치가 어디인지 등을 확인해야 할 때가 있습니다.

이럴 때 노드에 접속할 수 있는 몇 가지 방법을 알아보겠습니다.

## k9s 사용하기

### k9s 설치하기

[k9s](https://k9scli.io/)는 k8s 클러스터를 쉽게 관리할 수 있는 터미널 기반의 UI 도구입니다. k9s를 사용하면 k8s 클러스터의 상태를 실시간으로 확인하고, 리소스를 생성하거나 삭제할 수 있습니다. 또한 k8s 클러스터의 로그를 확인하거나, 특정 pod에 쉘로 접속하는 등의 작업을 쉽게 수행할 수 있습니다. 그래픽 환경이 아니라 처음엔 어색할 수 있지만, 익숙해지면 훨씬 빠르게 작업할 수 있습니다.

k9s는 다음과 같이 설치할 수 있습니다.

```bash
# MacOS
> brew install derailed/k9s/k9s

# Ubuntu(linuxbrew)
> brew install derailed/k9s/k9s

# Windows
> scoop install k9s
```

그리고 **k9s**를 실행하면 됩니다.

### k9s로 노드에 접속하기

k9s에서 **:pod** 명령으로 pod를 선택하고 **s**를 누르면 해당 pod에 쉘로 접속할 수 있습니다. 그런데 **:node** 명령으로 노드 목록을 표시해보면 노드에는 쉘로 접속할 수 있는 **s** 옵션이 표시되지 않습니다. 클러스터의 설정 파일을 찾아서 해당 클러스터에 대해서 쉘을 사용할 수 있도록 설정해줘야 합니다.

클러스터 설정 파일을 찾으려면 다음 명령을 실행합니다.

```bash
> k9s info
 ____  __.________
|    |/ _/   __   \______
|      < \____    /  ___/
|    |  \   /    /\___ \
|____|__ \ /____//____  >
        \/            \/

Version:           v0.32.5
Config:            /Users/onlifecoding/Library/Application Support/k9s/config.yaml
...
```

**Config** 항목에 표시되는 경로에 이동하면 **clusters** 폴더가 있고, 이 폴더에는 k9s로 접속한 적이 있는 k8s 클러스터의 설정 파일이 각 폴더 별로 저장되어 있습니다.

쉘로 접속하려는 클러스터의 폴더로 이동해서 **config.yaml** 파일의 **nodeShell** 항목을 **true**로 변경하면 됩니다.

```yaml
k9s:
  cluster: aks-test-stg
  namespace:
    active: default
    lockFavorites: false
    favorites:
      - default
  view:
    active: node
  featureGates:
    nodeShell: true # 노드에 쉘 접근 허용하기
  portForwardAddress: localhost
```

그리고 다시 k9s에서 **:node** 명령을 실행하면 노드에 쉘로 접속할 수 있는 **s** 옵션이 표시됩니다. 해당 노드를 선택하고 **s**를 누르면 노드에 pod이 배포되고, 해당 pod로 노드의 네트워크 통신 등을 테스트 할 수 있습니다.

참고로, 생성되는 pod는 앞서 **k9s info**로 확인했던 **Config** 파일에 설정되어 있으며, 기본 값은 이렇습니다.

```yaml
---
shellPod:
  image: busybox:1.35.0
  namespace: default
  limits:
    cpu: 100m
    memory: 100Mi
```

이 방법은 매우 간단하긴 하지만 노드 자체에는 접근할 수 없다는 단점이 있습니다.

## AKS 노드에 접속하기

AKS에서는 **kubectl debug**를 통해 노드에 접속할 수 있습니다.

우선 k9s나 kubectl을 통해 노드의 이름을 확인합니다.

```bash
> kubectl get nodes -o wide
NAME                              STATUS   ROLES    AGE   VERSION   INTERNAL-IP   EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION      CONTAINER-RUNTIME
aks-default-28123321-vmss000000   Ready    <none>   18m   v1.29.7   10.224.0.4    <none>        Ubuntu 22.04.4 LTS   5.15.0-1071-azure   containerd://1.7.20-1
```

그리고 접속하려는 노드의 이름을 사용하여 **kubectl debug** 명령을 실행합니다.

```bash
> kubectl debug node/aks-default-28123321-vmss000000 -it --image=mcr.microsoft.com/cbl-mariner/busybox:2.0
Creating debugging pod node-debugger-aks-default-28123321-vmss000000-2jj6t with container debugger on node aks-default-28123321-vmss000000.
If you don't see a command prompt, try pressing enter.
/
```

노드에 debug용 pod이 생성되고, 해당 pod에 접속되었습니다. 이제 다음 명령으로 node에 접근할 수 있습니다.

```bash
> chroot /host
```

노드에서 확인작업을 마치고 **exit** 빠져나오면, pod은 종료되지만 삭제되지 않고 남아있는데요, k9s에서 **ctrl + d**를 눌러 삭제하거나 **kubectl delete pod** 명령으로 삭제할 수 있습니다.

```bash
> kubectl delete pod node-debugger-aks-default-28123321-vmss000000-2jj6t
```

## 혹시 클라우드 프로바이더에 상관없이 노드에 접속하는 방법은?

마지막으로 [kubectl-plugins](https://github.com/luksa/kubectl-plugins)를 이용하는 방법이 있습니다. 이 방법은 클라우드 프로바이더에 상관없이 노드에 접속할 수 있는 방법인데요, 클러스터의 어드민 권한이 필요합니다.

## 참고자료

- [k9s NodeShell](https://k9scli.io/topics/shell/)
- [유지 관리 또는 문제 해결을 위해 AKS(Azure Kubernetes Service) 클러스터 노드에 연결](https://learn.microsoft.com/ko-kr/azure/aks/node-access)
- [kubectl-plugins](https://github.com/luksa/kubectl-plugins)
