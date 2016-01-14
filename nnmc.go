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

// Topic from forum
type Topic struct {
	href    string
	text    string
	year    string
	quality string
}

// Film all values
// ID          id
// Name        Название
// EngName     Английское название
// Href        Ссылка
// Year        Год
// Genre       Жанр
// Country     Производство
// Director    Режиссер
// Producer    Продюсер
// Actors      Актеры
// Description Описание
// Age         Возраст
// ReleaseDate Дата мировой премьеры
// RussianDate Дата премьеры в России
// Duration    Продолжительность
// Quality     Качество видео
// Translation Перевод
// Subtitles   Вид субтитров
// Video       Видео
// Audio       Аудио
type Film struct {
	ID            int64
	Name          string
	EngName       string
	Href          string
	Year          int64
	Genre         string
	Country       string
	Director      string
	Producer      string
	Actors        string
	Description   string
	Age           string
	ReleaseDate   string
	RussianDate   string
	Duration      int64
	Quality       string
	Translation   string
	SubtitlesType string
	Subtitles     string
	Video         string
	Audio         string
	Kinopoisk     float64
	Imdb          float64
	NNM           float64
	Sound         string
	Size          int64
	DateCreate    string
	Torrent       string
	Poster        string
	Hide          bool
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

// getHTML get body from url
func getHTML(url string, n *NNMc) ([]byte, error) {
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
	doc = replaceAll(doc, "&nbsp;", " ")
	doc = replaceAll(doc, "&amp;", "&")
	return doc, nil
}

// ParseForumTree get topics from forumTree
func (n *NNMc) ParseForumTree(url string) ([]Topic, error) {
	var (
		topics []Topic
		reTree = regexp.MustCompile(`<a href="(viewtopic.php\?t=\d+)"class="topictitle">(.+?)\s\((\d{4})\)\s(.+?)</a>`)
	)
	body, err := getHTML(url, n)
	if err != nil {
		return topics, err
	}
	if reTree.Match(body) == false {
		return topics, fmt.Errorf("No topic in body")
	}
	findResult := reTree.FindAllSubmatch(body, -1)
	for _, v := range findResult {
		var t Topic
		t.href = string(v[1])
		t.text = string(v[2])
		t.year = string(v[3])
		t.quality = string(v[4])
		topics = append(topics, t)
	}

	return topics, nil
}

// ParseTopic get film from topic
func (n *NNMc) ParseTopic(url string) (Film, error) {
	var (
		film     Film
		reTopic  = regexp.MustCompile(`<span style="font-weight: bold">(Производство|Жанр|Режиссер|Продюсер|Актеры|Описание|Возраст|Дата мировой премьеры|Дата премьеры в России|Продолжительность|Качество видео|Перевод|Вид субтитров|Субтитры|Видео|Аудио):<\/span>(.+?)<`)
		reDate   = regexp.MustCompile(`> (\d{1,2} .{3} \d{4}).{9}<`)
		reSize   = regexp.MustCompile(`Размер блока: \d{1,2} MB"> (\d{1,2},\d{1,2}|\d{3,4})`)
		reRating = regexp.MustCompile(`>(\d,\d)<\/span`)
		reDl     = regexp.MustCompile(`<a href="download\.php\?id=(\d{5,7})" rel="nofollow">Скачать<`)
	)
	body, err := getHTML(url, n)
	film.Href = url
	if err != nil {
		return film, err
	}
	if reTopic.Match(body) == false {
		return film, fmt.Errorf("No topic in body")
	}
	findAttrs := reTopic.FindAllSubmatch(body, -1)
	for _, v := range findAttrs {
		one := strings.Trim(string(v[1]), " ")
		two := strings.Trim(string(v[2]), " ")
		switch one {
		case "Производство":
			film.Country = two
		case "Жанр":
			film.Genre = two
		case "Режиссер":
			film.Director = two
		case "Продюсер":
			film.Producer = two
		case "Актеры":
			film.Actors = two
		case "Описание":
			film.Description = two
		case "Возраст":
			film.Age = two
		case "Дата мировой премьеры":
			film.ReleaseDate = two
		case "Дата премьеры в России":
			film.RussianDate = two
		case "Продолжительность":
			if i64, err := strconv.ParseInt(two, 10, 64); err == nil {
				film.Duration = i64
			}
		case "Качество видео":
			film.Quality = two
		case "Перевод":
			film.Translation = two
		case "Вид субтитров":
			film.SubtitlesType = two
		case "Субтитры":
			film.Subtitles = two
		case "Видео":
			film.Video = two
		case "Аудио":
			film.Audio = two
		}
	}
	if reDate.Match(body) == true {
		film.DateCreate = string(reDate.FindSubmatch(body)[1])
	}
	if reSize.Match(body) == true {
		size := string(reSize.FindSubmatch(body)[1])
		size = strings.Replace(size, ",", ".", -1)
		if s64, err := strconv.ParseFloat(size, 64); err == nil {
			if s64 < 100 {
				s64 = s64 * 1000
			}
			film.Size = int64(s64)
		}
	}
	if reRating.Match(body) == true {
		rating := string(reRating.FindSubmatch(body)[1])
		rating = strings.Replace(rating, ",", ".", -1)
		if r64, err := strconv.ParseFloat(rating, 64); err == nil {
			film.NNM = r64
		}
	}
	if reDl.Match(body) == false {
		return film, fmt.Errorf("No torrent url in body")
	}
	findDl := reDl.FindAllSubmatch(body, -1)
	film.Torrent = "http://nnm-club.me/forum/download.php?id=" + string(findDl[0][1])
	return film, nil
}

func replaceAll(body []byte, from string, to string) []byte {
	var reStr = regexp.MustCompile(from)
	result := reStr.ReplaceAll(body, []byte(to))
	return result
}
