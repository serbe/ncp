package ncp

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// NC values:
// client http.Client with cookie
type NC struct {
	client      http.Client
	baseAddress string
	debugNet    bool
}

// Topic from forum
type Topic struct {
	Href     string
	Name     string
	Year     string
	Quality  string
	Seeders  int
	Leechers int
	Body     []byte
}

// Film all values
// Section       Раздел форума
// Name          Название
// EngName       Английское название
// Href          Ссылка
// Year          Год
// Genre         Жанр
// Country       Производство
// RawCountry    Производство (оригинальная строка)
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
// Audio1        Аудио1
// Audio2        Аудио2
// Audio3        Аудио3
// Kinopoisk     Рейтинг кинопоиска
// Imdb          Рейтинг IMDb
// NNM           Рейтинг nnm-club
// Sound         Звук
// Size          Размер
// DateCreate    Дата создания раздачи
// Torrent       Ссылка на torrent
// Magnet        Ссылка на magnet
// Poster        Ссылка на постер
// Seeders       Количество раздающих
// Leechers      Количество скачивающих
type Film struct {
	Section       string
	Name          string
	EngName       string
	Href          string
	Year          int
	Genre         []string
	Country       []string
	RawCountry    string
	Director      []string
	Producer      []string
	Actor         []string
	Description   string
	Age           string
	ReleaseDate   string
	RussianDate   string
	Duration      string
	Quality       string
	Translation   string
	SubtitlesType string
	Subtitles     string
	Video         string
	Resolution    string
	Audio1        string
	Audio2        string
	Audio3        string
	Kinopoisk     float64
	IMDb          float64
	NNM           float64
	Size          int
	DateCreate    string
	Torrent       string
	Magnet        string
	Poster        string
	Seeders       int
	Leechers      int
}

// Init nnmc with login password
func Init(login, pass, address, proxyURL string, debug bool) (*NC, error) {
	n := new(NC)
	n.debugNet = debug
	n.baseAddress = address
	n.client = http.Client{
		Timeout: time.Duration(10) * time.Second,
	}
	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err == nil {
			n.client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxy),
				// DisableKeepAlives: true,
			}
		}
	}
	cookieJar, _ := cookiejar.New(nil)
	n.client.Jar = cookieJar

	var cookie = new([]*http.Cookie)
	err := load(cookie)
	if err == nil {
		var body []byte
		u, _ := url.Parse(address + "/forum/")
		n.client.Jar.SetCookies(u, *cookie)
		body, err = getHTML(address+"/forum/search.php", n)
		if err == nil {
			if !bytes.ContainsAny(body, login) {
				err = fmt.Errorf("Wrong cookies")
			}
		}
	}
	if err != nil {
		fmt.Println(err)
		urlPost := address + "/forum/login.php"
		form := url.Values{}
		form.Set("username", login)
		form.Add("password", pass)
		form.Add("redirect", "")
		form.Add("login", "âõîä")
		req, _ := http.NewRequest(http.MethodPost, urlPost, bytes.NewBufferString(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))
		_, err = n.client.Do(req)
		if err != nil {
			return nil, err
		}
		u, _ := url.Parse(address + "/forum/")
		err = save(n.client.Jar.Cookies(u))
	}
	return n, err
}

// getHTML get body from href
func getHTML(href string, n *NC) ([]byte, error) {
	time.Sleep(500 * time.Millisecond)
	u, err := url.Parse(href)
	if err != nil {
		return nil, err
	}
	resp, err := n.client.Get(href)
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
		var ch = toUtf(char)
		fmt.Fprintf(buffer, "%c", ch)
	}
	doc := buffer.Bytes()
	doc = replaceAll(doc, "&nbsp;", " ")
	doc = replaceAll(doc, "&amp;", "&")
	doc = replaceAll(doc, "&quot;", "'")
	doc = replaceAll(doc, "<br />", " ")
	doc = replaceAll(doc, "</span>:", ":</span>")
	doc = replaceAll(doc, "  ", " ")
	doc = removeTag(doc, `<span style="text-decoration:.+?">(.+?)</span>`)
	doc = removeTag(doc, `<span style="color:.+?">(.+?)</span>`)

	if n.debugNet {
		if href == n.baseAddress+"/forum/search.php" {
			err = ioutil.WriteFile("search.html", doc, 0600)
			if err != nil {
				return nil, err
			}
		}
		if href == n.baseAddress+"/forum/login.php" {
			err = ioutil.WriteFile("login.html", doc, 0600)
			if err != nil {
				return nil, err
			}
		}
		q := u.Query()
		t := q.Get("t")
		f := q.Get("f")
		if t != "" {
			err = ioutil.WriteFile(t+".html", doc, 0600)
		} else {
			if f != "" {
				err = ioutil.WriteFile(f+".html", doc, 0600)
			} else {
				err = ioutil.WriteFile(u.Path+".html", doc, 0600)
			}
		}
	}
	return doc, err
}

