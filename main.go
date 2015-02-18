package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"text/template"
)

const subredditsFile = "subreddits"

type post struct {
	Title string
	Url   string
}

type subreddit struct {
	Name  string
	Posts []post
}

func main() {
	subNames := parseSubredditsFile()
	subreddits := make([]subreddit, 0)
	for _, sub := range subNames {
		posts := getPosts(getSubredditJson(sub))
		subreddits = append(subreddits, subreddit{sub, posts})
	}
	sendEmail("shaneburkhart@gmail.com", "shaneburkhart@gmail.com", renderEmailHtml(subreddits))
}

func parseSubredditsFile() []string {
	bytes, err := ioutil.ReadFile(subredditsFile)
	if err != nil {
		log.Fatal(err)
	}
	return strings.Split(strings.TrimSpace(string(bytes)), "\n")
}

func getSubredditJson(subreddit string) map[string]interface{} {
	url := "http://reddit.com/r/" + subreddit + "/top.json"

	r, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	var j interface{}
	err = json.Unmarshal(b, &j)
	if err != nil {
		log.Fatal(err)
	}

	return j.(map[string]interface{})
}

func getPosts(json map[string]interface{}) []post {
	listing := json["data"].(map[string]interface{})
	children := listing["children"].([]interface{})

	posts := make([]post, 0, 25)
	for _, c := range children {
		p := c.(map[string]interface{})
		pData := p["data"].(map[string]interface{})
		ps := post{pData["title"].(string), pData["url"].(string)}
		posts = append(posts, ps)
	}

	return posts
}

func sendEmail(to string, from string, body []byte) {
	auth := smtp.PlainAuth("", "shaneburkhart@gmail.com", "password", "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, body)
	if err != nil {
		log.Fatal(err)
	}
}

func renderEmailHtml(subs []subreddit) []byte {
	t, err := template.New("email.html").ParseFiles("./email.html")
	if err != nil {
		log.Fatal(err)
	}

	buf := new(bytes.Buffer)
	err = t.Execute(buf, subs)
	if err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()
}
