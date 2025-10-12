package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	aitool "github.com/ledzpl/pcl/internal/ai"
	"github.com/ledzpl/pcl/internal/config"
	gittool "github.com/ledzpl/pcl/internal/git"
	jira "github.com/ledzpl/pcl/internal/jira"

	"github.com/briandowns/spinner"
	"github.com/manifoldco/promptui"
)

const (
	actionCreateJiraIssue  = "Jira 이슈 생성"
	actionCommitMessage    = "커밋 메시지 생성"
	defaultSpinnerFinalMsg = "완료\n"
)

const ansiReset = "\033[0m"

const pcl string = `
██╗  ██╗███████╗██╗     ██╗      ██████╗        ██████╗  ██████╗██╗     ██╗
██║  ██║██╔════╝██║     ██║     ██╔═══██╗       ██╔══██╗██╔════╝██║     ██║
███████║█████╗  ██║     ██║     ██║   ██║       ██████╔╝██║     ██║     ██║
██╔══██║██╔══╝  ██║     ██║     ██║   ██║       ██╔═══╝ ██║     ██║     ╚═╝
██║  ██║███████╗███████╗███████╗╚██████╔╝▄█╗    ██║     ╚██████╗███████╗██╗
╚═╝  ╚═╝╚══════╝╚══════╝╚══════╝ ╚═════╝ ╚═╝    ╚═╝      ╚═════╝╚══════╝╚═╝

`

var rainbowColors = []string{
	"\033[31m",
	"\033[33m",
	"\033[32m",
	"\033[36m",
	"\033[34m",
	"\033[35m",
}

func main() {
	configPath := flag.String("config", "config.json", "path to configuration file")
	flag.Parse()

	printRainbowASCIIArt(pcl)

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config from %s: %v", *configPath, err)
	}

	branches := gittool.GetBranches()

	p := promptui.Select{Label: "Select base branch", Items: branches}
	_, result, err := p.Run()
	if err != nil {
		return
	}

	diff := gittool.Diff(result)
	if IsBlank(diff) {
		log.Fatalf("비교할 변경점이 없습니다.")
	}

	actionPrompt := promptui.Select{
		Label: "실행할 작업 선택",
		Items: []string{actionCreateJiraIssue, actionCommitMessage},
	}
	_, action, err := actionPrompt.Run()
	if err != nil {
		return
	}

	switch action {
	case actionCreateJiraIssue:
		if err := cfg.ValidateForJira(); err != nil {
			log.Fatalf("설정이 올바르지 않습니다: %v", err)
		}

		s := startSpinner("Jira 이슈 생성 중... ", "Jira 이슈 생성 완료\n")

		accountId, err := jira.GetAccountId(cfg.JiraEmail, cfg.JiraHost, cfg.JiraAPIKey)
		if err != nil {
			s.FinalMSG = ""
			stopSpinner(s)
			log.Fatalf("failed to fetch Jira account ID: %v", err)
		}

		airesponse := aitool.Analysis(diff, accountId, cfg.JiraProject, cfg.OpenAIAPIKey)

		if err := jira.CreateIssue(airesponse, cfg.JiraEmail, cfg.JiraHost, cfg.JiraAPIKey); err != nil {
			s.FinalMSG = ""
			stopSpinner(s)
			log.Fatalf("failed to create Jira issue: %v", err)
		}

		stopSpinner(s)
		fmt.Println(airesponse)

	case actionCommitMessage:
		if err := cfg.ValidateForAI(); err != nil {
			log.Fatalf("설정이 올바르지 않습니다: %v", err)
		}

		s := startSpinner("커밋 메시지 생성 중... ", "커밋 메시지가 준비되었습니다.\n")
		message := aitool.CommitMessage(diff, cfg.OpenAIAPIKey)
		stopSpinner(s)
		fmt.Println(message)
	default:
		log.Fatalf("지원하지 않는 작업입니다: %s", action)
	}
}

func IsBlank(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

func startSpinner(prefix, final string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[38], 300*time.Millisecond)
	s.Prefix = prefix
	s.HideCursor = true
	if final == "" {
		final = defaultSpinnerFinalMsg
	}
	s.FinalMSG = final
	s.Start()
	return s
}

func stopSpinner(s *spinner.Spinner) {
	if s == nil {
		return
	}
	s.Stop()
	s.HideCursor = false
}

func printRainbowASCIIArt(s string) {
	lines := strings.Split(strings.Trim(s, "\n"), "\n")
	for i, line := range lines {
		color := rainbowColors[i%len(rainbowColors)]
		fmt.Println(color + line + ansiReset)
	}
	fmt.Println()
}