// ParseForumTree get topics from forumTree
func (n *NC) ParseForumTree(href string) ([]Topic, error) {
	var (
		topics []Topic
		reTree = regexp.MustCompile(
			`<a href="viewtopic.php\?t=(\d+).*?"class="topictitle">(.+?)\s\((\d{4})\)\s(.+?)</a>`,
		)
		// reAttrib = regexp.MustCompile(`\"Seeders\"><b>(\d*?)<.+?\"Leechers\"><b>(\d*?)<.+?<a href="(.+?)".+?>(.+?)</a`)
	)
	body, err := getHTML(n.baseAddress+href, n)
	if err != nil {
		return topics, err
	}
	if !reTree.Match(body) {
		return topics, fmt.Errorf("No topic in body")
	}
	findResult := reTree.FindAllSubmatch(body, -1)
	// findAttrib := reAttrib.FindAllSubmatch(body, -1)
	for _, v := range findResult {
		var t Topic
		t.Href = string(v[1])
		t.Name = string(v[2])
		t.Year = string(v[3])
		t.Quality = string(v[4])
		// t.Seeders, _ = strconv.Atoi(string(findAttrib[i][1]))
		// t.Leechers, _ = strconv.Atoi(string(findAttrib[i][2]))
		topics = append(topics, t)
	}
	return topics, nil
}

// ParseTopic get film from topic
func (n *NC) ParseTopic(topic Topic) (Film, error) {
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
	case 4:
		film.Name = strings.Trim(name[0], " ")
		film.EngName = strings.Trim(name[2], " ")
	}
	film.Name = strings.Replace(film.Name, ":", ": ", -1)
	film.Name = strings.Replace(film.Name, "  ", " ", -1)
	film.EngName = strings.Replace(film.EngName, ":", ": ", -1)
	film.EngName = strings.Replace(film.EngName, "  ", " ", -1)
	film.Href = topic.Href
	if year, err := strconv.Atoi(topic.Year); err == nil {
		film.Year = year
	}
	body, err := getHTML(n.baseAddress+"/forum/viewtopic.php?t="+film.Href, n)
	if err != nil {
		return film, err
	}
	topic.Body = body
	film.Section = topic.getSection()
	film.Country, film.RawCountry = topic.getCountry()
	film.Genre = topic.getGenre()
	film.Director = topic.getDirector()
	film.Producer = topic.getProducer()
	film.Actor = topic.getActor()
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
	film.Audio1 = topic.getAudio1()
	film.Audio2 = topic.getAudio2()
	film.Audio3 = topic.getAudio3()
	film.Torrent = topic.getTorrent()
	film.Magnet = topic.getMagnet()
	film.DateCreate = topic.getDate()
	film.Size = topic.getSize()
	film.NNM = topic.getRating()
	film.Poster = topic.getPoster()
	film.Seeders = topic.getSeeds()
	film.Leechers = topic.getLeechs()
	return film, nil
}

func replaceAll(body []byte, from, to string) []byte {
	var reStr = regexp.MustCompile(from)
	if !reStr.Match(body) {
		return body
	}
	return reStr.ReplaceAll(body, []byte(to))
}

