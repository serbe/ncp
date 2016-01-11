package nnmc

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

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
	ID          int64
	Name        string
	EngName     string
	Href        string
	Year        int64
	Genre       string
	Country     string
	Director    string
	Producer    string
	Actors      string
	Description string
	Age         string
	ReleaseDate string
	RussianDate string
	Duration    int64
	Quality     string
	Translation string
	Subtitles   string
	Video       string
	Audio       string
	Kinopoisk   float64
	Imdb        float64
	NNM         float64
	Sound       string
	Size        int64
	DateCreate  string
	Torrent     string
	Poster      string
	Hide        bool
}

// ParseForumTree get topics from forumTree
func ParseForumTree(body []byte) ([]Topic, error) {
	var reTree = regexp.MustCompile(`<a href="(viewtopic.php\?t=\d+)"class="topictitle">(.+?)\s\((\d{4})\)\s(.+?)</a>`)
	var topics []Topic
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
func ParseTopic(body []byte) (Film, error) {
	body = replaceAll(body, "&nbsp;", " ")
	var (
		film    Film
		reTopic = regexp.MustCompile(`<span style="font-weight: bold">(Производство|Жанр|Режиссер|Продюсер|Актеры|Описание|Возраст|Дата мировой премьеры|Дата премьеры в России|Продолжительность|Качество видео|Перевод|Вид субтитров|Видео|Аудио):<\/span>(.+?)<`)
		reTd    = regexp.MustCompile(`(?s)<tr class="row1">.+?<td class="genmed">.+?(Зарегистрирован|Размер|Рейтинг).+?<\/td>.+?<td class="genmed">.+?(\d{1,2},\d{1,2} GB|\d{3,4} MB|\d,\d|\d{1,2} .{3} \d{4}).+?</td>.+?</tr>`)
		reDl    = regexp.MustCompile(`<a href="download\.php\?id=(\d{5,7})" rel="nofollow">Скачать<`)
	)
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
			film.Subtitles = two
		case "Видео":
			film.Video = two
		case "Аудио":
			film.Audio = two
		}
	}
	if reTd.Match(body) == false {
		return film, fmt.Errorf("No <td> in body")
	}
	findTd := reTd.FindAllSubmatch(body, -1)
	if len(findTd) == 3 {
		film.DateCreate = string(findTd[0][2])
		size := string(findTd[1][2])
		size = strings.Replace(size, ",", ".", -1)
		fmt.Println(size)
		if s64, err := strconv.ParseInt(size, 10, 64); err == nil {
			if s64 < 80 {
				s64 = s64 * 1000
			}
			film.Size = s64
			fmt.Println(s64)
		}
		rating := string(findTd[2][2])
		rating = strings.Replace(size, ",", ".", -1)
		fmt.Println(rating)
		if r64, err := strconv.ParseFloat(rating, 64); err == nil {
			film.NNM = r64
			fmt.Println(r64)
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
