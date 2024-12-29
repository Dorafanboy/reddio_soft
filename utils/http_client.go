package utils

import (
	"fmt"
	http "github.com/bogdanfinn/fhttp"
	tlsClient "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"reddio/extra"
	"strings"
	"time"
)

func CreateHttpClient(proxies string) (tlsClient.HttpClient, error) {
	idleConnTimeout := 90 * time.Second

	options := []tlsClient.HttpClientOption{
		tlsClient.WithClientProfile(profiles.Chrome_120),
		tlsClient.WithRandomTLSExtensionOrder(),
		tlsClient.WithInsecureSkipVerify(),
		tlsClient.WithTimeoutSeconds(30),
		tlsClient.WithTransportOptions(&tlsClient.TransportOptions{
			DisableKeepAlives:      false,
			DisableCompression:     false,
			MaxIdleConns:           100,              // Максимальное количество простаивающих соединений
			MaxIdleConnsPerHost:    10,               // Максимальное количество простаивающих соединений на хост
			MaxConnsPerHost:        20,               // Максимальное количество соединений на хост
			MaxResponseHeaderBytes: 1 << 20,          // Максимальный размер заголовка ответа (1 MB)
			WriteBufferSize:        16 * 1024,        // Размер буфера записи (16 KB)
			ReadBufferSize:         16 * 1024,        // Размер буфера чтения (16 KB)
			IdleConnTimeout:        &idleConnTimeout, // Время ожидания простаивающего соединения (90 секунд)
			RootCAs:                nil,
		}),
	}
	if proxies != "" {
		options = append(options, tlsClient.WithProxyUrl(fmt.Sprintf("http://%s", proxies)))
	}

	client, err := tlsClient.NewHttpClient(tlsClient.NewNoopLogger(), options...)
	if err != nil {

		extra.Logger{}.Error("Failed to create Http Client: %s", err)
		return nil, err
	}

	return client, nil
}

func CookiesToHeader(allCookies map[string][]*http.Cookie) string {
	var cookieStrs []string
	for _, cookies := range allCookies {
		for _, cookie := range cookies {
			cookieStrs = append(cookieStrs, cookie.Name+"="+cookie.Value)
		}
	}
	return strings.Join(cookieStrs, "; ")
}
