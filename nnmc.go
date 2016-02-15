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
	"time"

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
// UpdatedAt     Дата обновления записи БД
// CreatedAt     Дата создания записи БД
type Film struct {
	ID            int64     `gorm:"column:id"             db:"id"             sql:"AUTO_INCREMENT"`
	Name          string    `gorm:"column:name"           db:"name"           sql:"type:text"`
	EngName       string    `gorm:"column:eng_name"       db:"eng_name"       sql:"type:text"`
	Href          string    `gorm:"column:href"           db:"href"           sql:"type:text"`
	Year          int64     `gorm:"column:year"           db:"year"`
	Genre         string    `gorm:"column:genre"          db:"genre"          sql:"type:text"`
	Country       string    `gorm:"column:country"        db:"country"        sql:"type:text"`
	Director      string    `gorm:"column:director"       db:"director"       sql:"type:text"`
	Producer      string    `gorm:"column:producer"       db:"producer"       sql:"type:text"`
	Actors        string    `gorm:"column:actors"         db:"actors"         sql:"type:text"`
	Description   string    `gorm:"column:description"    db:"description"    sql:"type:text"`
	Age           string    `gorm:"column:age"            db:"age"            sql:"type:text"`
	ReleaseDate   string    `gorm:"column:release_date"   db:"release_date"   sql:"type:text"`
	RussianDate   string    `gorm:"column:russian_date"   db:"russian_date"   sql:"type:text"`
	Duration      int64     `gorm:"column:duration"       db:"duration"`
	Quality       string    `gorm:"column:quality"        db:"quality"        sql:"type:text"`
	Translation   string    `gorm:"column:translation"    db:"translation"    sql:"type:text"`
	SubtitlesType string    `gorm:"column:subtitles_type" db:"subtitles_type" sql:"type:text"`
	Subtitles     string    `gorm:"column:subtitles"      db:"subtitles"      sql:"type:text"`
	Video         string    `gorm:"column:video"          db:"video"          sql:"type:text"`
	Audio         string    `gorm:"column:audio"          db:"audio"          sql:"type:text"`
	Kinopoisk     float64   `gorm:"column:kinopoisk"      db:"kinopoisk"`
	IMDb          float64   `gorm:"column:imdb"           db:"imdb"`
	NNM           float64   `gorm:"column:nnm"            db:"nnm"`
	Sound         string    `gorm:"column:sound"          db:"sound"          sql:"type:text"`
	Size          int64     `gorm:"column:size"           db:"size"`
	DateCreate    string    `gorm:"column:date_create"    db:"date_create"`
	Torrent       string    `gorm:"column:torrent"        db:"torrent"`
	Poster        string    `gorm:"column:poster"         db:"poster"`
	UpdatedAt     time.Time `gorm:"column:updated_at"     db:"updated_at"`
	CreatedAt     time.Time `gorm:"column:created_at"     db:"created_at"`
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
	time.Sleep(2000 * time.Millisecond)
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
		reTopic  = regexp.MustCompile(`<span style="font-weight: bold">(Производство|Жанр|Режиссер|Продюсер|Актеры|Описание фильма|Описание|Возраст|Дата мировой премьеры|Дата премьеры в России|Дата Российской премьеры|Дата российской премьеры|Продолжительность|Качество видео|Качество|Перевод|Вид субтитров|Субтитры|Видео|Аудио):\s*<\/span>(.+?)<`)
		reDate   = regexp.MustCompile(`> (\d{1,2} .{3} \d{4}).{9}<`)
		reSize   = regexp.MustCompile(`Размер блока: \d.+?B"> (\d{1,2},\d{1,2}|\d{3,4}|\d{1,2})\s`)
		reRating = regexp.MustCompile(`>(\d,\d|\d)<\/span>.+?\(Голосов:`)
		reDl     = regexp.MustCompile(`<a href="download\.php\?id=(\d{5,7})" rel="nofollow">Скачать<`)
		reImg    = regexp.MustCompile(`"postImg postImgAligned img-right" title="http:\/\/assets\.nnm-club\.ws\/forum\/image\.php\?link=(.+?jpe{0,1}g)`)
	)
	name := strings.Split(topic.Name, " / ")
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
	var reFn = regexp.MustCompile(`(\d{6})`)
	filename := string(reFn.Find([]byte(film.Href)))
	_ = ioutil.WriteFile(filename+".html", body, 0644)
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
		case "Описание фильма", "Описание":
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
			if caseInsensitiveContains(two, "не требуется") == false {
				film.Translation = two
			} else {
				film.Translation = "Не требуется"
			}
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
	if reDl.Match(body) == false {
		return film, fmt.Errorf("No torrent url in body")
	}
	findDl := reDl.FindAllSubmatch(body, -1)
	film.Torrent = "http://nnm-club.me/forum/download.php?id=" + string(findDl[0][1])
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
	if reImg.Match(body) == true {
		film.Poster = string(reImg.FindSubmatch(body)[1])
	}
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

func caseInsensitiveContains(s, substr string) bool {
	s, substr = strings.ToUpper(s), strings.ToUpper(substr)
	return strings.Contains(s, substr)
}
