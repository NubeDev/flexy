package hostService

import (
	"fmt"
	model "github.com/NubeDev/flexy/app/models"
	"log"
)

type Fields struct {
	Name string `json:"name" form:"name" validate:"required,min=1,max=100" minLength:"1" maxLength:"100"`
	Ip   string `json:"ip" form:"ip" validate:"required,min=4,max=15" minLength:"4" maxLength:"15"`
}

type Host struct{}

var host *Host

func Get() *Host {
	return host
}

func Init() *Host {
	host = &Host{}
	return host
}

func (inst *Host) Create(body *Fields) (*model.Host, error) {
	return model.CreateHost(&model.Host{
		Name: body.Name,
		IP:   body.Ip,
	})
}

func (inst *Host) GetHosts() ([]*model.Host, error) {
	hosts := model.GetHosts()
	if hosts == nil {
		return nil, fmt.Errorf("no hosts found")
	}
	return hosts, nil
}

func (inst *Host) GetHost(uuid string) (*model.Host, error) {
	host, err := model.GetHost(uuid)
	if err != nil {
		log.Printf("Error retrieving host: %v", err)
		return nil, err
	}
	return host, nil
}

func (inst *Host) Update(uuid string, body *Fields) (*model.Host, error) {
	// Create a Host object to update the fields
	updatedHost := &model.Host{
		Name: body.Name,
		IP:   body.Ip,
	}
	// Perform the update
	host, err := model.UpdateHost(uuid, updatedHost)
	if err != nil {
		log.Printf("Error updating host with UUID %s: %v", uuid, err)
		return nil, err
	}
	return host, nil
}

func (inst *Host) Delete(uuid string) (*model.Message, error) {
	err := model.DeleteHost(uuid)
	if err != nil {
		log.Printf("Error deleting host with UUID %s: %v", uuid, err)
		return nil, err
	}
	return &model.Message{Message: "deleted ok"}, nil
}
