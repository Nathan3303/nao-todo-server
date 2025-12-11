package ip2region

import (
	"sync"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

type Ip2Region interface {
	ParseIp(ip string) (region string, err error)
}

type Ip2RegionImpl struct {
	searcher *xdb.Searcher
}

var (
	ip2region *Ip2RegionImpl
	once      sync.Once
)

type SearchRes struct {
	Country  string
	Country2 string
	Province string
	City     string
	Company  string
}

const (
	UnknownIpAddress  = "未知 IP"
	UnknownIpSearcher = "IpSearcher 未初始化"
	PrivateIpAddress  = "不是一个有效的公网 IP"
)
