package elements

type Group struct {
	Schemas     []string `json:"schemas"`
	ID          string   `json:"id,omitempty"`
	ExternalId  string   `json:"externalId"`
	DisplayName string   `json:"displayName"`
	Meta        Meta     `json:"meta"`
	Member      []string `json:"member,omitempty"`
}

func NewGroup() *Group {
	return &Group{
		Schemas:     []string{"urn:ietf:params:scim:schemas:core:2.0:Group"},
		ID:          "",
		ExternalId:  "",
		DisplayName: "",
		Meta:        Meta{},
		Member:      nil,
	}
}

func (g *Group) CreateGroup(connector Connector) (*Group, error) {
	return connector.CreateGroup(g)
}

func (g *Group) GetGroups(connector Connector) ([]*Group, error) {
	return connector.GetGroup(g)
}

func (g *Group) UpdateGroup(connector Connector, updateOp UpdateOperation) (*Group, error) {
	return connector.UpdateGroup(g, updateOp)
}

func (g *Group) DeleteGroup(connector Connector) error {
	return connector.DeleteGroup(g)
}
