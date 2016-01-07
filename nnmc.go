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

// ParseForumTree get topics from forumTree
func ParseForumTree(body []byte) ([]Topic, error) {
	var reTopic = regexp.MustCompile(`<a href="(viewtopic.php\?t=\d+)"class="topictitle">(.+?)\s\((\d{4})\)\s(.+?)</a>`)
	var topics []Topic
	if reTopic.Match(body) == false {
		return topics, fmt.Errorf("No topic in body")
	}
	findResult := reTopic.FindAllSubmatch(body, -1)

	for v := range findResult {
		var t Topic
		t.href = string(findResult[v][1])
		t.text = string(findResult[v][2])
		t.year = string(findResult[v][3])
		t.quality = string(findResult[v][4])
		topics = append(topics, t)
	}
	return topics, nil
}
