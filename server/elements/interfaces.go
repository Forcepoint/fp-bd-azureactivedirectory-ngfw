package elements

type Connector interface {
	CreateUser(user *UserObj) (*UserObj, int, error)
	UpdateUser(user *UserObj, updateOp UpdateOperation) (*UserObj, error)
	GetUsers(user *UserObj) ([]UserObj, error)
	DeleteUser(user *UserObj) error
	CreateGroup(user *Group) (*Group, error)
	UpdateGroup(user *Group, updateOp UpdateOperation) (*Group, error)
	GetGroup(user *Group) ([]*Group, error)
	DeleteGroup(user *Group) error
	TokenPermission(t *TokenRequest) (bool, error)
	GetEntryPoints()
}
