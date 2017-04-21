package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type gameServerItemT struct {
	Size    uint32 `json:"size"`
	Count   uint32 `json:"count"`
	Timeout uint32 `json:"timeout"`
}

type configT struct {
	IPRouterList      []string                   `json:"ip_router_list"`
	ResolveDNSList    []string                   `json:"resolve_dns_list"`
	CheckSpeedURLList []string                   `json:"check_speed_url_list"`
	GameServerList    map[string]gameServerItemT `json:"game_server_list"`
}

/////////////////////////////////////////////////////

type DNSReportT struct {
	IP    string `json:"ip"`
	Time  string `json:"time"`
	Error string `json:"error"`
}

type URLSpeedResultItemT struct {
	Size  uint32 `json:"size"`
	Time  uint32 `json:"time"`
	Speed uint32 `json:"speed"`
	Error string `json:"error"`
}

type GameServerSpeedResultItemT struct {
	RecvCount uint32 `json:"recv_count"`
	MinTime   uint32 `json:"min_time"`
	MaxTime   uint32 `json:"max_time"`
	AveTime   uint32 `json:"ave_time"`
	Error     string `json:"error"`
}

type reportT struct {
	PublicIP          string                                `json:"public_ip"`
	SystemInfo        string                                `json:"systeminfo"`
	UserDNSServer     string                                `json:"user_dns_server"`
	Gs108DNS          DNSReportT                            `json:"gs108_dns"`
	IPRouterList      map[string]string                     `json:"ip_router_list"`
	ResolveDNSList    map[string]DNSReportT                 `json:"resolve_dns_list"`
	CheckSpeedURLList map[string]URLSpeedResultItemT        `json:"check_speed_url_list"`
	GameServerList    map[string]GameServerSpeedResultItemT `json:"game_server_list"`
}

func getConfig(url string, c *configT) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	d, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(d, c)
	if err != nil {
		return err
	}

	return nil
}
