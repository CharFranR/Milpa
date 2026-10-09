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

type elasticProductTransport struct {
	base http.RoundTripper
}

func (t *elasticProductTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if resp.Header.Get("X-Elastic-Product") == "" {
		resp.Header.Set("X-Elastic-Product", "Elasticsearch")
	}
	return resp, nil
}

func CreateESClient(clientConf dto.ESClient) (*elasticsearch.Client, error) {

	addresses := make([]string, 0, 2)
	for _, address := range []string{clientConf.Endpoint1, clientConf.Endpoint2} {
		if address != "" {
			addresses = append(addresses, address)
		}
	}

	cfg := elasticsearch.Config{
		Addresses: addresses,
		Username:  clientConf.Username,
		Password:  clientConf.Password,
		Transport: &elasticProductTransport{
			base: &http.Transport{
				MaxIdleConnsPerHost:   clientConf.MaxIdleConnsPerHost,
				ResponseHeaderTimeout: 10 * time.Second,
				DialContext:           (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
			},
		},
	}

	client, err := elasticsearch.NewClient(cfg)

	if err != nil {
		return nil, fmt.Errorf("CreateESClient error : %w", err)
	}

	return client, nil
}
