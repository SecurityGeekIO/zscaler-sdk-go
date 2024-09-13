package lssconfigcontroller

import (
	"fmt"
	"net/http"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v2/zpa/services"
)

type LSSFormats struct {
	Csv  string `json:"csv"`
	Tsv  string `json:"tsv"`
	Json string `json:"json"`
}

func GetFormats(service *services.Service, logType string) (*LSSFormats, *http.Response, error) {
	v := new(LSSFormats)
	relativeURL := fmt.Sprintf("%slssConfig/logType/formats", mgmtConfigTypesAndFormats)
	resp, err := service.NewRequestDo("GET", relativeURL, struct {
		LogType string `url:"logType"`
	}{
		LogType: logType,
	}, nil, &v)
	if err != nil {
		return nil, nil, err
	}

	return v, resp, nil
}
