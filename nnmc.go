package nnmc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html/charset"
)

// NNMc values:
// client http.Client with cookie
type NNMc struct {
	client http.Client
}

type topic struct {
	href string
	text string
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

//func parseString(text string, array []string) (string, string) {
//	result := ""
//	for _, value := range array {
//		index := strings.Index(strings.ToLower(strings.Replace(text, "ё", "е", -1)), strings.ToLower(value))
//		if index != -1 {
//			if result == "" {
//				result = value
//			} else {
//				result += ", " + value
//			}
//			text = text[0:index] + text[index+len(value):]
//		}
//	}
//	return text, result
//}

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

// ParseForumTree get topics from forumTree
func ParseForumTree(body []byte) ([]topic, error) {
//	var reTopic = regexp.MustCompile(`<a href="(viewtopic.php\?t=\d+)"class="topictitle">(.+?)\s\((\d{4})\)\s(.+?)\s\[(.+?)\]</a>`)
	var reTopic = regexp.MustCompile(`<a href="(viewtopic.php\?t=\d+)"class="topictitle">(.+?\(\d{4}\).+?)</a>`)
	var topics []topic
	if reTopic.Match(body) == false {
		return topics, fmt.Errorf("No topic in body")
	}
	findResult := reTopic.FindAllSubmatch(body, -1)

	for v := range findResult {
		var t topic
		t.href = string(findResult[v][1])
		t.text = string(findResult[v][2])
		topics = append(topics, t)
	}
	return topics, nil
}
