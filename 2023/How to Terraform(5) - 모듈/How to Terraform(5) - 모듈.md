# How to Terraform(5) - 모듈

## 시리즈 목차

- [How to Terraform(1) - 테라폼과의 가벼운(?) 첫 만남](../How%20to%20Terraform%281%29%20-%20%ED%99%98%EA%B2%BD%20%EC%84%A4%EC%A0%95%ED%95%98%EA%B8%B0%2FHow%20to%20Terraform%281%29%20-%20%ED%85%8C%EB%9D%BC%ED%8F%BC%EA%B3%BC%EC%9D%98%20%EA%B0%80%EB%B2%BC%EC%9A%B4%28%3F%29%20%EC%B2%AB%20%EB%A7%8C%EB%82%A8.md)
- [How to Terraform(2) - 변수와 출력](../How%20to%20Terraform%282%29%20-%20%EB%B3%80%EC%88%98%EC%99%80%20%EC%B6%9C%EB%A0%A5%2FHow%20to%20Terraform%282%29%20-%20%EB%B3%80%EC%88%98%EC%99%80%20%EC%B6%9C%EB%A0%A5.md)
- [How to Terraform(3) - 상태 관리](../How%20to%20Terraform%283%29%20-%20%EC%83%81%ED%83%9C%20%EA%B4%80%EB%A6%AC%2FHow%20to%20Terraform%283%29%20-%20%EC%83%81%ED%83%9C%20%EA%B4%80%EB%A6%AC.md)
- [How to Terraform(4) - 상태관리 개선하기](../How%20to%20Terraform%284%29%20-%20%EC%83%81%ED%83%9C%EA%B4%80%EB%A6%AC%20%EA%B0%9C%EC%84%A0%ED%95%98%EA%B8%B0%2FHow%20to%20Terraform%284%29%20-%20%EC%83%81%ED%83%9C%EA%B4%80%EB%A6%AC%20%EA%B0%9C%EC%84%A0%ED%95%98%EA%B8%B0.md)

Terraform을 사용하다보면 동일한 유형의 리소스를 여러번 생성해야 할 때가 있습니다. 예를 들면 개발, 스테이징, 운영 환경을 각각 분리된 EC2 인스턴스로 생성한다고 했을 때, 거의 동일한 EC2 인스턴스를 리소스로 정의하게 되겠죠. 지금까지는 동일한 코드를 복사해서 붙여넣는 방식으로 각 리소스를 정의했습니다. 그런데 이렇게 하면 EC2 인스턴스의 설정이 복잡해지고, 개수가 많아질 수록 중복되는 코드의 양이 늘어나겠죠? 만약 기존에 정의된 EC2 인스턴스의 AMI를 변경해야 한다면, 모든 코드를 찾아서 일일이 변경해야만 할 겁니다. 그리고 이렇게 하다 보면 실수할 확률이 높아지겠죠?

이렇게 동일한 리소스를 반복해서 사용해야 하는 경우, 리소스의 공통적인 내용을 재사용 가능한 형태로 정의해놓고 필요할 때마다 참조해서 사용하되, 바뀌는 부분만 따로 명시해줄 수 있으면 좋을 겁니다. 이런 재사용 가능한 형태로 정의된 리소스를 모듈이라고 합니다.

## 모듈이란?

Terraform 모듈은 동일한 유형의 리소스를 여러 번 생성할 때, 코드를 중복해서 작성하지 않도록 재사용 가능한 형태로 정의할 수 있습니다. 프로그래밍 언어에서는 함수나 클래스를 떠올려보시면 비슷합니다. 자 그럼 우선 모듈을 사용하기 전에 중복해서 리소스를 정의하는 코드를 먼저 확인해볼까요?

```terraform
// AWS 프로바이더 설정
provider "aws" {
  // 아시아 태평양(서울) 리전
  region = "ap-northeast-2"
}

// 개발 환경 EC2 인스턴스
resource "aws_instance" "example_dev" {
  // ap-northeast-2 리전의 Ubuntu 20.04 AMI
  ami           = "ami-0c6e5afdd23291f73"
  instance_type = "t2.micro"

  vpc_security_group_ids = [aws_security_group.example_dev.id]

  // 최초 인스턴스 생성 시에만 실행
  user_data = <<-EOF
              #!/bin/bash
              echo "Hello, World" > index.html
              nohup busybox httpd -f -p 8080 &
              EOF

  tags = {
    Name = "terraform-example"
  }
}

resource "aws_security_group" "example_dev" {
  name = "terraform-example"

  // 인바운드 규칙
  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

// 스테이징 환경 EC2 인스턴스
resource "aws_instance" "example2_stg" {
  // ap-northeast-2 리전의 Ubuntu 20.04 AMI
  ami           = "ami-0c6e5afdd23291f73"
  instance_type = "t2.micro"

  vpc_security_group_ids = [aws_security_group.example_stg.id]

  // 최초 인스턴스 생성 시에만 실행
  user_data = <<-EOF
              #!/bin/bash
              echo "Hello, World" > index.html
              nohup busybox httpd -f -p 8080 &
              EOF

  tags = {
    Name = "terraform-example"
  }
}

resource "aws_security_group" "example_stg" {
  name = "terraform-example-stg"

  // 인바운드 규칙
  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
```

