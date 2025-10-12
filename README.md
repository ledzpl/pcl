# pcl

로컬 Git 저장소의 변경 사항을 분석해 Jira Cloud 이슈나 Conventional Commits 형식의 커밋 메시지를 자동으로 생성하는 Go CLI입니다. 브랜치 선택부터 diff 추출, OpenAI GPT-5 호출, Jira REST API 연동까지 한 번에 수행하는 워크플로를 제공합니다.

> Go를 처음 학습하면서 만든 실험용 프로젝트입니다. 구조나 패턴은 참고용으로만 봐 주세요.

## 핵심 기능
- **로컬 변경 감지**: 현재 저장소의 브랜치 목록을 `promptui`로 보여주고, 선택한 기준 브랜치에 대해 `git merge-base --fork-point`를 활용한 diff(`git diff -U0 -M -w`)를 수집합니다.
- **AI 분석 파이프라인**: 수집한 diff를 OpenAI GPT-5 Chat Completions에 전달해 Jira 이슈 JSON 또는 Conventional Commits 커밋 메시지를 생성합니다. `internal/ai`에서 프롬프트와 출력 형식을 관리합니다.
- **Jira 자동화**: `internal/jira`가 Jira Cloud 계정의 Account ID를 조회한 뒤, Resty HTTP 클라이언트를 이용해 `/rest/api/3/issue`에 JSON을 POST합니다.
- **안전 장치**: diff가 비어 있으면 바로 종료하고, 사소한 변경만 있는 경우 AI 응답으로 `null`이 돌아오도록 프롬프트에서 제한합니다.
- **인터랙티브 UX**: 프롬프트 기반 메뉴와 스피너를 제공해 진행 상태를 시각적으로 보여줍니다.

## 동작 흐름
1. `pcl` 실행 → 로컬 저장소 브랜치 중 기준 브랜치를 선택합니다.
2. diff가 없으면 `"비교할 변경점이 없습니다."`로 종료됩니다.
3. "Jira 이슈 생성" 또는 "커밋 메시지 생성" 중 하나를 고릅니다.
4. 선택에 따라 설정값을 검사합니다. (Jira 이슈 생성은 OpenAI/Jira 관련 키 모두 필요, 커밋 메시지는 OpenAI 키만 필요)
5. 스피너가 돌면서 GPT-5가 diff를 분석합니다.
6. 결과를 표준 출력으로 제공합니다. Jira 이슈 생성은 성공 시 동일한 JSON을 두 번 출력합니다(생성 요청 전에 검토용, 요청 후 확인용).

## 설치
```bash
go install github.com/ledzpl/pcl@latest
```
또는 저장소를 직접 빌드할 수 있습니다.
```bash
git clone https://github.com/ledzpl/pcl.git
cd pcl
go build -o pcl
```

## 실행 옵션
- `-config`: 설정 파일 경로를 지정합니다. 기본값은 실행 디렉터리의 `config.json`입니다.

```bash
pcl -config ./config.json
```

## 설정 파일 (`config.json`)

| 키 | 설명 | 필수 조건 |
| --- | --- | --- |
| `openai_api_key` | GPT-5 Chat Completions 호출에 사용하는 OpenAI API 키 | 모든 기능 |
| `jira_api_key` | Jira Cloud Personal Access Token | Jira 이슈 생성 |
| `jira_host` | Jira 사이트 URL (예: `https://your-domain.atlassian.net`) | Jira 이슈 생성 |
| `jira_email` | Atlassian 계정 이메일 | Jira 이슈 생성 |
| `jira_project` | 이슈를 생성할 프로젝트 키 (예: `PCL`) | Jira 이슈 생성 |

예시:

```json
{
  "openai_api_key": "sk-...",
  "jira_api_key": "your-jira-token",
  "jira_host": "https://your-domain.atlassian.net",
  "jira_email": "you@example.com",
  "jira_project": "PCL"
}
```

## 패키지 구조
- `internal/git`: go-git을 활용해 브랜치 목록을 가져오고, 로컬 `git` 명령을 호출해 diff를 생성합니다.
- `internal/ai`: Jira 이슈용/커밋 메시지용 프롬프트와 OpenAI 클라이언트를 캡슐화합니다.
- `internal/jira`: Account ID 조회와 이슈 생성(기본 인증 헤더 포함)을 담당합니다.
- `internal/config`: JSON 설정 파일을 로드하고, Jira/AI 실행 전 필수 키의 존재를 검증합니다.
- `main.go`: CLI 진입점으로, 사용자 인터랙션과 전체 워크플로를 연결합니다.

## 문제 해결
- **필수 키 누락**: 실행 즉시 `"config: missing required keys"` 오류가 발생합니다. 설정 파일을 다시 확인하세요.
- **Jira API 실패**: HTTP 401/403 응답은 토큰·이메일·호스트 URL을 재검증해야 한다는 의미입니다. 응답 본문이 있으면 오류 메시지에 포함됩니다.
- **diff 추출 실패**: Git 저장소 루트에서 실행했는지, 기준 브랜치가 로컬에 존재하는지 확인하세요.
- **OpenAI 오류**: 네트워크 제한이나 모델 이름이 잘못된 경우 로그에 `log.Fatal`로 출력됩니다.

## 개발 참고
- 테스트: `go test ./...`
- 린트/포맷: 별도 설정은 없지만 `gofmt`, `golangci-lint` 등을 수동으로 사용할 수 있습니다.

## 라이선스
MIT License — 자세한 내용은 `LICENSE` 파일을 참고하세요.
