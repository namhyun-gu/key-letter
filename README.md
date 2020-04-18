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
certfile: (SSL/TLS 이용 시 *.crt 파일 경로)
certkeyfile: (SSL/TLS 이용 시 *.pem / *.key 파일 경로)
timeout: (서비스 타임아웃. 기본 30초)
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

## 처리 과정

![처리 과정](http://www.plantuml.com/plantuml/png/NOv13i8W44NtdE8VG2_0mjHLTQidmFBLfgKe32QzlQ1eckucC-_z-KqK4oIvKHrybEqEPuPNtn4wJvF5m5dLLJuMHgFbn624wBobQXENeiQ9px8gAYxp5rf7xEE01uFh1R6-WNZSYhXgKelyQ36IuYBlyrxZUCKzQnKJsiq_M3LvI6vy0m00)

## 예외 처리


|Method|Status|Reason|Detail|
|---|---|---|---|
|/VerifyCode|SUCCESS|-||
||FAILED|AUTH_FAILED|코드가 유효하지 않음|
|||INTERNAL_ERR|내부 오류 발생|
|||REJECT_HOST|공유 요청이 호스트에 의해 거절됨|
|||NO_HOST_WAITED|기다리는 호스트가 없음. 공유 상태가 아님|
|||RESPONSE_TIMEOUT|타임 아웃으로 종료됨|

## 정보

Namhyun, Gu – namhyun-gu@kakako.com

## 라이센스

본 서비스는 MIT 라이센스를 준수하며 ``LICENSE``에서 자세한 정보를 확인할 수 있습니다.