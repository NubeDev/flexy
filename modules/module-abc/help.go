package main

import "github.com/NubeDev/flexy/utils/guides"

var helpGuide guides.HelpGuide

func help() {
	methods := []guides.Method{
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

	module := guides.Module{
		Name:    "cli-tool",
		Methods: methods,
	}
	helpGuide = guides.HelpGuide{
		Modules: []guides.Module{module},
	}

}
