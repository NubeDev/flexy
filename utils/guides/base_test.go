package guides

import (
	"fmt"
	"testing"
)

func TestHelpGuide_GetMethodArgs(t *testing.T) {
	methods := []Method{
		{
			Name:        "createUser",
			Description: "Create a new user in the system",
			UseJSON:     true,
			JSONBody: `{
            "name": "string",
            "email": "string",
            "password": "string"
        }`,
		},
		{
			Name:        "getUser",
			Description: "Retrieve user details by ID",
			UseJSON:     false,
			Args:        []string{"userID"},
		},
	}
	module := Module{
		Name:    "cli-tool",
		Methods: methods,
	}
	h := HelpGuide{
		Modules: []Module{module},
	}

	for i, s := range h.GetMethods() {
		fmt.Println(i, s)
	}
	methodName := "createUser"
	fmt.Println(h.GetMethodArgs("getUser"))
	fmt.Println(h.GetMethodDetails(methodName))

}
