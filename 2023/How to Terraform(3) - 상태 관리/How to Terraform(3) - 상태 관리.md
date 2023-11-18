# How to Terraform(3) - 상태 관리

## 상태라니 무슨 상태?

상태가 뭔지 설명하기 전에 간단한 실험을 해보죠! 새로운 폴더를 생성하고 아래 코드를 `main.tf`에 추가하겠습니다.

```terraform
// AWS 프로바이더 설정
provider "aws" {
  // 아시아 태평양(서울) 리전
  region = "ap-northeast-2"
}
```

그리고 `tf init`을 실행해봅시다.

```bash
> tf init

Initializing the backend...

Initializing provider plugins...
...

Terraform has been successfully initialized!
...
```

맨 처음에 `backend`라는 표현이 등장합니다. 이게 웹 어플리케이션도 아닌데 backend는 뭘 말하는 걸까요~? 일단은 궁금증을 뒤로 하고 `tf plan`을 실행해봅시다.

```bash
> tf plan
No changes. Your infrastructure matches the configuration.

Terraform has compared your real infrastructure against your configuration and found no differences, so no changes
are needed.
```

아무 것도 바꿀 게 없다고 합니다. 리소스를 하나도 선언하지 않았으니 당연하겠죠. 그런데 여기서 `tf apply`를 실행해볼까요?

```bash
> tf apply
No changes. Your infrastructure matches the configuration.

Terraform has compared your real infrastructure against your configuration and found no differences, so no changes
are needed.

Apply complete! Resources: 0 added, 0 changed, 0 destroyed.
```

바꿀건 하나도 없지만 apply 명령이 완료되었습니다. 그런데 apply 명령을 실행한 폴더를 확인해보시면, `terraform.tfstate`라는 파일이 생겼습니다! 파일을 확인 해보면 이런 내용이 있습니다.

```json
{
  "version": 4,
  "terraform_version": "1.6.2",
  "serial": 1,
  "lineage": "2e6b0c74-0e92-bfe8-1197-36e5547d9375",
  "outputs": {},
  "resources": [],
  "check_results": null
}
```

이 파일에는 현재 Terraform 코드를 apply해서 구성된 리소스의 상태가 기록되고 있습니다. 즉, 가장 최근에 실행한 apply 명령을 통해 구성된 리소스의 상태(여기서는 aws 클라우드에 생성된 리소스의 상태)가 기록됩니다. 여러분이 plan 명령을 실행 할 때 뭐가 추가되고, 뭐가 수정되며 뭐가 삭제되는지를 이 파일에 기록된 현재 상태를 기준으로 판단하는 거죠. 이 파일에 포함된 속성은 다음과 같습니다.

- `version`: Terraform의 상태 파일 형식 버전입니다. 현재는 4입니다.
- `terraform_version`: Terraform의 버전입니다.
- `serial`: 상태 파일이 변경될 때마다 증가하는 숫자입니다. 이 숫자는 상태 파일이 변경될 때마다 증가하며, 이 숫자를 통해 상태 파일이 변경되었는지를 판단합니다.
- `lineage`: 상태 파일의 고유 식별자입니다. 처음 생성될 때 부여되고, 그 이후로는 변경되지 않습니다.
- `outputs`: Terraform 코드에서 `output`으로 정의한 출력 값들이 기록됩니다.
- `resources`: Terraform 코드에서 `resource`로 정의한 리소스들의 상태가 기록됩니다.
- `check_results`: Terraform 코드에서 `data`로 정의한 데이터 소스들의 상태가 기록됩니다.

그렇다면 EC2 인스턴스를 하나 추가해서 상태 파일이 어떻게 바뀌는지 확인해봅시다! `main.tf`에 아래 코드를 추가합니다.

```terraform
resource "aws_instance" "example" {
  // ap-northeast-2 리전의 Ubuntu 20.04 AMI
  ami           = "ami-0c6e5afdd23291f73"
  instance_type = "t2.micro"
}
```

