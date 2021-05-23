package git

import (
	"fmt"
	"strings"
	"time"

	"github.com/jojomi/go-script/v2"
)

type Commit struct {
	hash string

	repository *Repository
}

func newCommit(repository *Repository, hash string) *Commit {
	commit := Commit{
		repository: repository,
		hash:       hash,
	}
	return &commit
}

func (c *Commit) GetHash() string {
	return c.hash
}

func (c *Commit) GetFullHash() (string, error) {
	return c.getCommitValue("%H")
}

func (c *Commit) GetShortHash() (string, error) {
	return c.getCommitValue("%h")
}

func (c *Commit) GetMessage() (string, error) {
	return c.getCommitValue("%s")
}

func (c *Commit) GetBody() (string, error) {
	return c.getCommitValue("%b")
}

func (c *Commit) GetAuthorName() (string, error) {
	return c.getCommitValue("%aN")
}

func (c *Commit) GetAuthorEmail() (string, error) {
	return c.getCommitValue("%aE")
}

func (c *Commit) GetAuthorDate() (time.Time, error) {
	dateString, err := c.getCommitValue("%aD")
	if err != nil {
		return time.Time{}, err
	}
	const format = "Mon, 2 Jan 2006 15:04:05 -0700"
	d, err := time.Parse(format, dateString)
	if err != nil {
		return time.Time{}, err
	}
	return d, nil
}

func (c *Commit) GetAuthorDateRelative() (string, error) {
	return c.getCommitValue("%ar")
}

func (c *Commit) Equals(otherCommit *Commit) bool {
	return c.GetHash() == otherCommit.GetHash()
}

func (c *Commit) String() string {
	return "Commit " + c.GetHash()
}

func (c *Commit) getCommitValue(param string) (string, error) {
	// https://stackoverflow.com/a/31448684
	// https://git-scm.com/docs/git-log
	// git show bc1affe7b992fc406724bfe04093016f020d7c0a --pretty=format:"%aN" --no-patch
	command := script.LocalCommandFrom("git show")
	command.AddAll(c.GetHash(), "--pretty=format:"+param, "--no-patch")

	pr, err := c.repository.Execute(command)
	if !pr.Successful() {
		err = fmt.Errorf("could not get log data for commit %s", c.GetHash())
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(pr.Output()), nil
}
