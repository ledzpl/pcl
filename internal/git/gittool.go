package gittool

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func GetBranches() []string {

	repo, err := git.PlainOpen(".")
	if err != nil {
		log.Fatalf("저장소를 열 수 없습니다: %v", err)
	}

	branches, err := repo.Branches()
	if err != nil {
		log.Fatalf("브랜치를 가져올 수 없습니다: %v", err)
	}

	var result []string
	err = branches.ForEach(func(ref *plumbing.Reference) error {
		result = append(result, ref.Name().Short())
		return nil
	})

	if err != nil {
		log.Fatalf("브랜치를 순회하는 중 오류: %v", err)
	}

	return result
}

func Diff(src string) string {

	base := detectUpstream(src)

	diff, err := runGit("diff",
		"--no-color",
		"--no-ext-diff",
		"-U0",
		"-M",
		"-w",
		base)

	if err != nil {
		log.Fatal(err)
	}

	return diff
}

func detectUpstream(src string) string {
	if s, err := runGit("merge-base", "--fork-point", src); err == nil && s != "" {
		return s
	}
	return "HEAD"
}

func runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errb.String())
		if msg != "" {
			return "", fmt.Errorf("%v: %s", err, msg)
		}
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}
