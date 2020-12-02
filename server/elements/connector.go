package elements

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"strings"
)

type ForcepointConnector struct {
	EntryPoints map[string]string
}

func (c *ForcepointConnector) GetEntryPoints() {
	endPoint := fmt.Sprintf("http://%s:%s/api/v1/Entypoints", viper.GetString("CONNECTOR.HOSTNAME"),
		viper.GetString("CONNECTOR.PORT"))
	resp, err := http.Get(endPoint)
	if err != nil {
		LogFaTal(err, endPoint).Fatal("Failed in loading the entry points from the connector")
	}
	if resp == nil {
		LogFaTal(err, endPoint).Fatal("No response received from the connector")
	}
	if resp.StatusCode != http.StatusOK {
		LogFaTal(err, endPoint).Fatalf("received unexpected http status code: %d\n", resp.StatusCode)
	}
	// read the response body and convert it to a map
	result, err := ResponseToMap(resp.Body)
	if err != nil {
		LogFaTal(err, endPoint).Fatal("Failed in converting response body to a map object")
	}
	c.EntryPoints = result
	LogInfo(endPoint).Info("EntryPoints loaded")

}

func (c *ForcepointConnector) TokenPermission(t *TokenRequest) (bool, error) {
	//result := c.GetEntryPoints()
	//c.EntryPoints = result
	requestBody, err := json.Marshal(t)
	if err != nil {
		return false, err
	}
	bodyData := bytes.NewBuffer(requestBody)
	resp, err := http.Post(c.EntryPoints["TokenPermission"], "application/json", bodyData)
	if err != nil {
		return false, err
	}
	var responseMap map[string]string
	err = json.NewDecoder(resp.Body).Decode(&responseMap)
	if err != nil {
		return false, err
	}
	if responseMap["allow"] == "false" {
		return false, errors.New(responseMap["reason"])
	}
	return true, nil
}

func (c *ForcepointConnector) CreateUser(user *UserObj) (*UserObj, int, error) {
	var userUrl map[string]string
	userInfo := struct {
		LoginName   interface{} `json:"login_name"`
		DisplayName interface{} `json:"display_name"`
		Active      interface{} `json:"active"`
	}{
		LoginName:   user.UserName,
		DisplayName: user.DisplayName,
		Active:      user.Active,
	}
	requestBody, err := json.Marshal(userInfo)
	if err != nil {
	}
	bodyData := bytes.NewBuffer(requestBody)
	resp, err := http.Post(c.EntryPoints["CreateUser"], "application/json", bodyData)
	if err != nil {
		return user, resp.StatusCode, err
	}
	buff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	if err := json.Unmarshal(buff, &userUrl); err != nil {
		return nil, resp.StatusCode, err
	}
	if resp.StatusCode != http.StatusCreated {
		return user, resp.StatusCode, errors.New(fmt.Sprintf("unexpected http status code: %d is received", resp.StatusCode))
	}
	user.Schemas = user.Schemas[1:]
	parts := strings.Split(userUrl["userUrl"], "/")
	urlString := parts[len(parts)-1]
	user.Id = urlString
	return user, resp.StatusCode, nil
}

