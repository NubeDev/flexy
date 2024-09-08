package startup

import hostService "github.com/NubeDev/flexy/app/services/v1/host"

func InitServices() {
	hostService.Init()
}
