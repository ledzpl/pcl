package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	aitool "pcl/internal/ai"
	gittool "pcl/internal/git"
	jira "pcl/internal/jira"

	"github.com/briandowns/spinner"
	"github.com/manifoldco/promptui"
)

const pcl string = `
██╗  ██╗███████╗██╗     ██╗      ██████╗        ██████╗  ██████╗██╗     ██╗
██║  ██║██╔════╝██║     ██║     ██╔═══██╗       ██╔══██╗██╔════╝██║     ██║
███████║█████╗  ██║     ██║     ██║   ██║       ██████╔╝██║     ██║     ██║
██╔══██║██╔══╝  ██║     ██║     ██║   ██║       ██╔═══╝ ██║     ██║     ╚═╝
██║  ██║███████╗███████╗███████╗╚██████╔╝▄█╗    ██║     ╚██████╗███████╗██╗
╚═╝  ╚═╝╚══════╝╚══════╝╚══════╝ ╚═════╝ ╚═╝    ╚═╝      ╚═════╝╚══════╝╚═╝

`

func main() {

	fmt.Print(pcl)

	openai_key, _ := os.LookupEnv("OPENAI_API_KEY")
	if IsBlank(openai_key) {
		log.Fatalf("I need OPENAI_API_KEY, You should export OPENAI_API_KEY")
	}

	jira_key, _ := os.LookupEnv("JIRA_API_KEY")
	if IsBlank(jira_key) {
		log.Fatalf("I need JIRA_API_KEY, You should export JIRA_API_KEY")
	}

	jira_host, _ := os.LookupEnv("JIRA_HOST")
	if IsBlank(jira_host) {
		log.Fatalf("I need JIRA_HOST, You should export JIRA_HOST")
	}

	jira_email, _ := os.LookupEnv("JIRA_EMAIL")
	if IsBlank(jira_email) {
		log.Fatalf("I need JIRA_EMAIL, You should export JIRA_EMAIL")
	}

	jira_project, _ := os.LookupEnv("JIRA_PROJECT")
	if IsBlank(jira_project) {
		log.Fatalf("I need JIRA_PROJECT, You should export JIRA_PROJECT")
	}

	//

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

	s := spinner.New(spinner.CharSets[38], 300*time.Millisecond)
	s.Prefix = "Working... "
	s.HideCursor = true
	s.Start()

	accountId := jira.GetAccountId(jira_email, jira_host, jira_key)

	airesponse := aitool.Analysis(diff, accountId, jira_project)

	jira.CreateIssue(airesponse, jira_email, jira_host, jira_key)

	s.Stop()
}

func IsBlank(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}
