package rql

import (
	"errors"
	"fmt"
	hostService "github.com/NubeDev/flexy/app/services/v1/host"
	systeminfo "github.com/NubeDev/flexy/utils/system"
	"github.com/NubeDev/flexy/utils/systemctl"
	jsonql "github.com/NubeIO/jql"
	"github.com/dop251/goja"
	"strings"
	"sync"
)

// RQL defines an interface for rule management.
type RQL interface {
	AddRule(name, script string, props PropertiesMap) error
	GetRules() map[string]*Rule
	GetRule(name string) (*Rule, error)
	DeleteRule(name string) error
	RunAndDestroy(name, script string, props PropertiesMap) (goja.Value, error)
}

type EmbeddedServices struct {
	Host    *hostService.Host
	SystemD *systemctl.CTL
}

// New creates a new RuleEngine
func New(services *EmbeddedServices) RQL {
	re := &RuleEngine{
		rules:    make(map[string]*Rule),
		services: services,
		system:   systeminfo.NewSystem(),
		systemD:  services.SystemD,
	}
	return re
}

// PropertiesMap is a map of string to interface{}.
type PropertiesMap map[string]interface{}

type Rules interface {
	Execute() (goja.Value, error)
	initValue(key string, initialValue any) goja.Value
	updateValue(key string, newValue any) goja.Value
	getValue(key string) goja.Value
	log(call goja.FunctionCall) goja.Value
}

type Rule struct {
	script           string
	mu               sync.Mutex
	vm               *goja.Runtime
	subscribedTopics map[string]bool
	storage          map[string]any
}

func newRule(script string, vm *goja.Runtime, props PropertiesMap) *Rule {
	r := &Rule{
		script:           script,
		vm:               vm,
		subscribedTopics: make(map[string]bool),
		storage:          make(map[string]any),
	}
	for k, v := range props {
		err := vm.Set(k, v)
		if err != nil {
			return nil
		}
	}
	return r
}

func (inst *RuleEngine) DeleteRule(name string) error {
	if _, exists := inst.rules[name]; !exists {
		return errors.New("rule does not exist")
	}
	delete(inst.rules, name)
	return nil
}

func (inst *Rule) Execute() (goja.Value, error) {
	var err error
	defer inst.mu.Unlock()
	inst.mu.Lock()
	res, err := inst.vm.RunString(inst.script)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (inst *RuleEngine) RunAndDestroy(name, script string, props PropertiesMap) (goja.Value, error) {
	err := inst.AddRule(name, script, props)
	if err != nil {
		return nil, err
	}
	rule, err := inst.GetRule(name)
	if err != nil {
		return nil, err
	}
	result, err := rule.Execute()
	if err != nil {
		return nil, err
	}
	err = inst.DeleteRule(name)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (inst *Rule) initValue(key string, initialValue any) goja.Value {
	if _, exists := inst.storage[key]; !exists {
		inst.storage[key] = initialValue
	}
	return inst.vm.ToValue(nil)
}

func (inst *Rule) updateValue(key string, newValue any) goja.Value {
	inst.storage[key] = newValue
	return inst.vm.ToValue(nil)
}

func (inst *Rule) getValue(key string) goja.Value {
	if value, exists := inst.storage[key]; exists {
		return inst.vm.ToValue(value)
	}
	return inst.vm.ToValue(nil)
}

func (inst *Rule) log(call goja.FunctionCall) goja.Value {
	var logParts []string
	for _, arg := range call.Arguments {
		logParts = append(logParts, arg.String())
	}
	logMessage := strings.Join(logParts, " ")
	fmt.Println(logMessage)
	return goja.Undefined()
}

// RuleEngine struct
type RuleEngine struct {
	RQL
	rules    map[string]*Rule
	JQ       *jsonql.JQL // Add JQ property
	mu       sync.Mutex
	services *EmbeddedServices
	system   systeminfo.System
	systemD  *systemctl.CTL
}

func (inst *RuleEngine) AddRule(name, script string, props PropertiesMap) error {
	if _, exists := inst.rules[name]; exists {
		return errors.New("rule already exists")
	}
	vm := goja.New()
	r := newRule(script, vm, props)
	inst.rules[name] = r

	var err error
	err = vm.Set("log", r.log)
	if err != nil {
		return err
	}
	err = vm.Set("init", r.initValue)
	if err != nil {
		return err
	}
	err = vm.Set("set", r.updateValue)
	if err != nil {
		return err
	}
	err = vm.Set("get", r.getValue)
	if err != nil {
		return err
	}
	err = vm.Set("props", props)
	if err != nil {
		return err
	}
	err = vm.Set("hosts", inst.services.Host)
	if err != nil {
		return err
	}
	err = vm.Set("system", inst.system)
	if err != nil {
		return err
	}
	err = vm.Set("ctl", inst.systemD)
	if err != nil {
		return err
	}
	return nil
}

func (inst *RuleEngine) GetRules() map[string]*Rule {
	return inst.rules
}

func (inst *RuleEngine) GetRule(name string) (*Rule, error) {
	rule, ok := inst.rules[name]
	if !ok {
		return nil, errors.New(fmt.Sprintf("rule:%s does not exist", name))
	}
	return rule, nil
}