그리고 `tf plan`을 실행합니다.

```bash
> tf plan
Terraform used the selected providers to generate the following execution plan. Resource actions are indicated
with the following symbols:
  + create

Terraform will perform the following actions:

  # aws_instance.example will be created
  + resource "aws_instance" "example" {
      + ami                                  = "ami-0c6e5afdd23291f73"
      + arn                                  = (known after apply)
      + associate_public_ip_address          = (known after apply)
      + availability_zone                    = (known after apply)
      + cpu_core_count                       = (known after apply)
      + cpu_threads_per_core                 = (known after apply)
      + disable_api_stop                     = (known after apply)
      + disable_api_termination              = (known after apply)
      + ebs_optimized                        = (known after apply)
      + get_password_data                    = false
      + host_id                              = (known after apply)
      + host_resource_group_arn              = (known after apply)
      + iam_instance_profile                 = (known after apply)
      + id                                   = (known after apply)
      + instance_initiated_shutdown_behavior = (known after apply)
      + instance_lifecycle                   = (known after apply)
      + instance_state                       = (known after apply)
      + instance_type                        = "t2.micro"
      + ipv6_address_count                   = (known after apply)
      + ipv6_addresses                       = (known after apply)
      + key_name                             = (known after apply)
      + monitoring                           = (known after apply)
      + outpost_arn                          = (known after apply)
      + password_data                        = (known after apply)
      + placement_group                      = (known after apply)
      + placement_partition_number           = (known after apply)
      + primary_network_interface_id         = (known after apply)
      + private_dns                          = (known after apply)
      + private_ip                           = (known after apply)
      + public_dns                           = (known after apply)
      + public_ip                            = (known after apply)
      + secondary_private_ips                = (known after apply)
      + security_groups                      = (known after apply)
      + source_dest_check                    = true
      + spot_instance_request_id             = (known after apply)
      + subnet_id                            = (known after apply)
      + tags_all                             = (known after apply)
      + tenancy                              = (known after apply)
      + user_data                            = (known after apply)
      + user_data_base64                     = (known after apply)
      + user_data_replace_on_change          = false
      + vpc_security_group_ids               = (known after apply)
    }

Plan: 1 to add, 0 to change, 0 to destroy.

──────────────────────────────────────────────────────────────────────────────────────────────────────────────────

Note: You didn't use the -out option to save this plan, so Terraform can't guarantee to take exactly these actions
if you run "terraform apply" now.
```

현재는 생성된 리소스가 하나도 없기 때문에, 상태 값은 비어있는 상태죠. 그래서 Terraform 코드에 정의된 리소스는 모두 새로 추가됩니다. 이제 `tf apply`를 실행합니다.

```bash
> tf apply
...
Apply complete! Resources: 1 added, 0 changed, 0 destroyed.
```

이제 다시 `terraform.tfstate` 파일의 내용을 확인합니다.

```json
{
  "version": 4,
  "terraform_version": "1.6.2",
  "serial": 3,
  "lineage": "2e6b0c74-0e92-bfe8-1197-36e5547d9375",
  "outputs": {},
  "resources": [
    {
      "mode": "managed",
      "type": "aws_instance",
      "name": "example",
      "provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
      "instances": [
        {
          "schema_version": 1,
          "attributes": {
            "ami": "ami-0c6e5afdd23291f73",
            ...
          },
          "sensitive_attributes": [],
          "private": "..."
        }
      ]
    }
  ],
  "check_results": null
}
```

방금 전 apply 명령으로 추가된 리소스를 확인할 수 있습니다.

상태는 Terraform이 리소스를 관리할 때 어떻게 변경해야 할지 판단하는 기준이 됩니다. 그렇기 때문에 상태는 Terraform에서 가장 중요합니다. 상태가 제대로 관리되지 않으면 어떤 리소스를 어떻게 변경해야 할지 Terraform이 제대로 판단할 수 없겠죠.

