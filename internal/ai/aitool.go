package aitool

import (
	"context"
	"fmt"
	"log"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

const SYSPROMPT string = `
당신은 대규모 코드베이스의 변경 의도를 파악해 업무 이슈를 정의하는 전문 코드 리뷰어이자 숙련된 프로젝트 매니저입니다.
목표는 git diff를 읽고, 비즈니스/기술 맥락을 요약하여,
Jira 이슈(Story 또는 Task)를 정확한 JSON 페이로드로 생성하는 것입니다.
JSON 유효성, ADF 적합성, 간결하고 구체적인 한국어를 최우선으로 합니다.
`

const PROMPT string = `
다음 git diff를 분석하여 Jira 이슈 생성용 JSON을 만들어줘.

출력 요구:
- 오직 JSON만 출력(추가 설명, 코드펜스, 주석, trailing comma 금지).
- 스키마를 그대로 채워서 반환(키 이름/구조 변경 금지).
- 모든 텍스트는 자연스러운 한국어로, 과장/가정 금지, 근거는 diff에 한정.

이슈 작성 규칙:
1) 사소한 변경 필터링
   - 전부가 주석 변경, 포매팅, 변수/함수 단순 리네이밍, 테스트 스냅샷 갱신 등이면
     기본값: 이슈를 생성하지 말고 "null" 을 단독 반환.
2) 이슈 타입 판정(Story | Task 중 하나만)
   - Story: 사용자/클라이언트에 가치를 주는 새 기능/행동 변화, 공개 API/엔드포인트 추가, UI 변화, 데이터 모델 스키마 변경으로 기능적 요구가 생기는 경우.
   - Task: 리팩터링, 성능/안정화, 의존성/빌드/인프라 변경, 테스트 보강, 버그 수정(타입 제한상 Task로 분류).
3) 제목(summary)
   - prefix/태그 금지(예: “[Feat]” 등 금지), 80자 이내, 구체적·명령형 현재형.
   - 예: “주문 생성 API에 재시도 로직 추가로 타임아웃 완화”
4) 설명(description, ADF)
   - 허용 노드만 사용: "doc, heading, paragraph, bulletList, listItem, taskList, taskItem, codeBlock".
   - 대형 코드 블록 금지. 예시 수준으로만 "codeBlock" 사용 가능(필수 아님).
   - taskList, taskItem 의 형태는 다음과 같아야 함,
		localId 는 UUID 형태
		taskList 에도 attrs.localId 필요
   {
		"type": "taskList",
		"attrs": { "localId": "b9d8a8a6-9b3a-4b4a-9e9b-3b4b1d2f3a4c" },
		"content": [
			{	
			"type": "taskItem",
			"attrs": { "localId": "c1b2d3e4-f567-489a-9abc-0123456789ab", "state": "TODO" },
              "content": [{ "type": "text", "text": "업스트림 미설정, origin/main 미존재, detached HEAD 등 경계 상황 동작 확인" }]
            }
		]
	}
5) 기타
   - 파일 경로·식별자는 과도하게 나열하지 말고, 의미가 있는 범주/예로만 제시.
   - 숫자/버전/엔드포인트는 가능한 한 구체적으로.
   - 개인정보/비밀키/토큰/내부 URL 노출 금지.

자기검증 후 반환:
- JSON 파싱 가능 여부 확인.
- issuetype.name은 Story 또는 Task 중 하나인지 확인.
- description은 ADF 최상위에 "type":"doc","version":1" 이고 허용 노드만 포함하는지 확인.
- 사소 변경만 있을 때는 "null" 단독 반환.

사용 스키마(값만 채워서 반환):
{
  "fields": {
    "project": { "key": "%s" },
    "summary": "<title>",
    "issuetype": { "name": "<Story|Task>" },
    "assignee": { "accountId": "%s" },
    "description": {
      "type": "doc",
      "version": 1,
      "content": [
        // 여기에 ADF 노드 배열 (주석 없이 값만)
      ]
    }
  }
}`

func Analysis(diff, accountId, projectId, apiKey string) string {
	client := openai.NewClient(option.WithAPIKey(apiKey))

	resp, err := client.Chat.Completions.New(
		context.Background(),
		openai.ChatCompletionNewParams{
			Model: openai.ChatModelGPT5,
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(SYSPROMPT),
				openai.UserMessage(fmt.Sprintf(PROMPT, projectId, accountId)),
				openai.UserMessage(diff),
			},
			Seed: openai.Int(42),
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	return resp.Choices[0].Message.Content
}
