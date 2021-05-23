package git

import (
	"fmt"

	"github.com/jojomi/go-script/v2"
)

type Remote struct {
	repository *Repository
	name       string
}

func newRemote(repository *Repository, name string) *Remote {
	remote := Remote{
		repository: repository,
		name:       name,
	}
	return &remote
}

func (r *Remote) GetName() string {
	return r.name
}

func (r *Remote) HasBranch(name string) (bool, error) {
	// https://git-scm.com/docs/git-show-ref7
	// git show-ref --verify --quiet refs/remotes/<remote-name>/<remote-branch-name>
	command := script.LocalCommandFrom("git show-ref --verify --quiet")
	command.Add("refs/remotes/" + r.GetName() + "/" + name)

	pr, err := r.repository.Execute(command)
	if err != nil {
		return false, err
	}

	return pr.Successful(), nil
}

func (r *Remote) GetBranch(name string) (*RemoteBranch, error) {
	existing, err := r.HasBranch(name)
	if err != nil {
		return nil, err
	}
	if !existing {
		return nil, fmt.Errorf("could not find RemoteBranch %s", name)
	}
	return newRemoteBranch(r.repository, r, name), nil
}

func (r *Remote) GetBranches() ([]*RemoteBranch, error) {
	// git ls-remote --heads <remote>
	command := script.LocalCommandFrom("git ls-remote --heads")
	command.Add(r.GetName())

	pr, err := r.repository.Execute(command)
	if !pr.Successful() {
		err = fmt.Errorf("could not list remote branches on %s", r.GetName())
	}
	if err != nil {
		return []*RemoteBranch{}, err
	}

	branches := make([]*RemoteBranch, 0, 10)

	branchList, err := parseBranchList(pr.Output())
	if err != nil {
		return nil, err
	}

	for _, branch := range branchList {
		branches = append(branches, newRemoteBranch(r.repository, r, branch))
	}

	return branches, err
}

func (r *Remote) String() string {
	return "Remote " + r.GetName()
}
