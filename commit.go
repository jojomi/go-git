package git

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jojomi/go-script/v2"
)

type Commit struct {
	hash string

	repository *Repository
}

func newCommit(repository *Repository, hash string) (*Commit, error) {
	if !IsValidCommitHash(hash) {
		return nil, fmt.Errorf("invalid hash upon commit creation: %s", hash)
	}

	commit := Commit{
		repository: repository,
		hash:       hash,
	}
	return &commit, nil
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

func (c *Commit) GetPatchId() (string, error) {
	lc := &script.LocalCommand{}
	lc.AddAll("sh", "-c", fmt.Sprintf("git show %s | git patch-id", c.GetHash()))
	pr, err := c.repository.Execute(lc)
	if !pr.Successful() {
		err = fmt.Errorf("could not get patch-id for commit %s", c.GetHash())
	}
	if err != nil {
		return "", err
	}

	r := regexp.MustCompile(`^[^ ]+`)
	return r.FindString(pr.Output()), nil
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

// GetLocalBranches returns a list of all local branches containing this commit
func (c *Commit) GetLocalBranches() ([]*LocalBranch, error) {
	// git branch --contains <commit hash>
	localBranches := make([]*LocalBranch, 0)
	command := script.LocalCommandFrom("git branch --contains")
	hash, err := c.GetFullHash()
	if err != nil {
		return localBranches, nil
	}
	command.Add(hash)

	pr, err := c.repository.Execute(command)
	if !pr.Successful() {
		err = fmt.Errorf("could not list local branches containing commit %s", hash)
	}
	if err != nil {
		return localBranches, err
	}

	branchList, err := parseStarredBranchList(pr.Output())
	if err != nil {
		return nil, err
	}

	for _, branch := range branchList {
		localBranches = append(localBranches, newLocalBranch(c.repository, branch))
	}

	return localBranches, nil
}

// GetRemoteBranches returns a list of all branches on a given remote containing this commit
func (c *Commit) GetRemoteBranches(remote *Remote) ([]*RemoteBranch, error) {
	// git branch --all --contains <commit hash> (filtered by /remotes/<remote>/)
	remoteBranches := make([]*RemoteBranch, 0)
	command := script.LocalCommandFrom("git branch --all --contains")
	hash, err := c.GetFullHash()
	if err != nil {
		return remoteBranches, nil
	}
	command.Add(hash)

	pr, err := c.repository.Execute(command)
	if !pr.Successful() {
		err = fmt.Errorf("could not list remote branches containing commit %s", hash)
	}
	if err != nil {
		return remoteBranches, err
	}

	branchList, err := parseStarredBranchList(pr.Output())
	if err != nil {
		return nil, err
	}

	marker := "remotes/" + remote.GetName() + "/"
	for _, branch := range branchList {
		// filter
		if !strings.HasPrefix(branch, marker) {
			continue
		}
		branch = branch[len(marker):]
		if branch == "HEAD" {
			continue
		}

		remoteBranches = append(remoteBranches, newRemoteBranch(c.repository, remote, branch))
	}

	return remoteBranches, nil
}

func (c *Commit) EqualsPatchId(otherCommit *Commit) (bool, error) {
	thisPatchId, err := c.GetPatchId()
	if err != nil {
		return false, err
	}

	otherPatchId, err := otherCommit.GetPatchId()
	if err != nil {
		return false, err
	}

	return thisPatchId == otherPatchId, nil
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

func IsValidCommitHash(hash string) bool {
	r := regexp.MustCompile(`^[0-9a-f]{5,40}$`)
	return r.MatchString(hash)
}