> 그렇기 때문에 Terraform으로 생성된 리소스를 관리 콘솔에서 직접 변경하면 안됩니다! 상태 파일에 기록된 내용과 실제 리소스의 상태가 달라지기 때문에 Terraform이 제대로 동작하지 않을 수 있습니다.

확인이 끝났으므로 destroy 명령으로 생성된 리소스를 모두 제거합니다. 참고로 제거 작업이 끝난 후 상태 파일은 이렇습니다.

```json
{
  "version": 4,
  "terraform_version": "1.6.2",
  "serial": 5,
  "lineage": "2e6b0c74-0e92-bfe8-1197-36e5547d9375",
  "outputs": {},
  "resources": [],
  "check_results": null
}
```

`serial`을 제외하고는 모두 원래대로 돌아왔죠?

## 상태 파일 관리

보시다시피 별다른 설정이 없다면, 상태 파일은 현재 Terraform 명령을 실행하는 폴더에 생성됩니다. 실습을 위해 간단하게 테스트를 하는 경우라면 문제가 없겠지만, 계속 상태 파일을 이렇게 사용하는 건 매우 위험합니다. 그 이유는!

- output등의 민감한 정보가 상태 파일에 기록되기 때문에 파일에 접근할 수 있다면, 민감한 정보를 보호할 수 없습니다.
- 동일한 리소스에 대해, 여러 장소(컴퓨터)에서 Terraform을 실행해야 하는 경우 각 장소마다 동일한 상태 파일을 가지고 있어야 합니다. 그래서 git등의 버전 관리 시스템으로 상태 파일을 관리할 수도 있지만, `commit이나 pull을 까먹는다면 무슨 일이 벌어질까요?`

그래서 실제 Terraform을 사용하려면 상태 파일은 로컬 컴퓨터가 아닌, git같은 버전 관리 시스템도 아닌 어딘가 안전하고 항상 접근 가능한 곳에 저장되어야 합니다. 그리고 가능하면 암호화되어야 겠죠! 이 글 서두에 보셨던 `backend`가 바로 이런 상태 파일을 어딘가에서 안전하게 잘 관리해주는 역할을 합니다. 우리는 지금 `AWS`를 사용하고 있죠? 그렇다면 이제 `S3`를 통해 `backend`를 구성해보겠습니다.

## S3 backend 구성

`S3`로 `backend`를 구성하려면 우선 `S3`를 생성해야겠죠? 일단 Terraform으로 `S3`를 생성하고 그 이후에 `backend`를 구성해서 상태를 `backend`로 옮기겠습니다. 새로 폴더를 만들고 `main.tf`를 다음과 같이 작성합니다.

```terraform
// AWS 프로바이더 설정
provider "aws" {
  // 아시아 태평양(서울) 리전
  region = "ap-northeast-2"
}

# S3 버킷 생성
resource "aws_s3_bucket" "terraform_state" {
  bucket = "terraform-state-example-20231118"

  # 실수로 버킷이 삭제되지 않도록 하려면 아래 주석을 해제하세요.
  # lifecycle {
  #   prevent_destroy = true
  # }
}

# 버킷 버전 관리 활성화(상태 파일의 히스토리를 관리하기 위함)
resource "aws_s3_bucket_versioning" "enabled" {
  bucket = aws_s3_bucket.terraform_state.id

  versioning_configuration {
    status = "Enabled"
  }
}

# S3에 저장될 때 암호화 하도록 설정
resource "aws_s3_bucket_server_side_encryption_configuration" "default" {
  bucket = aws_s3_bucket.terraform_state.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# 모든 public access를 차단(실수로 버킷이 공개되지 않도록 하기 위함)
resource "aws_s3_bucket_public_access_block" "public_access" {
  bucket = aws_s3_bucket.terraform_state.id

  block_public_acls   = true
  block_public_policy = true
  ignore_public_acls  = true
  restrict_public_buckets = true
}

# 여러 사람이 동시에 상태 파일을 수정하지 못하도록 DynamoDB를 사용한 락 설정
resource "aws_dynamodb_table" "terraform_locks" {
  name = "terraform_locks"
  billing_mode = "PAY_PER_REQUEST"
  hash_key = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }
}
```

