package elements

type UserObj struct {
	Schemas     []string    `json:"schemas"`
	Id          interface{} `json:"id,omitempty"`
	DisplayName interface{} `json:"displayName,omitempty"`
	ExternalId  interface{} `json:"externalId,omitempty"`
	Meta        Meta        `json:"meta"`
	UserName    interface{} `json:"userName,omitempty"`
	Names       Name        `json:"name,omitempty"`
	Active      interface{} `json:"active,omitempty"`
}

func NewUser() UserObj {
	return UserObj{
		Schemas: []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
		Id:      "",
		Meta:    Meta{ResourceType: "User"},
		Active:  false,
	}
}

type Name struct {
	FamilyName interface{} `json:"familyName,omitempty"`
	GivenName  interface{} `json:"givenName,omitempty"`
}

type Meta struct {
	ResourceType string `json:"resourceType"`
}

type ScimError struct {
	Schemas []string `json:"schemas"`
	Status  int      `json:"status"`
	Detail  string   `json:"detail"`
}

func GenerateScimError(status int, details string) ScimError {
	return ScimError{
		Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
		Status:  status,
		Detail:  details,
	}
}

func (u *UserObj) PostUser(connector Connector) (*UserObj, int, error) {
	return connector.CreateUser(u)
}

func (u *UserObj) GetUsers(connector Connector) ([]UserObj, error) {
	return connector.GetUsers(u)
}

func (u *UserObj) UpdateUser(connector Connector, updateOp UpdateOperation) (*UserObj, error) {
	return connector.UpdateUser(u, updateOp)
}

func (u *UserObj) DeleteUser(connector Connector) error {
	return connector.DeleteUser(u)
}
