package activation

import (
	"errors"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zidentity"
)

const (
	activationStatusEndpoint = "/status"
	activationEndpoint       = "/status/activate"
)

type Activation struct {
	Status string `json:"status"`
}

func GetActivationStatus(service *zidentity.Service) (*Activation, error) {
	var activation Activation
	err := service.Client.Read(activationStatusEndpoint, &activation)
	if err != nil {
		return nil, err
	}

	return &activation, nil
}

func CreateActivation(service *zidentity.Service, activation Activation) (*Activation, error) {
	resp, err := service.Client.Create(activationEndpoint, activation)
	if err != nil {
		return nil, err
	}

	createdActivation, ok := resp.(*Activation)
	if !ok {
		return nil, errors.New("object returned from api was not an activation pointer")
	}

	return createdActivation, nil
}
