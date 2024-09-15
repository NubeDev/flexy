package rqlservice

import (
	hostService "github.com/NubeDev/flexy/app/services/v1/host"
	"github.com/NubeDev/flexy/utils/rql"
)

var r rql.RQL

func RQL() rql.RQL {
	return r
}

func BootRQL() {
	r = rql.New(&rql.EmbeddedServices{
		Host: hostService.Get(),
	})
}