우선 `aws_s3_bucket` 리소스를 통해 `S3` 버킷을 생성하는데요, 이 버킷에 상태 파일을 저장하게 됩니다. 그리고 필요한 경우 상태 파일의 히스토리를 버전 별로 관리할 수 있도록 버전 관리를 활성화 합니다. 버전 관리를 사용하도록 설정한다면 뭔가 상태 파일에 문제가 있을 경우 이전 버전으로 복구할 수 있겠죠?

그리고 상태 파일은 암호화되어 저장되도록 설정했고, 실수로 버킷이 공개되지 않도록 모든 public access를 차단했습니다. 마지막으로 여러 사람이 동시에 상태 파일을 수정하지 못하도록 DynamoDB를 사용한 락을 설정했습니다. 이제 `tf init`을 실행해봅시다.

```bash
> tf init
...
Terraform has been successfully initialized!
...
```

그리고 `tf plan`으로 변경될 리소스를 확인합니다.

```bash
> tf plan
...
Plan: 5 to add, 0 to change, 0 to destroy.
...
```

이어서 `tf apply`로 리소스를 생성합니다.

```bash
> tf apply
...
Apply complete! Resources: 5 added, 0 changed, 0 destroyed.
```

> 만약 이전글에서 사용했던 것처럼 제한된 IAM 사용자 권한을 사용하는 경우 "sts:GetCallerIdentity","s3:CreateBucket","dynamodb:CreateTable","dynamodb:DescribeTable","s3:ListBucket","s3:GetBucketPolicy","s3:GetBucketAcl","s3:GetBucketCORS","s3:GetBucketWebsite","s3:GetBucketVersioning","s3:GetAccelerateConfiguration","s3:GetBucketRequestPayment","s3:GetBucketLogging","s3:GetLifecycleConfiguration","s3:GetReplicationConfiguration","s3:GetEncryptionConfiguration","s3:GetBucketObjectLockConfiguration","s3:GetBucketTagging","s3:PutBucketVersioning","dynamodb:DescribeContinuousBackups","s3:PutBucketPublicAccessBlock","s3:PutEncryptionConfiguration","dynamodb:DescribeTimeToLive","dynamodb:ListTagsOfResource","s3:GetBucketPublicAccessBlock","s3:GetObject","s3:GetObjectVersion","dynamodb:GetItem","dynamodb:PutItem","dynamodb:DeleteItem" 등의 권한이 추가로 필요합니다.

`S3`는 잘 생성되었지만, 여전히 `terraform.tfstate`파일은 현재 폴더에 그대로 남아있습니다. 파일 내용을 확인해보면, 방금 생성한 `S3`의 정보가 담겨있습니다. 아직 `backend`설정을 하지 않았기 때문에 여전히 상태 파일은 로컬 컴퓨터에서 관리하고 있는 거죠. 그렇다면 `backend`를 `S3`로 설정하겠습니다. `main.tf`에서 기존에 작성된 코드 위에 다음 코드를 추가합니다.

```terraform
# Terraform 설정
terraform {
  # 필요한 프로바이더 목록과 버전 정보
  required_providers {
    aws = {
      source = "hashicorp/aws"
      version = "~> 5.0" # 5.x 버전 사용
    }
  }

  # 상태 파일을 S3에 저장하도록 설정
  backend "s3" {
    bucket = "terraform-state-example-20231118"
    # 상태 파일이 저장될 경로
    key    = "example/terraform.tfstate"
    region = "ap-northeast-2"

    # 상태 파일을 락으로 사용할 DynamoDB 테이블
    dynamodb_table = "terraform_locks"
    # 상태 파일을 암호화
    encrypt        = true
  }
}

...
```

Terraform 블록을 선언하고 사용할 프로바이더와 버전 정보, 그리고 `S3`를 `backend`로 사용하도록 설정했습니다. 이제 `tf init`을 실행하면 로컬에 저장된 상태 파일을 새로 지정된 `S3`로 이동하게 됩니다.

