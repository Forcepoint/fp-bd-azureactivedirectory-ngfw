package elements

type Configs struct {
	Issuer           string `json:"issuer"`
	ProductAccessKey string `json:"product_access_key"`
	ProductEndPoint  string `json:"product_end_point"`
}
type TokenRequest struct {
	ProductName string `json:"productName"`
	UserName    string `json:"username"`
	Password    string `json:"password"`
}

type ToScimResourceUsers struct {
	Schemas      []string  `json:"schemas"`
	Resources    []UserObj `json:"Resources"`
	StartIndex   int       `json:"startIndex"`
	TotalResult  int       `json:"totalResults"`
	ItemsPerPage int       `json:"itemsPerPage"`
}

type ToScimResourceGroups struct {
	Schemas      []string `json:"schemas"`
	Resources    []*Group `json:"Resources"`
	StartIndex   int      `json:"startIndex"`
	TotalResult  int      `json:"totalResults"`
	ItemsPerPage int      `json:"itemsPerPage"`
}

type UpdateOperation struct {
	Schemas    []string    `json:"schemas"`
	Id         string      `json:"id,omitempty"`
	Operations []Operation `json:"Operations"`
}

type Operation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}
