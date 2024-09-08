package hostService

import (
	model "github.com/NubeDev/flexy/app/models"
)

type Fields struct {
	Name string `json:"name" form:"name" validate:"required,min=1,max=100" minLength:"1" maxLength:"100"`
	Ip   string `json:"ip" form:"ip" validate:"required,min=4,max=15" minLength:"4" maxLength:"15"`
}

func Create(body *Fields) (*model.Host, error) {
	return model.CreateHost(&model.Host{
		Name: body.Name,
		IP:   body.Ip,
	})
}
