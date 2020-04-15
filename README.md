# key-letter
> 가볍고 안전한 키 교환 서비스

사용자 간 키(서비스 목적에 따른 문자열 값)를 교환할 때 안전하게 전달하기 위한 서비스 

## 클라이언트

- [Android]() (WIP)

## 사용한 기술 및 라이브러리

- [gRPC](https://grpc.io/)
- [pquerna/otp](https://github.com/pquerna/otp)
- [go-redis/redis](https://github.com/go-redis/redis)
- [grpc-ecosystem/go-grpc-middleware](https://github.com/grpc-ecosystem/go-grpc-middleware)
- [go-yaml/yaml](https://github.com/go-yaml/yaml)

## 필요 조건

- Go 언어 1.14 혹은 그 이상
- Redis 서버
    - 직접 구축하여도 되나, [RedisLabs](https://redislabs.com/) 에서 무료로 이용할 수 있습니다.

## 설치 방법

1. 이 리포지토리를 클론합니다.

```sh
git clone https://github.com/namhyun-gu/key-letter.git
```

2. 클론한 리포지토리에 이동하고, 서비스에 필요한 Go 패키지를 설치합니다.

```sh
cd key-letter
go get -u
```

3. 서비스에 필요한 설정을 루트 디렉토리의 config.yaml에 작성합니다.

```yaml
port: *(서비스 포트)
redis:
  addr: *(Redis 주소)
  password: *(Redis 비밀번호)
opts:
  issuer: *(코드 생성을 위한 발급자 정보)
  period: (코드 유효 시간(초). 기본 30초)
  digits: (코드 길이 (6 혹은 8). 기본 6)
  algorithm: (코드 생성 알고리즘 (SHA1, SHA256, SHA512, MD5). 기본 SHA1)
```
**\* 는 필수 설정입니다.**

4. 서비스를 실행합니다.

```sh
go run main.go
```

## 키 전달 과정

1. 사용자 A가 원하는 키를 서비스에 전송하여 코드를 전달 받습니다.

2. 사용자 A가 받은 키를 원하는 사용자 B에게 전달하여 키를 받을 수 있도록 합니다.

3. 사용자 B는 A로부터 받은 키를 통해 서비스에 키를 요청합니다.

4. 서비스는 사용자 B의 요청을 받아 사용자 A에게 전달 여부를 확인합니다.

5. 사용자 A가 허용한다면, 서비스가 사용자 B에게 키를 전달합니다.

## 정보

Namhyun, Gu – namhyun-gu@kakako.com

MIT 라이센스를 준수하며 ``LICENSE``에서 자세한 정보를 확인할 수 있습니다.