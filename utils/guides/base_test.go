package guides

import (
	"github.com/NubeDev/flexy/utils/helpers/pprint"
	"testing"
)

func TestHelpGuide_GetMethodArgs(t *testing.T) {
	arg1 := NewArgFloat("num1")
	arg2 := NewArgFloat("num2")

	method := NewMethod("mathAdd", "Adds two numbers", "<appID>.post.math.add.run", false, "", []Args{arg1, arg2})

	module := NewModule("MathOperations", []Method{method})

	guide := NewHelpGuide([]Module{module})

	pprint.PrintJSON(guide)

}
