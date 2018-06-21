package ncp

import (
	"regexp"
	"strconv"
	"strings"
)

func (t *Topic) getSection() string {
	return findStr(t.Body, `<a href="viewforum.php\?f=\d+?" class="nav">(.+?)</a>`)
}

func (t *Topic) getRating() float64 {
	var (
		str    = findStrNoClean(t.Body, `>(\d,\d|\d)</span>.+?\(Голосов:`)
		rating float64
	)
	if str != "" {
		str = strings.Replace(str, ",", ".", -1)
		rating, _ = strconv.ParseFloat(str, 64)
	}
	return rating
}

func (t *Topic) getSize() int {
	var (
		str  = findStrNoClean(t.Body, `Размер блока: \d.+?B"> (\d{1,2},\d{1,2}|\d{3,4}|\d{1,2})\s`)
		size int
	)
	if str != "" {
		str = strings.Replace(str, ",", ".", -1)
		if s64, err := strconv.ParseFloat(str, 64); err == nil {
			if s64 < 100 {
				s64 = s64 * 1000
			}
			size = int(s64)
		}
	}
	return size
}

func (t *Topic) getTorrent() string {
	return findStr(t.Body, `<a href="download\.php\?id=(\d{5,7})" rel="nofollow">Скачать<`)
	// http://nnm-club.me/forum/download.php?id=
}

func (t *Topic) getMagnet() string {
	return findStr(t.Body, `href="magnet:\?xt=urn:btih:(.+?)(?:"|&)`)
	// magnet:?xt=urn:btih:
}

func (t *Topic) getPoster() string {
	return findStrNoClean(t.Body, `"postImg postImgAligned img-right" title="http://assets\..+?/forum/image\.php\?link=(.+?(?:jpg|jpeg|png))"`)
}

func (t *Topic) getDate() string {
	return replaceDate(findStrNoClean(t.Body, `> (\d{1,2} .{3} \d{4}).{9}<`))
}

func (t *Topic) getSeeds() int {
	return findInt(t.Body, `<span class="seed">\[ <b>(\d{1,5})\s`)
}

func (t *Topic) getLeechs() int {
	return findInt(t.Body, `<span class="leech">\[ <b>(\d{1,5})\s`)
}

func getResolution(str string) string {
	var (
		reRes      = regexp.MustCompile(`(\d{3,4}x\d{3,4}|\d{3,4}X\d{3,4}|\d{3,4}х\d{3,4}|\d{3,4}Х\d{3,4})`)
		resolution string
	)
	if reRes.MatchString(str) {
		resolution = reRes.FindString(str)
	}
	return resolution
}

func (t *Topic) getCountry() (countryName []string, rawCountry string) {
	rawCountry = findStr(t.Body, `<span style="font-weight: bold">(?:Производство|Страна):\s*</span>(.+?)<`)
	lowerRawCountry := strings.ToLower(rawCountry)
	for _, item := range counriesList {
		i := strings.Index(lowerRawCountry, strings.ToLower(item))
		if i != -1 && item != "" {
			countryName = append(countryName, item)
			lowerRawCountry = lowerRawCountry[:i] + lowerRawCountry[i+len(item):]
		}
	}
	return
}

func (t *Topic) getGenre() []string {
	return findArrayOfStr(t.Body, `<span style="font-weight: bold">Жанр:\s*</span>(.+?)<`)
}

func (t *Topic) getDirector() []string {
	return findArrayOfStr(t.Body, `<span style="font-weight: bold">Режиссер:\s*</span>(.+?)<`)
}

func (t *Topic) getProducer() []string {
	return findArrayOfStr(t.Body, `<span style="font-weight: bold">Продюсер:\s*</span>(.+?)<`)
}

func (t *Topic) getActor() []string {
	return findArrayOfStr(t.Body, `<span style="font-weight: bold">Актеры:\s*</span>(.+?)<`)
}

func (t *Topic) getDescription() string {
	return findStr(t.Body, `<span style="font-weight: bold">(?:Описание фильма|Описание мультфильма|Описание|О фильме):\s*</span>(.+?)<`)
}

func (t *Topic) getAge() string {
	return findStr(t.Body, `<span style="font-weight: bold">Возраст:\s*</span>(.+?)<`)
}

func (t *Topic) getReleaseDate() string {
	var ret = findStr(t.Body, `<span style="font-weight: bold">Дата мировой премьеры:\s*</span>(.+?)<`)
	ret = replaceDate(ret)
	return ret
}

func (t *Topic) getRussianDate() string {
	var ret = findStr(t.Body, `<span style="font-weight: bold">(?:Дата премьеры в России|Дата Российской премьеры|Дата российской премьеры):\s*</span>(.+?)<`)
	ret = replaceDate(ret)
	return ret
}

func (t *Topic) getDuration() string {
	var duration = findStr(t.Body, `<span style="font-weight: bold">Продолжительность:\s*</span>(.+?)<`)
	if duration == "" {
		reDuration := regexp.MustCompile(`\sПродолжительность\s+?&#58; (\d{1,2}) ч\. (\d{1,2}) м\.`)
		if reDuration.Match(t.Body) {
			submatch := reDuration.FindSubmatch(t.Body)
			hour := string(submatch[1])
			minute := string(submatch[2])
			if len(hour) == 1 {
				hour = "0" + hour
			}
			if len(minute) == 1 {
				minute = "0" + minute
			}
			duration = hour + ":" + minute + ":00"
		}
	}
	if len(duration) < 5 {
		duration = ""
	}
	return duration
}

func (t *Topic) getQuality() string {
	return findStr(t.Body, `<span style="font-weight: bold">(?:Качество видео|Качество):\s*</span>(.+?)<`)
}

func (t *Topic) getTranslation() string {
	return findStr(t.Body, `<span style="font-weight: bold">Перевод:\s*</span>(.+?)<`)
}

func (t *Topic) getSubtitlesType() string {
	return findStr(t.Body, `<span style="font-weight: bold">Вид субтитров:\s*</span>(.+?)<`)
}

func (t *Topic) getSubtitles() string {
	return findStr(t.Body, `<span style="font-weight: bold">Субтитры:\s*</span>(.+?)<`)
}

func (t *Topic) getVideo() string {
	return findStr(t.Body, `<span style="font-weight: bold">Видео:\s*</span>(.+?)<`)
}

func (t *Topic) getAudio1() string {
	return findStr(t.Body, `<span style="font-weight: bold">(?:Аудио\s?:\s*|Аудио\s?.?1.?:\s*)</span>(.+?)<`)
}

func (t *Topic) getAudio2() string {
	return findStr(t.Body, `<span style="font-weight: bold">Аудио\s?.?2.?:\s*</span>(.+?)<`)
}

func (t *Topic) getAudio3() string {
	return findStr(t.Body, `<span style="font-weight: bold">Аудио\s?.?3.?:\s*</span>(.+?)<`)
}
