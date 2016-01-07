package nnmc

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"

	"golang.org/x/net/html/charset"
)

// NNMc values:
// client http.Client with cookie
type NNMc struct {
	client http.Client
}

// Init nnmc with login password
func Init(login string, password string) (*NNMc, error) {
	var client http.Client
	cookieJar, _ := cookiejar.New(nil)
	client.Jar = cookieJar
	urlPost := "http://nnm-club.me/forum/login.php"
	form := url.Values{}
	form.Set("username", login)
	form.Add("password", password)
	form.Add("redirect", "")
	form.Add("login", "âõîä")
	req, _ := http.NewRequest("POST", urlPost, bytes.NewBufferString(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))
	_, err := client.Do(req)
	return &NNMc{client: client}, err
}

// GetHTML get body from url
func (n *NNMc) GetHTML(url string) ([]byte, error) {
	resp, err := n.client.Get(url)
	if err != nil {
		log.Println("client Get error:", err)
		return nil, err
	}
	defer resp.Body.Close()
	utf8body, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		log.Println("Encoding error:", err)
		return nil, err
	}
	doc, err := ioutil.ReadAll(utf8body)
	if err != nil {
		log.Println("ioutil.ReadAll error:", err)
	}
	return doc, nil
}
