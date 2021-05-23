package git

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/jojomi/go-script/v2"
)

// cache
var gitVersion *semver.Version

func GetGitVersion() (*semver.Version, error) {
	if gitVersion != nil {
		return gitVersion, nil
	}

	sc := script.NewContext()
	sc.MustCommandExist("git")
	command := script.LocalCommandFrom("git version")
	pr, err := sc.ExecuteSilent(command)
	if !pr.Successful() {
		err = fmt.Errorf("version command failed")
	}
	if err != nil {
		return &semver.Version{}, err
	}
	v := pr.Output()
	v = strings.Replace(v, "git version", "", 1)
	v = strings.TrimSpace(v)

	version, err := semver.NewVersion(v)

	if err == nil {
		gitVersion = version
	}
	return version, err
}

func MustGetGitVersion() *semver.Version {
	version, err := GetGitVersion()
	if err != nil {
		panic(err)
	}
	return version
}

var regexpStarredBranchList = regexp.MustCompile(`^\s*\*?\s*([^ ]*)`)
var regexpBranchList = regexp.MustCompile(`^[0-9a-f]{5,40}\s+(.*)$`)

func parseStarredBranchList(input string) ([]string, error) {
	var (
		branches   = make([]string, 0)
		branchName string
		matches    []string
	)
	for _, line := range strings.Split(strings.TrimSpace(input), "\n") {
		matches = regexpStarredBranchList.FindStringSubmatch(line)
		if len(matches) < 2 || matches[1] == "" {
			return nil, fmt.Errorf("invalid line format in branch list")
		}
		branchName = matches[1]
		branches = append(branches, branchName)
	}
	return branches, nil
}

func parseBranchList(input string) ([]string, error) {
	var (
		branches   = make([]string, 0)
		branchName string
		matches    []string
	)
	for _, line := range strings.Split(strings.TrimSpace(input), "\n") {
		matches = regexpBranchList.FindStringSubmatch(line)
		if len(matches) < 2 || matches[1] == "" {
			return nil, fmt.Errorf("invalid line format in branch list")
		}

		branchName = strings.TrimSpace(strings.Replace(matches[1], "refs/heads/", "", 1))
		branches = append(branches, branchName)
	}
	return branches, nil
}