func replaceDate(s string) string {
	s = strings.Replace(s, " Янв ", ".01.", -1)
	s = strings.Replace(s, " января ", ".01.", -1)
	s = strings.Replace(s, " Фев ", ".02.", -1)
	s = strings.Replace(s, " февраля ", ".02.", -1)
	s = strings.Replace(s, " Мар ", ".03.", -1)
	s = strings.Replace(s, " марта ", ".03.", -1)
	s = strings.Replace(s, " Апр ", ".04.", -1)
	s = strings.Replace(s, " апреля ", ".04.", -1)
	s = strings.Replace(s, " Май ", ".05.", -1)
	s = strings.Replace(s, " мая ", ".05.", -1)
	s = strings.Replace(s, " Июн ", ".06.", -1)
	s = strings.Replace(s, " июня ", ".06.", -1)
	s = strings.Replace(s, " Июл ", ".07.", -1)
	s = strings.Replace(s, " июля ", ".07.", -1)
	s = strings.Replace(s, " Авг ", ".08.", -1)
	s = strings.Replace(s, " августа ", ".08.", -1)
	s = strings.Replace(s, " Сен ", ".09.", -1)
	s = strings.Replace(s, " сентября ", ".09.", -1)
	s = strings.Replace(s, " Окт ", ".10.", -1)
	s = strings.Replace(s, " октября ", ".10.", -1)
	s = strings.Replace(s, " Ноя ", ".11.", -1)
	s = strings.Replace(s, " ноября ", ".11.", -1)
	s = strings.Replace(s, " Дек ", ".12.", -1)
	s = strings.Replace(s, " декабря ", ".12.", -1)
	split := strings.Split(s, ".")
	if len(split[0]) == 1 {
		s = "0" + s
	}
	return s
}

// func caseInsensitiveContains(s, substr string) bool {
// 	s, substr = strings.ToUpper(s), strings.ToUpper(substr)
// 	return strings.Contains(s, substr)
// }

func cleanStr(str string) string {
	var reSpan = regexp.MustCompile("<span .*?>")
	str = reSpan.ReplaceAllString(str, "")
	str = strings.Replace(str, "</span>", "", -1)
	str = strings.Trim(str, " ")
	return str
}

func removeTag(body []byte, tag string) []byte {
	var reTag = regexp.MustCompile(tag)
	if !reTag.Match(body) {
		return body
	}
	tags := reTag.FindAllSubmatch(body, -1)
	for _, item := range tags {
		body = bytes.Replace(body, item[0], item[1], 1)
	}
	return body
}

func stringToStruct(in string) []string {
	var out []string
	itemsArray := strings.Split(in, ",")
	for _, item := range itemsArray {
		trimItem := strings.Trim(item, "\" ")
		if item != "" {
			out = append(out, trimItem)
		}
	}
	return out
}

// Encode via Gob to file
func save(object interface{}) error {
	file, err := os.Create("acc.gb")
	if err == nil {
		encoder := gob.NewEncoder(file)
		err = encoder.Encode(object)
	}
	_ = file.Close()
	return err
}

// Decode Gob file
func load(object interface{}) error {
	if !existsFile("acc.gb") {
		return fmt.Errorf("File acc.gb not exist")
	}
	file, err := os.Open("acc.gb")
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(object)
	}
	_ = file.Close()
	return err
}

func existsFile(file string) bool {
	_, err := os.Stat(file)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func findInt(b []byte, reg string) int {
	var (
		ret int
		re  = regexp.MustCompile(reg)
	)
	if re.Match(b) {
		s := re.FindSubmatch(b)
		ret, _ = strconv.Atoi(string(s[1]))
	}
	return ret
}

func findArrayOfStr(b []byte, reg string) []string {
	var (
		ret []string
		re  = regexp.MustCompile(reg)
	)
	if re.Match(b) {
		str := string(re.FindSubmatch(b)[1])
		str = cleanStr(str)
		str = strings.Trim(str, ".")
		ret = stringToStruct(str)
	}
	return ret
}

func findStr(b []byte, reg string) string {
	return cleanStr(findStrNoClean(b, reg))
}

func findStrNoClean(b []byte, reg string) string {
	var (
		ret string
		re  = regexp.MustCompile(reg)
	)
	if re.Match(b) {
		ret = string(re.FindSubmatch(b)[1])
		ret = cleanStr(ret)
	}
	return ret
}
