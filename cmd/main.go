package main

import (
	"crypto/tls"
	"fmt"
	"naotodoserver/conf"
	"naotodoserver/infrastructure"
	"naotodoserver/interfaces/routers"
	"net/http"
	"runtime"
	"time"
)

func main() {
	// @step 0. 加载配置文件
	conf.InitConfig()

	// @step 1. 限制 Go 进程资源
	runtime.GOMAXPROCS(conf.Conf.Server.GoMaxProc)

	// @step 2. 初始化日志系统
	infrastructure.LoadLogger()

	// @step 3. 加载 Infrastructure 层
	infrastructure.LoadDBs()
	svc := infrastructure.LoadDomains()
	infrastructure.WireSSE()
	infrastructure.LoadCron(svc)

	// @step 3. 加载路由
	router := routers.InitRouters(svc)

	addr := fmt.Sprintf("%s:%s", conf.Conf.Server.Ip, conf.Conf.Server.Port)

	// @step 4. 根据是否配置 TLS 证书，选择 HTTPS 或 HTTP 启动
	if conf.Conf.Server.CertFile != "" && conf.Conf.Server.KeyFile != "" {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			},
		}
		srv := &http.Server{
			Addr:         addr,
			Handler:      router,
			TLSConfig:    tlsConfig,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
		if err := srv.ListenAndServeTLS(
			conf.Conf.Server.CertFile,
			conf.Conf.Server.KeyFile,
		); err != nil {
			panic("HTTPS 服务器启动失败 - " + err.Error())
		}
	} else {
		// 开发环境：纯 HTTP
		if err := router.Run(addr); err != nil {
			panic("HTTP 服务器启动失败 - " + err.Error())
		}
	}
}
