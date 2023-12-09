# How to Terraform(6) - 모듈 버전관리

## 시리즈 목차

- [How to Terraform(1) - 테라폼과의 가벼운(?) 첫 만남](../How%20to%20Terraform%281%29%20-%20%ED%99%98%EA%B2%BD%20%EC%84%A4%EC%A0%95%ED%95%98%EA%B8%B0%2FHow%20to%20Terraform%281%29%20-%20%ED%85%8C%EB%9D%BC%ED%8F%BC%EA%B3%BC%EC%9D%98%20%EA%B0%80%EB%B2%BC%EC%9A%B4%28%3F%29%20%EC%B2%AB%20%EB%A7%8C%EB%82%A8.md)
- [How to Terraform(2) - 변수와 출력](../How%20to%20Terraform%282%29%20-%20%EB%B3%80%EC%88%98%EC%99%80%20%EC%B6%9C%EB%A0%A5%2FHow%20to%20Terraform%282%29%20-%20%EB%B3%80%EC%88%98%EC%99%80%20%EC%B6%9C%EB%A0%A5.md)
- [How to Terraform(3) - 상태 관리](../How%20to%20Terraform%283%29%20-%20%EC%83%81%ED%83%9C%20%EA%B4%80%EB%A6%AC%2FHow%20to%20Terraform%283%29%20-%20%EC%83%81%ED%83%9C%20%EA%B4%80%EB%A6%AC.md)
- [How to Terraform(4) - 상태관리 개선하기](../How%20to%20Terraform%284%29%20-%20%EC%83%81%ED%83%9C%EA%B4%80%EB%A6%AC%20%EA%B0%9C%EC%84%A0%ED%95%98%EA%B8%B0%2FHow%20to%20Terraform%284%29%20-%20%EC%83%81%ED%83%9C%EA%B4%80%EB%A6%AC%20%EA%B0%9C%EC%84%A0%ED%95%98%EA%B8%B0.md)
- [How to Terraform(5) - 모듈](../How%20to%20Terraform%285%29%20-%20%EB%AA%A8%EB%93%88%2FHow%20to%20Terraform%285%29%20-%20%EB%AA%A8%EB%93%88.md)

[이전 글](../How%20to%20Terraform%285%29%20-%20%EB%AA%A8%EB%93%88%2FHow%20to%20Terraform%285%29%20-%20%EB%AA%A8%EB%93%88.md)에서는 모듈을 활용해서 반복되는 리소스를 재사용 가능한 형태로 정의하는 방법을 알아봤습니다. 그런데 모듈로 재사용성을 높이는 건 좋지만, 문제가 있긴 했습니다. 모듈을 업데이트해야 하는 경우, 이미 모듈을 사용해서 생성된 리소스 역시 업데이트되는 모듈의 영향을 받게 된다는 점입니다. 모듈을 업데이트 하되 기존 리소스는 기존 모듈을 그대로 사용할 수 있도록 하위호환성을 도입하려면 어떻게 해야 할까요?

## Git 으로 모듈 버전 관리하기

여기서는 Git을 이용해서 모듈의 버전을 관리하고, 리소스도 각 모듈의 버전을 지정해서 참조하도록 해서 모듈과 리소스 간의 의존성을 관리하는 방법을 알아보겠습니다. 이렇게 하면 모듈을 업데이트할 때 기존 리소스에 영향을 주지 않으면서도 모듈을 업데이트할 수 있습니다. Git으로 모듈 버전을 관리하려면 2개의 Git 리포지토리가 필요합니다. 하나는 모듈을 관리하는 리포지토리, 또 하나는 모듈을 사용하여 리소스를 정의하는 리포지토리입니다. 우선 모듈을 관리하는 리포지토리를 만들어 보겠습니다.

## 모듈 리포지토리 만들기

