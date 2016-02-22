package ncp

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
)

// NCp values:
// client http.Client with cookie
type NCp struct {
	client http.Client
}

// Topic from forum
type Topic struct {
	Href    string
	Name    string
	Year    string
	Quality string
	Body    []byte
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
// Resolution    Разрешение видео
// Audio         Аудио
// Kinopoisk     Рейтинг кинопоиска
// Imdb          Рейтинг IMDb
// NNM           Рейтинг nnm-club
// Sound         Звук
// Size          Размер
// DateCreate    Дата создания раздачи
// Torrent       Ссылка на torrent
// Poster        Ссылка на постер
// Seeders       Количество раздающих
// Leechers      Количество скачивающих
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
	Resolution    string    `gorm:"column:resolution"     db:"resolution"     sql:"type:text"`
	Audio         string    `gorm:"column:audio"          db:"audio"          sql:"type:text"`
	Kinopoisk     float64   `gorm:"column:kinopoisk"      db:"kinopoisk"`
	IMDb          float64   `gorm:"column:imdb"           db:"imdb"`
	NNM           float64   `gorm:"column:nnm"            db:"nnm"`
	Sound         string    `gorm:"column:sound"          db:"sound"          sql:"type:text"`
	Size          int64     `gorm:"column:size"           db:"size"`
	DateCreate    string    `gorm:"column:date_create"    db:"date_create"    sql:"type:text"`
	Torrent       string    `gorm:"column:torrent"        db:"torrent"        sql:"type:text"`
	Poster        string    `gorm:"column:poster"         db:"poster"         sql:"type:text"`
	Seeders       int64     `gorm:"column:seeders"        db:"seeders"`
	Leechers      int64     `gorm:"column:leechers"       db:"leechers"`
	UpdatedAt     time.Time `gorm:"column:updated_at"     db:"updated_at"`
	CreatedAt     time.Time `gorm:"column:created_at"     db:"created_at"`
}

// Init nnmc with login password
func Init(login string, password string) (*NCp, error) {
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
	return &NCp{client: client}, err
}

// getHTML get body from url
func getHTML(url string, n *NCp) ([]byte, error) {
	resp, err := n.client.Get(url)
	if err != nil {
		log.Println("client Get error:", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("ioutil.ReadAll error:", err)
	}
	buffer := bytes.NewBufferString("")
	for _, char := range body {
		var ch = Utf(char)
		fmt.Fprintf(buffer, "%c", ch)
	}
	doc := buffer.Bytes()
	doc = replaceAll(doc, "&nbsp;", " ")
	doc = replaceAll(doc, "&amp;", "&")
	return doc, nil
}

// ParseForumTree get topics from forumTree
func (n *NCp) ParseForumTree(url string) ([]Topic, error) {
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
func (n *NCp) ParseTopic(topic Topic) (Film, error) {
	var (
		film Film
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
	topic.Body = body
	film.Country = topic.getCountry()
	film.Genre = topic.getGenre()
	film.Director = topic.getDirector()
	film.Producer = topic.getProducer()
	film.Actors = topic.getActors()
	film.Description = topic.getDescription()
	film.Age = topic.getAge()
	film.ReleaseDate = topic.getReleaseDate()
	film.RussianDate = topic.getRussianDate()
	film.Duration = topic.getDuration()
	film.Quality = topic.getQuality()
	film.Translation = topic.getTranslation()
	film.SubtitlesType = topic.getSubtitlesType()
	film.Subtitles = topic.getSubtitles()
	film.Video = topic.getVideo()
	film.Resolution = getResolution(film.Video)
	film.Audio = topic.getAudio()
	film.Torrent = topic.getTorrent()
	film.DateCreate = topic.getDate()
	film.Size = topic.getSize()
	film.NNM = topic.getRating()
	film.Poster = topic.getPoster()
	film.Seeders = topic.getSeeds()
	film.Leechers = topic.getLeechs()
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

func cleanStr(str string) string {
	var reSpan = regexp.MustCompile("<span .*?>")
	str = strings.Replace(str, "<br />", "", -1)
	str = reSpan.ReplaceAllString(str, "")
	str = strings.Trim(str, " ")
	return str
}
