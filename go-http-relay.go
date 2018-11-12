package main

import (
	"crypto/tls"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

type App struct {
	httpClient *http.Client
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest(
		r.Method,
		viper.Get("target_url").(string)+r.RequestURI,
		io.Reader(r.Body),
	)
	checkErr(err)

	req.Header.Add("Content-type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := a.httpClient.Do(req)

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	checkErr(err)

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(b)
	checkErr(err)
}

func loadConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/go-http-relay/")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	checkErr(err)
}

func main() {
	loadConfig()

	proxyRawUrl := fmt.Sprintf("socks5://%s:%s@%s",
		viper.Get("proxy_username").(string),
		viper.Get("proxy_password").(string),
		viper.Get("proxy_url").(string),
	)

	u, err := url.Parse(proxyRawUrl)
	checkErr(err)

	tr := &http.Transport{
		Proxy: http.ProxyURL(u),
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}

	client := &http.Client{Transport: tr}

	a := App{client}

	r := mux.NewRouter()

	// handle any request
	r.PathPrefix("/").Handler(&a)

	err = http.ListenAndServe(":"+viper.Get("app_port").(string), r)
	checkErr(err)
}
