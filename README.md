# pcl

코드 변경(diff)을 분석해 Jira Cloud에 적합한 이슈를 자동으로 생성해 주는 Go 기반 CLI 도구입니다. 로컬 Git 저장소의 변경 사항을 읽어 OpenAI GPT-5 모델에 전달하고, 반환된 JSON 페이로드를 Jira REST API `/rest/api/3/issue` 엔드포인트에 바로 업로드합니다.

- go lang 을 처음 써보는 거라 언어에 익숙하지 않습니다. 좋은 go lang code 는 아닌 것 같으니 아이디어만 참고해주시면 좋겠습니다.
- 대부분 코드와 문서는 codex 가 자동 생성한 결과물 입니다,.

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

## 설정 파일
`pcl`은 실행 시 자격 증명을 JSON 설정 파일에서 읽어옵니다. 기본 경로는 현재 작업 디렉터리의 `config.json`이며, 다른 경로를 사용하려면 `-config` 플래그를 전달하세요.

```bash
pcl -config /path/to/config.json
```

설정 파일에 포함해야 하는 키는 다음과 같습니다.

| 키 | 설명 |
| --- | --- |
| `openai_api_key` | OpenAI GPT-5 채팅 컴플리션을 호출할 때 사용하는 API 키 |
| `jira_api_key` | Atlassian Cloud 개인용 토큰 |
| `jira_host` | Jira 호스트 URL (예: `https://your-domain.atlassian.net`) |
| `jira_email` | Jira Cloud 계정 이메일 |
| `jira_project` | 이슈를 생성할 프로젝트 키 (예: `DEVOPS`) |

예시 (`config.json`):

```json
{
  "openai_api_key": "sk-...",
  "jira_api_key": "your-jira-token",
  "jira_host": "https://your-domain.atlassian.net",
  "jira_email": "you@example.com",
  "jira_project": "PCL"
}
```

## 사용 방법
1. 로컬 저장소에서 작업 브랜치 변경 사항을 모두 커밋합니다.
2. 터미널에서 `pcl`을 실행합니다.
3. 프롬프트에서 기준 브랜치를 선택합니다. - main 에서 분리됐으면 main, dev 에서 분리됐으면 dev 선택
4. 스피너가 표시되는 동안 OpenAI가 diff를 해석하고 Jira용 JSON을 생성합니다.
5. 생성된 JSON을 그대로 Jira REST API에 전송하여 이슈를 생성합니다.
6. 작업이 성공하면 Jira에서 새 이슈를 확인할 수 있습니다.

> local 저장소 기준으로 diff 를 뜹니다.

> diff가 전부 주석/포맷 수정 등 사소한 변경으로 판단되면 이슈를 만들지 않고 종료합니다.

## 동작 구성 요소
- `internal/git`: go-git과 로컬 `git` 명령을 이용해 브랜치 목록과 diff를 가져옵니다.
- `internal/ai`: OpenAI Chat Completions(`gpt-5`)을 호출해 Jira 이슈 스키마에 맞는 JSON을 생성합니다.
- `internal/jira`: Resty HTTP 클라이언트로 `/myself`에서 Account ID를 조회하고 `/issue`에 JSON을 POST합니다.

## 문제 해결
- **설정 파일 누락/오타**: 설정 파일 경로가 맞는지, 필수 키(`openai_api_key`, `jira_api_key`, `jira_host`, `jira_email`, `jira_project`)에 값이 채워졌는지 확인하세요.
- **계정 권한 부족**: Jira API 응답이 401/403일 경우 토큰과 이메일, 프로젝트 키를 재검증합니다.
- **브랜치 감지 실패**: 저장소 루트에서 실행했는지, Git 저장소가 초기화되어 있는지 확인합니다.

## 라이선스
MIT License. 자세한 내용은 `LICENSE` 파일을 참고하세요.
