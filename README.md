# pcl

코드 변경(diff)을 분석해 Jira Cloud에 적합한 이슈를 자동으로 생성해 주는 Go 기반 CLI 도구입니다. 로컬 Git 저장소의 변경 사항을 읽어 OpenAI GPT-5 모델에 전달하고, 반환된 JSON 페이로드를 Jira REST API `/rest/api/3/issue` 엔드포인트에 바로 업로드합니다.

## 주요 특징
- 변경 의도 파악: `git diff` 정보를 분석해 Story/Task 유형을 선택하고, 요약과 ADF 형식 설명을 자동 생성합니다.
- 인터랙티브 워크플로: 현재 저장소의 브랜치를 탐색해 기준 브랜치를 고르는 프롬프트 UI를 제공합니다.
- Jira 연동 자동화: Atlassian Cloud 계정의 Account ID를 조회한 뒤, 즉시 이슈를 생성합니다.
- 안전 장치: 의미 있는 변경이 없으면 `null`을 반환해 불필요한 이슈 생성을 막습니다.

## 요구 사항
- Go 1.25 이상
- Git 저장소(로컬 변경 사항이 존재해야 합니다)
- OpenAI API 키 및 Jira Cloud API 접근 권한

## 설치
```bash
go install github.com/ledzpl/pcl@latest
```
또는 저장소를 클론한 뒤 직접 빌드할 수 있습니다.
```bash
git clone https://github.com/ledzpl/pcl.git
cd pcl
go build -o pcl
```

## 환경 변수 설정
다음 환경 변수가 모두 설정되어 있어야 실행할 수 있습니다.

| 이름 | 설명 |
| --- | --- |
| `OPENAI_API_KEY` | OpenAI GPT-5 채팅 컴플리션을 호출할 때 사용하는 API 키 |
| `JIRA_API_KEY` | Atlassian Cloud 개인용 토큰(Email + API Token 조합) |
| `JIRA_HOST` | Jira 호스트 URL (예: `https://your-domain.atlassian.net`) |
| `JIRA_EMAIL` | Jira Cloud 계정 이메일 |
| `JIRA_PROJECT` | 이슈를 생성할 프로젝트 키 (예: `PCL`) |

예시:
```bash
export OPENAI_API_KEY="sk-..."
export JIRA_API_KEY="your-jira-token"
export JIRA_HOST="https://your-domain.atlassian.net"
export JIRA_EMAIL="you@example.com"
export JIRA_PROJECT="PCL"
```

## 사용 방법
1. 로컬 저장소에서 작업 브랜치를 checkout하고 커밋되지 않은 변경 또는 추가 커밋이 있는지 확인합니다.
2. 터미널에서 `pcl`을 실행합니다.
3. 프롬프트에서 기준 브랜치를 선택합니다. 선택한 브랜치와 현재 `HEAD` 사이의 diff가 자동 계산됩니다(`git diff --no-color --no-ext-diff -U0 -M -w <base>...HEAD`).
4. 스피너가 표시되는 동안 OpenAI가 diff를 해석하고 Jira용 JSON을 생성합니다.
5. 생성된 JSON을 그대로 Jira REST API에 전송하여 이슈를 생성합니다.
6. 작업이 성공하면 Jira에서 새 이슈를 확인할 수 있습니다.

> diff가 전부 주석/포맷 수정 등 사소한 변경으로 판단되면 이슈를 만들지 않고 종료합니다.

## 동작 구성 요소
- `internal/git`: go-git과 로컬 `git` 명령을 이용해 브랜치 목록과 diff를 가져옵니다.
- `internal/ai`: OpenAI Chat Completions(`gpt-5`)을 호출해 Jira 이슈 스키마에 맞는 JSON을 생성합니다.
- `internal/jira`: Resty HTTP 클라이언트로 `/myself`에서 Account ID를 조회하고 `/issue`에 JSON을 POST합니다.

## 개발 & 테스트
개발자 환경에서 동작을 검증할 때는 다음 명령을 참고하세요.
```bash
go test ./...
```
테스트 커버리지는 향후 추가 예정이며, PR 전에는 최소한 기본 빌드(`go build ./...`)가 성공하는지 확인해 주세요.

## 문제 해결
- **환경 변수 미설정**: 실행 즉시 에러 메시지를 출력하고 종료합니다. 위 표의 변수들이 올바르게 설정되었는지 확인하세요.
- **계정 권한 부족**: Jira API 응답이 401/403일 경우 토큰과 이메일, 프로젝트 키를 재검증합니다.
- **브랜치 감지 실패**: 저장소 루트에서 실행했는지, Git 저장소가 초기화되어 있는지 확인합니다.

## 라이선스
MIT License. 자세한 내용은 `LICENSE` 파일을 참고하세요.
