package git

import (
	"fmt"
	"strings"

	"github.com/jojomi/go-script/v2"
)

type RemoteBranch struct {
	name string

	repository *Repository
	remote     *Remote
}

func newRemoteBranch(repository *Repository, remote *Remote, name string) *RemoteBranch {
	b := RemoteBranch{
		repository: repository,
		remote:     remote,
		name:       name,
	}
	return &b
}

func (b *RemoteBranch) GetName() string {
	return b.name
}

func (b *RemoteBranch) GetFullName() string {
	return b.remote.GetName() + "/" + b.name
}

func (b *RemoteBranch) Delete() error {
	// git push <remote> --delete <branch name>
	command := script.LocalCommandFrom("git push")
	command.AddAll(b.remote.GetName(), "--delete", b.GetName())

	pr, err := b.repository.Execute(command)
	if !pr.Successful() {
		err = fmt.Errorf("could not delete remote branch %s on %s", b.GetName(), b.remote)
	}
	return err
}

func (b *RemoteBranch) IsMergedTo(target *RemoteBranch) (bool, error) {
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

func (r *RemoteBranch) GetHeadCommit() (*Commit, error) {
	command := script.LocalCommandFrom("git rev-parse")
	command.Add(r.GetFullName())

	pr, err := r.repository.Execute(command)
	if !pr.Successful() {
		err = fmt.Errorf("getting HEAD commit faile for RemoteBranch %s", r.GetName())
	}
	if err != nil {
		return &Commit{}, err
	}

	hash := strings.TrimSpace(pr.Output())

	return newCommit(r.repository, hash)
}

func (b *RemoteBranch) Equals(otherRemoteBranch *RemoteBranch) bool {
	return b.GetFullName() == otherRemoteBranch.GetFullName()
}

func (b *RemoteBranch) String() string {
	return fmt.Sprintf("Remote branch %s @ %s", b.GetName(), b.remote.GetName())
}
