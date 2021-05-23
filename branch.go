package git

type Branch interface {
	GetName() string
	GetFullName() string // including remote if applicable
	GetHeadCommit() (*Commit, error)
	IsMainBranch() (bool, error)
	IsMergedTo(target Branch) (bool, error)
	Delete() error
	String() string
}
