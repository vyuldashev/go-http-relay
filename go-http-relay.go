package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

type RequestError struct {
	Error string `json:"error"`
}

func NewErrorResponse(err error) []byte {
	re := RequestError{err.Error()}
	e, _ := json.Marshal(re)

	return e
}

type App struct {
	httpClient *http.Client
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	// prepare request
	req, err := http.NewRequest(
		r.Method,
		viper.Get("target_url").(string)+r.RequestURI,
		io.Reader(r.Body),
	)

	if err != nil {
		w.Write(NewErrorResponse(err))
		return
	}

	req.Header.Add("Content-type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := a.httpClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		w.Write(NewErrorResponse(err))
		return
	}

	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		w.Write(NewErrorResponse(err))
		return
	}

	_, err = w.Write(b)
}

func setDefaultConfig() {
	viper.SetDefault("proxy_url", "")
	viper.SetDefault("proxy_username", "")
	viper.SetDefault("proxy_password", "")
}

func loadConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/go-http-relay/")
	viper.AddConfigPath(".")

	setDefaultConfig()

	err := viper.ReadInConfig()
	checkErr(err)
}

func proxyUrl() *url.URL {

	proxyRawUrl := viper.Get("proxy_url").(string)

	// add optional auth for proxy
	if len(viper.Get("proxy_username").(string)) > 0 && len(viper.Get("proxy_password").(string)) > 0 {
		sl := strings.Split(proxyRawUrl, "://")
		sl[1] = fmt.Sprintf("%s:%s@%s",
			viper.Get("proxy_username").(string),
			viper.Get("proxy_password").(string),
			sl[1],
		)

		proxyRawUrl = strings.Join(sl, "://")
	}

	u, err := url.Parse(proxyRawUrl)
	checkErr(err)

	return u
}

func main() {
	loadConfig()

	tr := &http.Transport{
		Proxy: http.ProxyURL(proxyUrl()),
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}

	client := &http.Client{Transport: tr}

	a := App{client}

	r := mux.NewRouter()

	// handle any request
	r.PathPrefix("/").Handler(&a)

	err := http.ListenAndServe(":"+viper.Get("app_port").(string), r)
	checkErr(err)
}
