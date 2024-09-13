package guides

import "fmt"

type Method struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Args        []string `json:"args,omitempty"`      // Arguments if it's not a JSON body
	JSONBody    string   `json:"json_body,omitempty"` // JSON body schema/example
	UseJSON     bool     `json:"use_json"`            // Indicates if this method uses JSON body
}
type Module struct {
	Name    string   `json:"name"`
	Methods []Method `json:"methods"`
}

// HelpGuide to store all modules and provide methods to access data
type HelpGuide struct {
	Modules []Module `json:"modules"`
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

func (hg *HelpGuide) GetMethodArgs(methodName string) ([]string, error) {
	for _, module := range hg.Modules {
		for _, method := range module.Methods {
			if method.Name == methodName {
				return method.Args, nil
			}
		}
	}
	return nil, fmt.Errorf("method not found")
}
