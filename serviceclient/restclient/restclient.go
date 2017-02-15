package restclient

import (
	"github.com/coffeehc/microserviceboot/base"
	"github.com/coffeehc/microserviceboot/loadbalancer"
	"context"
	"github.com/coffeehc/commons/httpcommons/client"
)


func newHttpClient(cxt context.Context, serviceInfo base.ServiceInfo, balancer loadbalancer.Balancer, defaultOption *client.HTTPClientOptions) client.HTTPClient {
	option := &client.HTTPClientOptions{
		Timeout:defaultOption.GetTimeout(),
		DialerTimeout:defaultOption.GetDialerTimeout(),
		DialerKeepAlive:defaultOption.GetDialerKeepAlive(),
		TransportTLSHandshakeTimeout:defaultOption.GetTransportTLSHandshakeTimeout(),
		TransportResponseHeaderTimeout:defaultOption.GetTransportResponseHeaderTimeout(),
		TransportIdleConnTimeout:defaultOption.GetTransportIdleConnTimeout(),
		TransportMaxIdleConns:defaultOption.GetTransportMaxIdleConns(),
		TransportMaxIdleConnsPerHost:defaultOption.GetTransportMaxIdleConnsPerHost(),
		Transport:defaultOption.Transport,
		GlobalHeaderSetting:defaultOption.GlobalHeaderSetting,
	}
	if option.Transport == nil {
		option.GetTransport().Dial = &_BalanceDialer{
			Timeout:option.GetTimeout(),
			KeepAlive:option.GetDialerKeepAlive(),
			balancer:balancer,
		}
	}
	return  client.NewHTTPClient(option)
}