func (c *ForcepointConnector) UpdateUser(user *UserObj, updateOp UpdateOperation) (*UserObj, error) {
	updateInfo := struct {
		UserId     string      `json:"user_id"`
		Operations []Operation `json:"operations"`
	}{
		UserId:     updateOp.Id,
		Operations: updateOp.Operations,
	}
	requestBody, err := json.Marshal(updateInfo)
	if err != nil {
	}
	bodyData := bytes.NewBuffer(requestBody)
	response, err := http.Post(c.EntryPoints["UpdateUser"], "application/json", bodyData)
	if err != nil {
		return user, err
	}
	if response.StatusCode != http.StatusOK {
		return user, errors.New(fmt.Sprintf("unexpected http status code: %d", response.StatusCode))
	}
	user.UserName = updateInfo.UserId
	users, err := c.GetUsers(user)
	if err != nil {
		return user, errors.New("failed in requesting the updated user")
	}
	/*
		{"schemas":["urn:ietf:params:scim:api:messages:2.0:PatchOp"],"Operations":[{"op":"Add","path":"displayName","value":"dlo.bagari@corkbizdev.onmicrosoft.com"},{"op":"Add","path":"name.givenName","value":"Dlo"},{"op":"Add","path":"name.familyName","value":"Bagari"},{"op":"Add","path":"externalId","value":"b98e93d8-181a-41b3-9746-b472e574d0ed"}]}

	*/
	responseUser := &users[0]
	for _, op := range updateOp.Operations {
		if op.Path == "displayName" {
			responseUser.DisplayName = op.Value
		}
		if op.Path == "name.givenName" {
			responseUser.Names.GivenName = op.Value
		}
		if op.Path == "name.familyName" {
			responseUser.Names.FamilyName = op.Value
		}
		if op.Path == "externalId" {
			responseUser.ExternalId = op.Value
		}
	}
	return responseUser, nil
}

func (c *ForcepointConnector) GetUsers(user *UserObj) ([]UserObj, error) {
	var allUsers []UserObj
	var recUsers []map[string]interface{}
	endPoint := c.EntryPoints["GetUsers"]
	if user.Id != "" {
		endPoint = fmt.Sprintf("%s?id=%s", endPoint, user.Id)
	}
	resp, err := http.Get(endPoint)
	if err != nil {
		return []UserObj{*user}, err
	}
	if resp == nil {
		return []UserObj{*user}, errors.New("no response received from connector")
	}
	if resp.StatusCode != http.StatusOK {
		return []UserObj{*user}, errors.New(fmt.Sprintf("got unexpected http status code: %d",
			resp.StatusCode))
	}
	if err := json.NewDecoder(resp.Body).Decode(&recUsers); err != nil {
		return []UserObj{*user}, err
	}
	if len(recUsers) == 1 {
		newUser := NewUser()
		newUser.Id = recUsers[0]["id"]
		newUser.Active = recUsers[0]["active"]
		allUsers = append(allUsers, newUser)
	} else {
		for _, u := range recUsers {
			newUser := NewUser()
			newUser.Id = u["id"]
			newUser.Active = u["active"]
			allUsers = append(allUsers, newUser)
		}
	}
	return allUsers, nil
}

func (c *ForcepointConnector) DeleteUser(user *UserObj) error {
	client := &http.Client{}
	endPoint := c.EntryPoints["DeleteUser"]
	if strings.HasSuffix(endPoint, "{id}") {
		endPoint = strings.Replace(endPoint, "{id}", "", 1)
	}
	deleteUrl := fmt.Sprintf("%s%s", endPoint, user.UserName)
	req, err := http.NewRequest("DELETE", deleteUrl, nil)
	if err != nil {
		return err
	}
	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return errors.New(fmt.Sprintf("got unexpected http status code for deleting a user: %d", resp.StatusCode))
	}

	return nil
}

func (c *ForcepointConnector) CreateGroup(group *Group) (*Group, error) {
	groupInfo := struct {
		DisplayName string `json:"display_name"`
		ExternalId  string `json:"external_id"`
	}{
		DisplayName: group.DisplayName,
		ExternalId:  group.ExternalId,
	}
	//TODO: send the request
	fmt.Println(groupInfo)
	return group, nil
}

func (c *ForcepointConnector) UpdateGroup(group *Group, updateOp UpdateOperation) (*Group, error) {
	UpdateInfo := struct {
		GroupId    string
		Operations []Operation
	}{
		GroupId:    updateOp.Id,
		Operations: updateOp.Operations,
	}
	//TODO: send request to the end point
	fmt.Println(UpdateInfo)
	return group, nil
}

func (c *ForcepointConnector) GetGroup(group *Group) ([]*Group, error) {
	if group.DisplayName != "" {
		//TODO: send request for one group
	} else {
		//TODO: send request to get all exists groups
	}
	return []*Group{group}, nil
}

func (c *ForcepointConnector) DeleteGroup(user *Group) error {
	//TODO: send request to delete a group
	return nil
}
