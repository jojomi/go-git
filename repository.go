package git

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver"
	script "github.com/jojomi/go-script/v2"
)

type Repository struct {
	path string

	// cached values
	mainLocalBranchName string
}

func OpenRepository(path string) (*Repository, error) {
	r := Repository{
		path: path,
	}
	return &r, nil
}

func (r *Repository) GetPath() string {
	return r.path
}

func (r *Repository) GetRemote(name string) (*Remote, error) {
	existing, err := r.HasRemote(name)
	if err != nil {
		return nil, err
	}
	if !existing {
		return nil, fmt.Errorf("could not find remote %s", name)
	}
	return newRemote(r, name), nil
}

func (r *Repository) GetRemotes() ([]*Remote, error) {
	command := script.LocalCommandFrom("git remote show")
	pr, err := r.Execute(command)
	if !pr.Successful() {
		err = fmt.Errorf("could not list remotes")
	}
	if err != nil {
		return []*Remote{}, err
	}

	remotes := make([]*Remote, 0, 10)
	for _, line := range strings.Split(strings.TrimSpace(pr.Output()), "\n") {
		remotes = append(remotes, newRemote(r, line))
	}

	return remotes, err
}

func (r *Repository) HasRemote(name string) (bool, error) {
	command := script.LocalCommandFrom("git remote show")
	command.Add(name)

	pr, err := r.Execute(command)
	if err != nil {
		return false, err
	}

	return pr.Successful(), nil
}

func (r *Repository) GetBranch(name string) (*LocalBranch, error) {
	existing, err := r.HasBranch(name)
	if err != nil {
		return nil, err
	}
	if !existing {
		return nil, fmt.Errorf("could not find local branch %s", name)
	}
	return newLocalBranch(r, name), nil
}

func (r *Repository) GetBranches() ([]*LocalBranch, error) {
	// https://stackoverflow.com/a/66190381
	// git show-ref --heads
	command := script.LocalCommandFrom("git show-ref --heads")
	pr, err := r.Execute(command)
	if !pr.Successful() {
		err = fmt.Errorf("could not list local LocalBranches")
	}
	if err != nil {
		return []*LocalBranch{}, err
	}

	localBranches := make([]*LocalBranch, 0, 10)

	branchList, err := parseBranchList(pr.Output())
	if err != nil {
		return nil, err
	}

	for _, branch := range branchList {
		localBranches = append(localBranches, newLocalBranch(r, branch))
	}

	return localBranches, err
}

func (r *Repository) String() string {
	return "Repository at " + r.GetPath()
}

func (r *Repository) HasBranch(name string) (bool, error) {
	// https://stackoverflow.com/q/5167957
	// git show-ref --verify --quiet refs/heads/<LocalBranch-name>
	command := script.LocalCommandFrom("git show-ref --verify --quiet")
	command.Add("refs/heads/" + name)

	pr, err := r.Execute(command)
	if err != nil {
		return false, err
	}

	return pr.Successful(), nil
}

func (r *Repository) GetMainBranch() (*LocalBranch, error) {
	// cache
	if r.mainLocalBranchName != "" {
		return newLocalBranch(r, r.mainLocalBranchName), nil
	}

	candidates := make([]string, 0, 3)

	// check config
	command := script.LocalCommandFrom("git config init.defaultLocalBranch")
	pr, err := r.Execute(command)
	if err != nil {
		return nil, err
	}
	configLocalBranch := strings.TrimSpace(pr.Output())
	if configLocalBranch != "" {
		candidates = append(candidates, configLocalBranch)
	}

	candidates = append(candidates, "master", "main", "primary")

	var existing bool
	for _, candidate := range candidates {
		existing, err = r.HasBranch(candidate)
		if err != nil {
			return nil, err
		}
		if !existing {
			continue
		}

		// put to cache
		r.mainLocalBranchName = candidate

		return newLocalBranch(r, candidate), nil
	}

	return nil, fmt.Errorf("no main branch found")
}

func (r *Repository) GetCurrentBranch() (*LocalBranch, error) {
	// https://stackoverflow.com/a/6245587

	v := MustGetGitVersion()
	min, err := semver.NewConstraint(">= 2.22")
	if err != nil {
		return &LocalBranch{}, err
	}

	var command *script.LocalCommand

	// git LocalBranch --show-current
	command = script.LocalCommandFrom("git branch --show-current")

	if !min.Check(v) {
		// Fallback
		// git rev-parse --abbrev-ref HEAD
		command = script.LocalCommandFrom("git rev-parse --abbrev-ref HEAD")
	}

	// common parsing due to same output structure
	pr, err := r.Execute(command)
	if !pr.Successful() {
		err = fmt.Errorf("failed getting current LocalBranch")
	}
	if err != nil {
		return &LocalBranch{}, err
	}
	return newLocalBranch(r, strings.TrimSpace(pr.Output())), nil
}

func (r *Repository) GetCurrentCommit() (Commit, error) {
	return Commit{}, nil
}

var sc = script.NewContext()

func (r *Repository) Execute(c script.Command) (pr *script.ProcessResult, err error) {
	workingDir := r.GetPath()
	if workingDir == "" {
		return nil, fmt.Errorf("repository path not set")
	}
	sc.SetWorkingDir(workingDir)

	return sc.ExecuteSilent(c)
}
