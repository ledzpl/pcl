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

const pcl string = `
██╗  ██╗███████╗██╗     ██╗      ██████╗        ██████╗  ██████╗██╗     ██╗
██║  ██║██╔════╝██║     ██║     ██╔═══██╗       ██╔══██╗██╔════╝██║     ██║
███████║█████╗  ██║     ██║     ██║   ██║       ██████╔╝██║     ██║     ██║
██╔══██║██╔══╝  ██║     ██║     ██║   ██║       ██╔═══╝ ██║     ██║     ╚═╝
██║  ██║███████╗███████╗███████╗╚██████╔╝▄█╗    ██║     ╚██████╗███████╗██╗
╚═╝  ╚═╝╚══════╝╚══════╝╚══════╝ ╚═════╝ ╚═╝    ╚═╝      ╚═════╝╚══════╝╚═╝

`

func main() {
	configPath := flag.String("config", "config.json", "path to configuration file")
	flag.Parse()

	fmt.Print(pcl)

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

	s := spinner.New(spinner.CharSets[38], 300*time.Millisecond)
	s.Prefix = "Working... "
	s.HideCursor = true
	s.FinalMSG = "Done"
	s.Start()
	defer s.Stop()

	accountId, err := jira.GetAccountId(cfg.JiraEmail, cfg.JiraHost, cfg.JiraAPIKey)
	if err != nil {
		s.FinalMSG = ""
		log.Fatalf("failed to fetch Jira account ID: %v", err)
	}

	airesponse := aitool.Analysis(diff, accountId, cfg.JiraProject, cfg.OpenAIAPIKey)

	if err := jira.CreateIssue(airesponse, cfg.JiraEmail, cfg.JiraHost, cfg.JiraAPIKey); err != nil {
		s.FinalMSG = ""
		log.Fatalf("failed to create Jira issue: %v", err)
	}
}

func IsBlank(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}
