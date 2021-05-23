package git

import (
	"fmt"
	"strings"

	"github.com/jojomi/go-script/v2"
)

type LocalBranch struct {
	name               string
	trackingRemoteName string

	repository *Repository
}

func newLocalBranch(repository *Repository, name string) *LocalBranch {
	b := LocalBranch{
		repository: repository,
		name:       name,
	}
	return &b
}

func (b *LocalBranch) GetName() string {
	return b.name
}

func (b *LocalBranch) GetFullName() string {
	return b.name
}

func (b *LocalBranch) GetHeadCommit() (*Commit, error) {
	command := script.LocalCommandFrom("git rev-parse")
	command.Add(b.GetName())

	pr, err := b.repository.Execute(command)
	if !pr.Successful() {
		err = fmt.Errorf("getting HEAD commit faile for LocalBranch %s", b.GetName())
	}
	if err != nil {
		return &Commit{}, err
	}

	hash := strings.TrimSpace(pr.Output())

	commit := newCommit(b.repository, hash)
	return commit, nil
}

func (b *LocalBranch) IsMainBranch() (bool, error) {
	main, err := b.repository.GetMainBranch()
	if err != nil {
		return false, err
	}
	return main.Equals(b), nil
}

func (b *LocalBranch) IsMergedTo(target Branch) (bool, error) {
	localBranchHead, err := b.GetHeadCommit()
	if err != nil {
		return false, err
	}
	targetHead, err := target.GetHeadCommit()
	if err != nil {
		return false, err
	}

	// both have the same HEAD commit? -> merged by definition
	if localBranchHead.Equals(targetHead) {
		return true, nil
	}

	// https://stackoverflow.com/a/40011122
	// git merge-base <commit-hash-step1> <commit-hash-step2>
	command := script.LocalCommandFrom("git merge-base")
	command.AddAll(localBranchHead.GetHash(), targetHead.GetHash())

	pr, err := b.repository.Execute(command)
	if !pr.Successful() {
		err = fmt.Errorf("could not find merge-base between %s and %s", localBranchHead.GetHash(), targetHead.GetHash())
	}
	if err != nil {
		return false, err
	}
	mergeBase := strings.TrimSpace(pr.Output())
	return mergeBase == localBranchHead.GetHash(), nil
}

func (b *LocalBranch) GetTrackingRemote(r *Remote) (*RemoteBranch, error) {
	if b.trackingRemoteName != "" {
		return newRemoteBranch(b.repository, r, b.trackingRemoteName), nil
	}

	// https://serverfault.com/a/384862
	// git rev-parse --symbolic-full-name master@{u}
	command := script.LocalCommandFrom("git rev-parse --symbolic-full-name")
	command.Add(b.GetName() + "@{u}")

	pr, err := b.repository.Execute(command)
	if !pr.Successful() {
		err = fmt.Errorf("could not find tracking LocalBranch for %s", b.GetName())
	}
	if err != nil {
		return nil, err
	}

	trackingName := strings.Replace(strings.TrimSpace(pr.Output()), "refs/remotes/", "", 1)

	if trackingName != "" {
		b.trackingRemoteName = trackingName
	}

	return newRemoteBranch(b.repository, r, trackingName), nil
}

func (b *LocalBranch) Delete() error {
	// git LocalBranch -d <local_LocalBranch>
	command := script.LocalCommandFrom("git LocalBranch --delete")
	command.Add(b.GetName())

	pr, err := b.repository.Execute(command)
	if !pr.Successful() {
		err = fmt.Errorf("could not delete LocalBranch %s", b.GetName())
	}
	if err != nil {
		return err
	}

	return nil
}

func (b *LocalBranch) Equals(otherLocalBranch *LocalBranch) bool {
	return b.GetName() == otherLocalBranch.GetName()
}

func (b *LocalBranch) String() string {
	return "Local branch " + b.GetName()
}