모듈을 관리하는 리포지토리를 terraform-modules, 모듈을 사용해서 리소스를 관리하는 리포지토리를 terraform-live라고 가정하고 진행하겠습니다. 먼저 terraform-modules 리포지토리를 생성하고, [이전 글](../How%20to%20Terraform%285%29%20-%20%EB%AA%A8%EB%93%88%2FHow%20to%20Terraform%285%29%20-%20%EB%AA%A8%EB%93%88.md)에서 작성했던 EC2 인스턴스 모듈을 이 리포지토리에 옮깁니다. 폴더 구조는 다음과 같습니다.

```
terraform-modules
└── modules
    └── ec2
        ├── main.tf
        ├── outputs.tf
        └── variables.tf
```

그리고 이 리포지토리를 GitHub에 푸시합니다. 그리고 다음 명령을 통해 모듈의 버전을 태그합니다.

```bash
> git tag -a v0.0.1 -m "EC2 모듈 추가"

~/src/terraform-modules main
> git push --follow-tags
Enumerating objects: 1, done.
Counting objects: 100% (1/1), done.
Writing objects: 100% (1/1), 185 bytes | 185.00 KiB/s, done.
Total 1 (delta 0), reused 0 (delta 0), pack-reused 0
To https://github.com/netscout/terraform-modules.git
 * [new tag]         v0.0.1 -> v0.0.1
```

> 만약 이전 커밋에 지정된 태그를 삭제하고 싶다면 `git tag -d v0.0.1`의 형태로 명령을 사용하고 새로운 커밋에 다시 태그를 붙이면 됩니다. 이미 태그를 푸시한 상태라면 `git push --delete origin v0.0.1` 명령을 사용하면 됩니다.

## 모듈 리포지토리 참조하기

EC2 모듈의 첫 버전을 v0.0.1로 태그 했습니다. 이제 리소스 리포지토리를 만들어서 이 모듈을 참조해보겠습니다. 우선 terraform-live라는 리포지토리를 생성하고, [이전 글](../How%20to%20Terraform%285%29%20-%20%EB%AA%A8%EB%93%88%2FHow%20to%20Terraform%285%29%20-%20%EB%AA%A8%EB%93%88.md)에서 작성했던 EC2 인스턴스 리소스를 이 리포지토리로 옮겨줍니다. 폴더 구조는 다음과 같습니다.

terraform-live 폴더 안에 dev 폴더를 만들고, 그 안에 ec2 폴더를 만들고, service1 폴더를 만들어서 EC2 인스턴스 리소스를 옮겨줍니다.

```
terraform-live
└── dev
    └── ec2
        └── service1
            ├── main.tf
            └── outputs.tf
```

이전 글에서 작성한 `main.tf`는 로컬에 있는 모듈을 참조하도록 설정되어 있습니다. 로컬 대신에 Git의 모듈 리포지토리를 참조하도록 변경하겠습니다.

```terraform
// AWS 프로바이더 설정
provider "aws" {
  // 아시아 태평양(서울) 리전
  region = "ap-northeast-2"
}

// 모듈 사용
module "ec2" {
  //source = "../../modules/ec2"
  source = "github.com/netscout/terraform-modules//modules/ec2?ref=v0.0.1"

  // 모듈에 전달할 인자
  env = "dev"
  port = 8080
}
```

source 속성을 자세히 보시면 중간에 `//`가 들어가 있습니다. `//`의 앞쪽에는 모듈이 저장된 git 리포지토리의 주소를, 그리고 뒤 쪽에는 모듈의 경로와 태그로 지정된 버전을 명시합니다. 모듈을 참조하기 위한 Terraform 코드 작성은 완료됐지만 아직 한 가지 더 검토해야 할 게 남아있습니다.

## Terraform 에게 Git 리포지토리 접근 허용해주기

아마도 모듈 리포지토리는 private 리포지토리로 생성되었을 겁니다. 회사에서 작업하는 모듈에 대한 정의를 public 하게 공개할 수는 없기 때문이죠. 그렇다면 Terraform이 private 리포지토리에 접근할 수 있는 방법을 알려줘야 합니다.

