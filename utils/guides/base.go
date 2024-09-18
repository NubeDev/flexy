package guides

import "fmt"

type Args struct {
	Name string `json:"name"`
	Type string `json:"type"` // string, int, float, bool
}

type Method struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Topic       string `json:"string"`
	Args        []Args `json:"args,omitempty"`      // Arguments with type information
	JSONBody    string `json:"json_body,omitempty"` // JSON body schema/example
	UseJSON     bool   `json:"use_json,omitempty"`  // Indicates if this method uses JSON body
}

// NewMethod creates a new Method instance
func NewMethod(name, description, subject string, useJSON bool, jsonBody string, args []Args) Method {
	return Method{
		Name:        name,
		Description: description,
		Topic:       subject,
		Args:        args,
		JSONBody:    jsonBody,
		UseJSON:     useJSON,
	}
}

type Module struct {
	Name    string   `json:"name"`
	Methods []Method `json:"methods"`
}

// NewModule creates a new Module instance
func NewModule(name string, methods []Method) Module {
	return Module{
		Name:    name,
		Methods: methods,
	}
}

// HelpGuide to store all modules and provide methods to access data
type HelpGuide struct {
	Modules []Module `json:"modules"`
}

// NewHelpGuide creates a new HelpGuide instance
func NewHelpGuide(modules []Module) *HelpGuide {
	return &HelpGuide{
		Modules: modules,
	}
}

func (hg *HelpGuide) GetMethods() []string {
	var methods []string
	for _, module := range hg.Modules {
		for _, method := range module.Methods {
			methods = append(methods, method.Name)
		}
	}
	return methods
}

func (hg *HelpGuide) GetMethodDetails(methodName string) (map[string]interface{}, error) {
	for _, module := range hg.Modules {
		for _, method := range module.Methods {
			if method.Name == methodName {
				details := make(map[string]interface{})
				details["description"] = method.Description

				if method.UseJSON {
					details["json_body"] = method.JSONBody
				} else {
					details["args"] = method.Args
				}
				return details, nil
			}
		}
	}
	return nil, fmt.Errorf("method not found")
}

func (hg *HelpGuide) GetMethodDescription(methodName string) (string, error) {
	for _, module := range hg.Modules {
		for _, method := range module.Methods {
			if method.Name == methodName {
				return method.Description, nil
			}
		}
	}
	return "", fmt.Errorf("method not found")
}

func (hg *HelpGuide) GetMethodArgs(methodName string) ([]Args, error) {
	for _, module := range hg.Modules {
		for _, method := range module.Methods {
			if method.Name == methodName {
				return method.Args, nil
			}
		}
	}
	return nil, fmt.Errorf("method not found")
}

// NewArgString creates a new argument of type string
func NewArgString(name string) Args {
	return Args{
		Name: name,
		Type: "string",
	}
}

// NewArgBool creates a new argument of type bool
func NewArgBool(name string) Args {
	return Args{
		Name: name,
		Type: "bool",
	}
}

// NewArgInt creates a new argument of type int
func NewArgInt(name string) Args {
	return Args{
		Name: name,
		Type: "int",
	}
}

// NewArgFloat creates a new argument of type float
func NewArgFloat(name string) Args {
	return Args{
		Name: name,
		Type: "float",
	}
}