개발 환경과 스테이징 환경에서 동작할 EC2 인스턴스를 2개 정의해놓았습니다. 대부분의 코드가 중복되어 있죠? 간단한 코드지만 계속해서 늘어난다고 생각하면 좀 별로이긴합니다. 그러면 일단 중복을 제거할 수 있도록 모듈을 작성해보겠습니다. `modules/ec2` 폴더를 만들고, 그 안에 `ec2` 폴더를 만들어서 `main.tf` 파일을 생성합니다.

```terraform
// 로컬 변수 정의
locals {
  // EC2 인스턴스에 할당할 태그
  security_group_name = "terraform-example-${var.env}"
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
}
```

그리고 `variables.tf` 파일을 생성해서 변수를 정의합니다.

```terraform
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

마지막으로 `outputs.tf` 파일을 생성해서 EC2 인스턴스의 public IP 주소를 출력하도록 설정합니다.

```terraform
output "public_ip" {
  value = aws_instance.ec2.public_ip
}
```

이제 모듈을 사용하는 코드를 작성해보겠습니다. `dev/ec2` 폴더에 `main.tf` 파일을 생성하고, 다음과 같이 작성합니다.

```terraform
// AWS 프로바이더 설정
provider "aws" {
  // 아시아 태평양(서울) 리전
  region = "ap-northeast-2"
}

// 모듈 사용
module "ec2" {
  // 모듈의 소스 경로 지정
  source = "../../modules/ec2"

  // 모듈에 전달할 인자
  env = "dev"
  port = 8080
}
```

모듈의 소스 경로를 상대 경로로 지정해주고, 모듈에 전달하고 싶은 변수를 지정해줍니다. 그리고 `stg/ec2` 폴더에도 `main.tf` 파일을 생성하고, 다음과 같이 작성합니다.

```terraform
// AWS 프로바이더 설정
provider "aws" {
  // 아시아 태평양(서울) 리전
  region = "ap-northeast-2"
}

// 모듈 사용
module "ec2" {
  source = "../../modules/ec2"

  // 모듈에 전달할 인자
  env = "stg"
  port = 8080
}
```

그리고 모듈의 output을 이용해서 EC2 인스턴스의 public IP 주소를 출력하도록 `outputs.tf` 파일을 생성합니다.

```terraform
output "public_ip" {
  value = module.ec2.public_ip
}
```

참고로 폴더 구조는 다음과 같습니다.

```
.
├── modules
│   └── ec2
│       ├── main.tf
│       └── variables.tf
|       └── outputs.tf
├── dev
│   └── ec2
│       └── main.tf
│       └── outputs.tf
└── stg
    └── ec2
        └── main.tf
        └── outputs.tf
```

우선 `dev/ec2` 폴더에서 `tf init` 명령을 실행합니다.

```bash
> cd dev/ec2
> tf init

Initializing the backend...
Initializing modules...
- ec2 in ../../modules/ec2

Initializing provider plugins...
...
```

가장 먼저 지정된 모듈의 소스 경로에서 모듈을 가져와서 초기화 하는 걸 확인할 수 있죠? 그리고 `tf plan` 명령을 실행해볼까요?

> 참고로 모듈을 수정한 다음에 수정된 모듈을 반영하려면 `tf get -update`명령을 실행해야 합니다.

```bash
> tf plan
...
Plan: 2 to add, 0 to change, 0 to destroy.

Changes to Outputs:
  + public_ip = (known after apply)
...

> tf apply
...
Apply complete! Resources: 2 added, 0 changed, 0 destroyed.

Outputs:

public_ip = "..."
```

모듈에 설정된 내용대로 인스턴스가 생성되었고, 모듈의 출력으로 설정된 public_ip도 출력을 확인할 수 있습니다. 이제 `stg/ec2` 폴더에서도 동일한 명령을 실행해서 인스턴스를 추가로 생성할 수 있습니다.

## 정리

모듈을 활용하면 코드의 중복을 줄일 수 있으며, 수작업으로 인해 실수를 방지할 수 있습니다. 하지만 모듈은 또 다른 단점을 가지고 있습니다. 모듈을 수정하게 되면 모듈을 참조하는 모든 리소스의 상태 관리에 영향을 미친다는 점입니다. 어떻게 하면 모듈을 업데이트하면서도 특정 리소스는 이전 모듈을 참조하고, 특정 리로스는 새로 업데이트된 모듈을 참조하도록 할 수 있을까요? 다음 시간에는 그 방법을 알아보겠습니다!

# 참고자료

- [Terraform Up and Running, 3rd Edition](https://www.terraformupandrunning.com/)

# License

[저작자표시-비영리-변경금지 2.0 대한민국 (CC BY-NC-ND 2.0 KR)](https://creativecommons.org/licenses/by-nc-nd/2.0/kr/)