### SSH 키 생성 및 등록하기

Terraform이 모듈 리포지토리를 참조하기 위해서 SSH 키를 생성해서 github에 등록해야 합니다. 여기서는 ed25519 알고리즘을 사용해서 키를 생성하겠습니다. ed25519는 rsa보다 보안이 더 강화된 알고리즘으로, 최근에는 ed25519를 사용하는 것을 권장하고 있습니다. 다음 명령을 실행하여 키를 생성합니다.

```bash
# "your_email@example.com" 대신에 github 이메일 주소를 입력하세요.
> ssh-keygen -t ed25519 -C "your_email@example.com"
# ssh-agent를 실행하여, 키를 등록한다.
> eval "$(ssh-agent -s)"
> ssh-add ~/.ssh/id_ed25519
```

그리고 생성된 키의 내용을 클립보드에 복사한 후, 그 내용을 github에 등록합니다. 키의 내용을 클립보드에 복사하려면 ubuntu에서는 xclip을 사용하고,

```bash
> cat ~/.ssh/id_ed25519.pub | xclip -selection clipboard
```

mac에서는 pbcopy를 사용하면 됩니다.

```bash
> cat ~/.ssh/id_ed25519.pub | pbcopy
```

이제 클립보드에 복사된 키를 github에 등록할 차례입니다. github에 접속하고 우측 상단의 아바타 버튼을 누르고 `Settings` > `SSH and GPG keys` > `New SSH key`를 클릭합니다.

`Title`에는 키를 등록하는 컴퓨터의 이름을 적고, `Key`에는 위에서 복사한 키를 붙여넣으면 완료입니다. 그러면 모듈을 잘 참조하는지 확인해볼까요?

```bash
> cd dev/ec2/service1
> tf init
Initializing the backend...
Initializing modules...
Downloading git::https://github.com/netscout/terraform-modules.git?ref=v0.0.1 for ec2...
- ec2 in .terraform/modules/ec2/modules/ec2

...
Terraform has been successfully initialized!
...
```

정상적으로 잘 참조하고 있습니다! 이제 모듈을 업데이트 해볼까요?

### 모듈 업데이트하기

모듈 리포지토리의 EC2 모듈을 업데이트 해보겠습니다. `modules/ec2/main.tf` 파일을 열고, locals의 변수 정의를 수정하고 aws_security_group.sg 리소스에 ssh 접속을 위한 인바운트 규칙을 추가합니다.

```terraform
// 로컬 변수 정의
locals {
  // EC2 인스턴스에 할당할 태그
  security_group_name = "${var.name}-${var.env}"
}

resource "aws_instance" "ec2" {
  // ap-northeast-2 리전의 Ubuntu 20.04 AMI
  ami           = var.ami_id # 변수 ami_id 참조
  instance_type = var.instance_type

  vpc_security_group_ids = [aws_security_group.sg.id]

  // 최초 인스턴스 생성 시에만 실행
  user_data = <<-EOF
              #!/bin/bash
              echo "Hello, World" > index.html
              nohup busybox httpd -f -p ${var.port} &
              EOF

  tags = {
    Name = "terraform-example"
  }
}

resource "aws_security_group" "sg" {
  name = local.security_group_name

  // 인바운드 규칙
  ingress {
    from_port   = var.port
    to_port     = var.port
    protocol    = "tcp"
    cidr_blocks = var.cidr_blocks
  }

  // ssh 접속을 위한 인바운드 규칙(새로 추가)
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = var.cidr_blocks
  }
}
```

다음으로 variables.tf 파일을 열고, name 변수를 추가합니다.

