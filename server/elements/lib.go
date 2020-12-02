package elements

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strconv"
)

func GetCountAndStartIndex(args url.Values) (int, int) {
	var count int
	var startIndex int
	var err error
	if countInQuery, ok := args["count"]; ok {
		count, err = strconv.Atoi(countInQuery[0])
		if err != nil {
			err = errors.Wrap(err, "Failed in reading count arg from request.args")
			log.Fatal(err)
		}
	} else {
		count = 20
	}
	if startIndexInQuery, ok := args["startIndex"]; ok {
		startIndex, err = strconv.Atoi(startIndexInQuery[0])
		if err != nil {
			err = errors.Wrap(err, "Failed in reading startIndex arg from request.args")
			log.Fatal(err)
		}
	} else {
		startIndex = 1
	}
	return count, startIndex

}

func ScimSourceUsers(u []UserObj, startIndex int, count int) ToScimResourceUsers {
	return ToScimResourceUsers{
		Schemas:      []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
		Resources:    u,
		StartIndex:   startIndex,
		TotalResult:  len(u),
		ItemsPerPage: count,
	}
}

func ScimSourceGroups(g []*Group, startIndex int, count int) ToScimResourceGroups {
	return ToScimResourceGroups{
		Schemas:      []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
		Resources:    g,
		StartIndex:   startIndex,
		TotalResult:  len(g),
		ItemsPerPage: count,
	}
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func ResponseToMap(body io.Reader) (map[string]string, error) {
	buff, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	var objMap map[string]string
	if err := json.Unmarshal(buff, &objMap); err != nil {
		return nil, err
	}
	return objMap, nil
}
