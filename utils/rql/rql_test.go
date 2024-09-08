package rql

import (
	"fmt"
	jsonql "github.com/NubeIO/jql"
	"testing"
)

type r struct {
	Result any
	JQ     *jsonql.JQL
}

func (r *r) Print(in ...any) {
	fmt.Printf("%v \n", in)
}

func (r *r) Add(a, b int) int {
	return a + b
}

func TestRule_publish(t *testing.T) {

	props := make(PropertiesMap)
	client := "RQL"

	newClient := &r{
		JQ: jsonql.New(),
	}

	props[client] = newClient
	jsonString := `{"id":"rubix-manager"}`
	props["jsonString"] = jsonString

	ruleEngine := New(nil)
	name := "rule"
	script := `
	var jsonString = '{"id":"rubix-manager"}';
	var q = "id='rubix-manager'";
	var result = jql(jsonString, q);
	if (result != null) {
		log("Result: " + JSON.stringify(result.Response));
		RQL.Result = JSON.stringify(result.Response)
	} else {
		log("Query failed");
	}
`

	err := ruleEngine.AddRule(name, script, props)
	if err != nil {
		fmt.Println(err)
		return
	}
	rule, err := ruleEngine.GetRule(name)
	execute, err := rule.Execute()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(execute)

}
