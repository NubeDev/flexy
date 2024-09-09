package execute

import (
	"fmt"
	"github.com/go-cmd/cmd"
	"strings"
	"time"
)

type Execute interface {
	AddTimeout(timeout int) Execute
	Run(name string, args ...string) *Response
}

type execute struct {
	cmd     *cmd.Cmd
	timeout time.Duration
}

func New() Execute {
	return &execute{}
}

func (e *execute) AddTimeout(timeout int) Execute {
	e.timeout = time.Duration(timeout) * time.Second
	return e
}

func (e *execute) Run(name string, args ...string) *Response {
	e.cmd = cmd.NewCmd(name, args...)
	if name == "" {
		return &Response{
			Error: "command name can not be empty, try something like; pwd, uptime",
		}
	}

	statusChan := e.cmd.Start() // non-blocking

	// create a timeout
	if e.timeout > 0 {
		timeout := time.After(e.timeout)
		select {
		case <-statusChan:
			// command finished
		case <-timeout:
			// command timed out
			e.cmd.Stop() // optional: try to stop the process
			return &Response{
				status: cmd.Status{Error: fmt.Errorf("command timed out")},
			}
		}
	} else {
		<-statusChan // wait for the command to finish
	}

	// retrieve the status using the function
	status := e.cmd.Status()
	r := &Response{}
	r.status = status
	r.Response = r.AsString()
	if r.status.Error != nil {
		r.Error = fmt.Sprintf("%v", r.AsError())
	}
	return r
}

type Response struct {
	status   cmd.Status
	Response string `json:"response"`
	Error    string `json:"error"`
}

func (r *Response) AsString() string {
	return strings.Join(r.status.Stdout, "\n")
}

func (r *Response) GetErrors() []string {
	return r.status.Stderr
}

func (r *Response) AsError() error {
	if r.status.Error != nil {
		return r.status.Error
	}
	return nil
}
