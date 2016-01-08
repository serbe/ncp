package nnmc

import (
	"fmt"
	"regexp"
	"strconv"
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
	Sound       string
	Size        string
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
	var film Film
	var reTopic = regexp.MustCompile(`<span style="font-weight: bold">(Производство|Жанр|Режиссер|Продюсер|Актеры|Описание|Возраст|Дата мировой премьеры|Дата премьеры в России|Продолжительность|Качество видео|Перевод|Вид субтитров|Видео|Аудио):<\/span>(.+?)<br \/>`)
	var reTr = regexp.MustCompile(`(?s)<tr\sclass="row1">(.+?)</tr>`)
	var reTd = regexp.MustCompile(`(?s)<td\sclass="genmed">(.+?)</td>`)
	if reTopic.Match(body) == false {
		return film, fmt.Errorf("No topic in body")
	}
	findAttrs := reTopic.FindAllSubmatch(body, -1)
	for _, v := range findAttrs {
		one := string(v[1])
		two := string(v[2])
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

	if reTr.Match(body) == false {
		return film, fmt.Errorf("No <tr> in body")
	}
	findTr := reTr.FindAllSubmatch(body, -1)
	for _, tr := range findTr {
		if reTd.Match(tr[1]) == false {
			findTd := reTd.FindAllSubmatch(body, -1)
			fmt.Println(len(findTd))
			for _, v := range findTd {
				fmt.Println(string(v[1]))
			}
		}
	}

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
