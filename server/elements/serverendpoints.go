package elements

import (
	"github.com/gorilla/mux"
	"net/http"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

var Routes = []Route{
	{
		Name:        "getUsers",
		Method:      "GET",
		Pattern:     "/scim/v2/Users",
		HandlerFunc: LogToScreen(ValidateJsonWebToken(GetUsers)),
	},
	{
		Name:        "getUsersByID",
		Method:      "GET",
		Pattern:     "/scim/v2/Users/{id}",
		HandlerFunc: LogToScreen(ValidateJsonWebToken(GetUserById)),
	},
	{
		Name:        "AddUsers",
		Method:      "POST",
		Pattern:     "/scim/v2/Users",
		HandlerFunc: LogToScreen(ValidateJsonWebToken(PostUser)),
	},
	{
		Name:        "UpdateUsers",
		Method:      "PATCH",
		Pattern:     "/scim/v2/Users/{id}",
		HandlerFunc: LogToScreen(ValidateJsonWebToken(UpdateUser)),
	},
	{
		Name:        "root",
		Method:      "GET",
		Pattern:     "/scim/v2",
		HandlerFunc: Root,
	},
	{
		Name:        "DeleteUser",
		Method:      "DELETE",
		Pattern:     "/scim/v2/Users/{id}",
		HandlerFunc: LogToScreen(ValidateJsonWebToken(DeleteUser)),
	},

	{
		Name:        "GetGroup",
		Method:      "GET",
		Pattern:     "/scim/v2/Groups/{id}",
		HandlerFunc: LogToScreen(ValidateJsonWebToken(GetGroupById)),
	},
	{
		Name:        "GetGroups",
		Method:      "GET",
		Pattern:     "/scim/v2/Groups",
		HandlerFunc: LogToScreen(ValidateJsonWebToken(GetGroups)),
	},
	{
		Name:        "CreateGroup",
		Method:      "POST",
		Pattern:     "/scim/v2/Groups",
		HandlerFunc: LogToScreen(ValidateJsonWebToken(CreateGroup)),
	},
	{
		Name:        "UpdateGroup",
		Method:      "PATCH",
		Pattern:     "/scim/v2/Groups/{id}",
		HandlerFunc: LogToScreen(ValidateJsonWebToken(UpdateGroup)),
	},
	{
		Name:        "DeleteGroup",
		Method:      "Delete",
		Pattern:     "/scim/v2/Groups/{id}",
		HandlerFunc: LogToScreen(ValidateJsonWebToken(DeleteGroup)),
	},
	{
		Name:        "GetToken",
		Method:      "POST",
		Pattern:     "/scim/v2/token",
		HandlerFunc: LogToScreen(GetToken),
	},
}

func AddRoutes(router *mux.Router) *mux.Router {
	for _, route := range Routes {
		router.Methods(route.Method).Path(route.Pattern).Handler(route.HandlerFunc)
	}
	return router
}
