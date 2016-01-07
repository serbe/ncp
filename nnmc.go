package nnmc

import (
	"fmt"
	"regexp"
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
func ParseForumTree(body []byte) ([]RawTopic, error) {
	var reTree = regexp.MustCompile(`<a href="(viewtopic.php\?t=\d+)"class="topictitle">(.+?)\s\((\d{4})\)\s(.+?)</a>`)
	var topics []RawTopic
	if reTree.Match(body) == false {
		return topics, fmt.Errorf("No topic in body")
	}
	findResult := reTree.FindAllSubmatch(body, -1)

	for _, v := range findResult {
		var t RawTopic
		t.href = string(v[1])
		t.text = string(v[2])
		t.year = string(v[3])
		t.quality = string(v[4])
		topics = append(topics, t)
	}
	return topics, nil
}

func ParseTopic(body []byte) (Film, error) {
	var film Film
	var reTopic = regexp.MustCompile(`<span style="font-weight: bold">(Производство|Жанр|Режиссер|Продюсер|Актеры|Описание|Возраст|Дата мировой премьеры|Дата премьеры в России|Продолжительность|Качество видео|Перевод|Вид субтитров|Видео|Аудио):<\/span>(.+?)<br \/>`)
	if reTopic.Match(body) == false {
		return film, fmt.Errorf("No topic in body")
	}
	findResult := reTopic.FindAllSubmatch(body, -1)
	for _, v := range findResult {
		switch v[1] {
		case "Производство":
			film.Country = v[2]
		case "Жанр":
			film.Genre = v[2]
		case "Режиссер":
			film.Director = v[2]
		case "Продюсер":
			film.Producer = v[2]
		case "Актеры":
			film.Actors = v[2]
		case "Описание":
			film.Description = v[2]
		case "Возраст":
			film.Age = v[2]
		case "Дата мировой премьеры":
			film.ReleaseDate = v[2]
		case "Дата премьеры в России":
			film.RussianDate = v[2]
		case "Продолжительность":
			film.Duration = v[2]
		case "Качество видео":
			film.Quality = v[2]
		case "Перевод":
			film.Translation = v[2]
		case "Вид субтитров":
			film.Subtitles = v[2]
		case "Видео":
			film.Video = v[2]
		case "Аудио":
			film.Audio = v[2]
		}

	}
	return film, nil
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
