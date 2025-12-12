package ip2region

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/sirupsen/logrus"
)

func GetIp2RegionImpl() Ip2Region {
	once.Do(func() {
		searcher, err := xdb.NewWithFileOnly(xdb.IPv4, "infrastructure/ip2region/ip2region.xdb")
		if err != nil {
			panic("ip2region.xdb 初始化失败 - " + err.Error())
		} else {
			ip2region = &Ip2RegionImpl{
				searcher: searcher,
			}
		}
	})
	return ip2region
}

func isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false // 无效的IP地址
	}

	// 检查私有IP地址范围
	if ip.IsLoopback() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() {
		return true // 环回地址或链路本地地址视为内网地址
	}

	// 使用IP网络类型来判断
	_, ipNet, _ := net.ParseCIDR("10.0.0.0/8")
	if ipNet.Contains(ip) {
		return true
	}
	_, ipNet, _ = net.ParseCIDR("172.16.0.0/12")
	if ipNet.Contains(ip) {
		return true
	}
	_, ipNet, _ = net.ParseCIDR("192.168.0.0/16")
	if ipNet.Contains(ip) {
		return true
	}
	_, ipNet, _ = net.ParseCIDR("169.254.0.0/16") // APIPA (Automatic Private IP Addressing) 地址范围
	if ipNet.Contains(ip) {
		return true
	}
	_, ipNet, _ = net.ParseCIDR("fc00::/7") // IPv6唯一本地地址范围（ULA）
	if ipNet.Contains(ip) {
		return true
	}

	return false // 不是内网地址
}

func parseSearchResult(searchResRaw string) (searchRes SearchRes) {
	splited := strings.Split(searchResRaw, "|")
	if len(splited) != 5 {
		logrus.Warn("IP 归属地解析失败，异常的 IP 地址")
		return
	}

	searchRes.Country = splited[0]
	searchRes.Country = splited[1]
	searchRes.Province = splited[2]
	searchRes.City = splited[3]
	searchRes.Company = splited[4]

	return
}

func makeShortRegion(searchRes SearchRes) string {
	if searchRes.Province != "0" && searchRes.City != "0" {
		return fmt.Sprintf("%s·%s", searchRes.Province, searchRes.City)
	}
	if searchRes.Country != "0" && searchRes.Province != "0" {
		return fmt.Sprintf("%s·%s", searchRes.Country, searchRes.Province)
	}
	if searchRes.Country != "0" {
		return searchRes.Country
	}
	return UnknownIpAddress
}

func (ip2region *Ip2RegionImpl) ParseIp(ip string) (region string, err error) {
	if ip2region.searcher == nil {
		return "", errors.New(UnknownIpSearcher)
	}

	if isPrivateIP(ip) {
		return "未知", nil
	}

	searchRes, err := ip2region.searcher.SearchByStr(ip)
	if err != nil {
		return "", errors.New(UnknownIpAddress)
	}

	region = makeShortRegion(parseSearchResult(searchRes))
	return region, nil
}