```terraform
// name 변수 추가
variable "name" {
  type = string
}

variable "env" {
  type = string
  // 변수 값이 "dev", "stg", "prod" 중에 하나로 설정되야함
  validation {
    condition = contains(["dev", "stg", "prod"], var.env)
    error_message = "env must be one of dev, stg, prod"
  }
}

variable "ami_id" {
  type = string
  default = "ami-0c6e5afdd23291f73"
}

variable "instance_type" {
  type = string
  default = "t2.micro"
}

variable "port" {
  type = number
}

variable "cidr_blocks" {
  type = list(string)
  default = ["0.0.0.0/0"]
}
```

그리고 변경 내용을 커밋하고, 모듈 리포지토리를 푸시, 그리고 태그를 추가합니다.

```bash
> git tag -a v0.0.2 -m "EC2에 SSH 인바운드 규칙 추가"
> git push --follow-tags
Enumerating objects: 1, done.
Counting objects: 100% (1/1), done.
Writing objects: 100% (1/1), 207 bytes | 207.00 KiB/s, done.
Total 1 (delta 0), reused 0 (delta 0), pack-reused 0
To https://github.com/netscout/terraform-modules.git
 * [new tag]         v0.0.2 -> v0.0.2
```

### 업데이트된 모듈 참조하기

이제 업데이트된 모듈을 참조해보겠습니다. terraform-live/dev/ec2/service2 폴더를 추가하고, terraform-live/dev/ec2/service1 폴더에 작성된 main.tf, outputs.tf 파일을 복사해서 붙여넣습니다. 그리고 service2 폴더에 있는 main.tf 파일을 열고, 다음과 같이 모듈의 버전을 v0.0.2로 변경하고 name 변수에 인자를 할당합니다.

```terraform
// AWS 프로바이더 설정
provider "aws" {
  // 아시아 태평양(서울) 리전
  region = "ap-northeast-2"
}

// 모듈 사용
module "ec2" {
  //source = "../../modules/ec2"
  source = "github.com/netscout/terraform-modules//modules/ec2?ref=v0.0.2"

  // 모듈에 전달할 인자
  name="service2"
  env = "dev"
  port = 8080
}
```

init 명령을 실행해서 모듈 참조를 확인합니다.

```bash
> cd dev/ec2/service2
> tf init
Initializing the backend...
Initializing modules...
Downloading git::https://github.com/netscout/terraform-modules.git?ref=v0.0.2 for ec2...
- ec2 in .terraform/modules/ec2/modules/ec2
...
```

v0.0.2 버전의 모듈의 정상적으로 참조하고 있죠? v0.0.1을 참조하는 service1을 다시 확인해볼까요?

```bash
# 모듈 다시 업데이트 하기
> tf get -update
> tf plan
...
```

service1의 리소스 정의에는 모듈 v0.0.2에서 새로 추가된 name 변수가 지정되어 있지 않지만, name 변수가 아직 추가되지 않는 v0.0.1을 참조하고 있기 때문에 모듈의 버전 변경과 관계없이 잘 동작하는 걸 확인할 수 있습니다.

## 정리

Git으로 모듈의 버전을 관리하면 모듈이 변경되더라도 이전에 생성된 리소스에는 영향을 주지 않고 안전하게 모듈을 업데이트할 수 있습니다. 점점 좋아지고 있죠? 그런데 여전히 중복은 남아있습니다. service1과 service2 에는 각각 provider에 대한 정의가 있습니다. 그리고 여기서는 backend를 사용하지 않았지만, S3같은 backend를 사용하면 그 설정 역시 각 리소스에 중복해서 들어가게 될 겁니다. 다음에는 이런 중복을 제거할 수 있는 방법을 알아보겠습니다.

# 참고자료

- [Terraform Up and Running, 3rd Edition](https://www.terraformupandrunning.com/)

# License

[저작자표시-비영리-변경금지 2.0 대한민국 (CC BY-NC-ND 2.0 KR)](https://creativecommons.org/licenses/by-nc-nd/2.0/kr/)
