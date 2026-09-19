package elasticSsearch

import (
	"crypto/tls"
	"fmt"
	"milpa/aplication/dto"
	"net"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

func CreateESClient(clientConf dto.ESClient) (*elasticsearch.Client, error) {

	cfg := elasticsearch.Config{
		Addresses: []string{
			clientConf.Endpoint1,
			clientConf.Endpoint2,
		},
		Username: clientConf.Username,
		Password: clientConf.Password,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   clientConf.MaxIdleConnsPerHost,
			ResponseHeaderTimeout: time.Second,
			DialContext:           (&net.Dialer{Timeout: time.Second}).DialContext,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	client, err := elasticsearch.NewClient(cfg)

	if err != nil {
		return nil, fmt.Errorf("CreateESClient error : %w", err)
	}

	return client, nil
}