```bash
> tf init
Initializing the backend...
Acquiring state lock. This may take a few moments...
Do you want to copy existing state to the new backend?
  Pre-existing state was found while migrating the previous "local" backend to the
  newly configured "s3" backend. No existing state was found in the newly
  configured "s3" backend. Do you want to copy this state to the new "s3"
  backend? Enter "yes" to copy and "no" to start with an empty state.

  Enter a value: yes

Releasing state lock. This may take a few moments...

Successfully configured the backend "s3"! Terraform will automatically
use this backend unless the backend configuration changes.
...
```

로컬에 저장된 상태를 `S3`로 복사하겠냐는 질문이 나오는데, `yes`를 입력하면 로컬에 저장된 상태 파일이 `S3`로 복사됩니다. 그리고 `terraform.tfstate` 파일이 남아있긴 하지만, 내용을 확인 해보면 텅텅 비어 있습니다. 이제 삭제해도 문제가 없겠죠! 참고로 `backend`에서 상태 파일을 로컬로 다운로드 하고 싶다면 `terraform state pull` 명령을 사용하면 됩니다.

> `iamlive`를 실행 중이라면 `tls: failed to verify certificate: x509` 오류나 `request send failed, Post "https://sts.ap-northeast-2.amazonaws.com/": proxyconnect tcp: dial tcp 127.0.0.1:80` 등의 오류가 발생할 수 있습니다. `iamlive`의 실행을 위해 선언해둔 환경 변수 때문인데요, 해당 환경 변수를 주석처리하고 터미널 세션을 다시 시작하거나, `unset` 등의 명령어로 현재 터미널 세션에서 환경 변수를 제거할 수 있습니다.

그렇다면 이제 간단한 EC2 인스턴스를 추가해서 상태가 잘 관리되는지 확인하겠습니다. `main.tf`에 다음 코드를 추가합니다.

```terraform
...

resource "aws_instance" "example" {
  // ap-northeast-2 리전의 Ubuntu 20.04 AMI
  ami           = "ami-0c6e5afdd23291f73"
  instance_type = "t2.micro"
}
```

`tf plan`과 `tf apply`를 실행하여 동작을 확인합니다.

```bash
> tf plan
...
Plan: 1 to add, 0 to change, 0 to destroy.
...

> tf apply
...
Apply complete! Resources: 1 added, 0 changed, 0 destroyed.
```

더 이상 로컬에 상태 파일이 생성되지 않지만, 원격의 `S3`에서 상태를 읽어와서 잘 관리하고 있네요!

# 마치며

이제 우리는 상태를 로컬 파일이 아니라 `S3`를 통해 안전하고 안정적으로 그리고 어디서나 활용할 수 있도록 변경할 수 있습니다. 하지만 여전히 불안한 점은 남아있습니다. EC2를 추가하면서 plan, apply를 실행해보신 분들은 아시겠지만, 여전히 이미 생성된 `S3`에 대한 상태를 매번 비교하고 있습니다. 하나의 상태에 많은 리소스가 연결되면 이렇게 아주 작은 부분만을 바꿀 때도 전체 상태와 비교를 하게 됩니다. 그리고 아무리 리소스 파일이 공유된 상태라고는 하지만, 여러 사람이 동시에 서로 다른 부분에서 코드를 변경하면 예상하지 어려운 상황이 발생하겠죠. 그래서 상태 파일은 가능하면 작게 유지하는 게 좋습니다. 다음 시간에는 상태 파일을 작게 유지하는 방법에 대해 알아보겠습니다.

# 참고자료

- [Terraform Up and Running, 3rd Edition](https://www.terraformupandrunning.com/)

# License

[저작자표시-비영리-변경금지 2.0 대한민국 (CC BY-NC-ND 2.0 KR)](https://creativecommons.org/licenses/by-nc-nd/2.0/kr/)
