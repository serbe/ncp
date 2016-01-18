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
	Href    string
	Name    string
	Year    string
	Quality string
}

// Film all values
// ID            id
// Name          Название
// EngName       Английское название
// Href          Ссылка
// Year          Год
// Genre         Жанр
// Country       Производство
// Director      Режиссер
// Producer      Продюсер
// Actors        Актеры
// Description   Описание
// Age           Возраст
// ReleaseDate   Дата мировой премьеры
// RussianDate   Дата премьеры в России
// Duration      Продолжительность
// Quality       Качество видео
// Translation   Перевод
// SubtitlesType Вид субтитров
// Subtitles     Субтитры
// Video         Видео
// Audio         Аудио
// Kinopoisk     Рейтинг кинопоиска
// Imdb          Рейтинг IMDb
// NNM           Рейтинг nnm-club
// Sound         Звук
// Size          Размер
// DateCreate    Дата создания раздачи
// Torrent       Ссылка на torrent
// Poster        Ссылка на постер
// Hide          Скрывать в общем списке
type Film struct {
	ID            int64   `gorm:"column:id" sql:"AUTO_INCREMENT"`
	Name          string  `gorm:"column:name"`
	EngName       string  `gorm:"column:eng_name"`
	Href          string  `gorm:"column:href"`
	Year          int64   `gorm:"column:year"`
	Genre         string  `gorm:"column:genre"`
	Country       string  `gorm:"column:country"`
	Director      string  `gorm:"column:director"`
	Producer      string  `gorm:"column:producer"`
	Actors        string  `gorm:"column:actors"`
	Description   string  `gorm:"column:description"`
	Age           string  `gorm:"column:age"`
	ReleaseDate   string  `gorm:"column:release_date"`
	RussianDate   string  `gorm:"column:russian_date"`
	Duration      int64   `gorm:"column:duration"`
	Quality       string  `gorm:"column:quality"`
	Translation   string  `gorm:"column:translation"`
	SubtitlesType string  `gorm:"column:subtitles_type"`
	Subtitles     string  `gorm:"column:subtitles"`
	Video         string  `gorm:"column:video"`
	Audio         string  `gorm:"column:audio"`
	Kinopoisk     float64 `gorm:"column:kinopoisk"`
	IMDb          float64 `gorm:"column:imdb"`
	NNM           float64 `gorm:"column:nnm"`
	Sound         string  `gorm:"column:sound"`
	Size          int64   `gorm:"column:size"`
	DateCreate    string  `gorm:"column:date_create"`
	Torrent       string  `gorm:"column:torrent"`
	Poster        string  `gorm:"column:poster"`
	Hide          bool    `gorm:"column:hide" sql:"default:0"`
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
		t.Href = "http://nnm-club.me/forum/" + string(v[1])
		t.Name = string(v[2])
		t.Year = string(v[3])
		t.Quality = string(v[4])
		topics = append(topics, t)
	}

	return topics, nil
}

// ParseTopic get film from topic
func (n *NNMc) ParseTopic(topic Topic) (Film, error) {
	var (
		film     Film
		reTopic  = regexp.MustCompile(`<span style="font-weight: bold">(Производство|Жанр|Режиссер|Продюсер|Актеры|Описание|Возраст|Дата мировой премьеры|Дата премьеры в России|Дата Российской премьеры|Дата российской премьеры|Продолжительность|Качество видео|Качество|Перевод|Вид субтитров|Субтитры|Видео|Аудио):<\/span>(.+?)<`)
		reDate   = regexp.MustCompile(`> (\d{1,2} .{3} \d{4}).{9}<`)
		reSize   = regexp.MustCompile(`Размер блока: \d{1,2} MB"> (\d{1,2},\d{1,2}|\d{3,4}|\d{1,2})\s`)
		reRating = regexp.MustCompile(`>(\d,\d|\d)<\/span>.+?\(Голосов:`)
		reDl     = regexp.MustCompile(`<a href="download\.php\?id=(\d{5,7})" rel="nofollow">Скачать<`)
	)
	name := strings.Split(topic.Name, "/")
	switch len(name) {
	case 1:
		film.Name = strings.Trim(name[0], " ")
	case 2:
		film.Name = strings.Trim(name[0], " ")
		film.EngName = strings.Trim(name[1], " ")
	case 3:
		film.Name = strings.Trim(name[0], " ")
		film.EngName = strings.Trim(name[1], " ")
	}
	film.Href = topic.Href
	if year64, err := strconv.ParseInt(topic.Year, 10, 64); err == nil {
		film.Year = year64
	}
	body, err := getHTML(film.Href, n)
	if err != nil {
		return film, err
	}
	if reTopic.Match(body) == false {
		return film, fmt.Errorf("No topic in body")
	}
	findAttrs := reTopic.FindAllSubmatch(body, -1)
	for _, v := range findAttrs {
		one := strings.Trim(string(v[1]), " ")
		two := strings.Replace(string(v[2]), "<br />", "", -1)
		two = strings.Trim(two, " ")
		switch one {
		case "Производство":
			film.Country = two
		case "Жанр":
			film.Genre = strings.ToLower(two)
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
		case "Дата премьеры в России", "Дата российской премьеры", "Дата Российской премьеры":
			film.RussianDate = two
		case "Продолжительность":
			if i64, err := strconv.ParseInt(two, 10, 64); err == nil {
				film.Duration = i64
			}
		case "Качество видео", "Качество":
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
		film.DateCreate = replaceDate(string(reDate.FindSubmatch(body)[1]))
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

func replaceDate(s string) string {
	s = strings.Replace(s, " Янв ", ".01.", -1)
	s = strings.Replace(s, " Фев ", ".02.", -1)
	s = strings.Replace(s, " Мар ", ".03.", -1)
	s = strings.Replace(s, " Апр ", ".04.", -1)
	s = strings.Replace(s, " Май ", ".05.", -1)
	s = strings.Replace(s, " Июн ", ".06.", -1)
	s = strings.Replace(s, " Июл ", ".07.", -1)
	s = strings.Replace(s, " Авг ", ".08.", -1)
	s = strings.Replace(s, " Сен ", ".09.", -1)
	s = strings.Replace(s, " Окт ", ".10.", -1)
	s = strings.Replace(s, " Ноя ", ".11.", -1)
	s = strings.Replace(s, " Дек ", ".12.", -1)
	return s
}
