package main

import (
	"encoding/json"
	"fmt"
	. "hbre2ip/models"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	if err := doMain(); err != nil {
		fmt.Println(err)
	}
}

var ApiRoot = "https://www.instapaper.com/api/add"

/*
 *
 */
func doMain() error {

	secret, err := getSecret()
	if err != nil {
		return err
	}

	history, err := getHistory()
	if err != nil {
		return err
	}

	entries, err := getEntries()
	if err != nil {
		return err
	}

	newHistory, err := postToInstapaper(secret, history, entries)
	err = saveHistory(newHistory)

	return err
}

/*
 *
 */
func getSecret() (*Secret, error) {

	dir, err := getConfigDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "secret.json")

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	secret := &Secret{}
	err = json.Unmarshal(b, secret)
	if err != nil {
		return nil, fmt.Errorf("failed to load secret.json : %w", err)
	}

	return secret, nil
}

/*
 *
 */
func getHistory() (*History, error) {

	dir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(dir, "history.json")
	if _, err := os.Stat(path); err != nil {
		return &History{Entries: []Entry{}}, nil
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	history := &History{}
	err = json.Unmarshal(b, history)
	if err != nil {
		return nil, fmt.Errorf("failed to load history.json : %w", err)
	}

	return history, nil
}

/*
 *
 */
func getEntries() ([]Entry, error) {

	res, err := http.Get("https://hatenablog.com")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	entries := []Entry{}
	container := doc.Find("div.serviceTop-staffPicks-list")
	container.Children().Each(func(i int, s *goquery.Selection) {
		href, isOK := s.Find("a").Attr("href")
		if isOK {
			title := strings.TrimSpace(s.Find(".entry-title").Text())
			entries = append(entries, Entry{Title: title, Url: href})
		}
	})

	return entries, nil
}

/*
 *
 */
func postToInstapaper(secret *Secret, history *History, entries []Entry) (*History, error) {

	newHistory := &History{}

	for _, entry := range entries {

		newHistory.Add(entry)
		if history.Contains(entry) {
			fmt.Println("already posted : " + entry.Title)
			continue
		}

		values := url.Values{
			"username": []string{secret.UserName},
			"password": []string{secret.Password},
			"url":      []string{entry.Url},
		}
		r := strings.NewReader(values.Encode())

		fmt.Println("post :", entry.Title+" - "+entry.Url)

		res, err := http.Post(ApiRoot, "application/x-www-form-urlencoded", r)
		if err != nil {
			return nil, err
		}

		rb, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		fmt.Println("   ", res.StatusCode, ":", string(rb))
	}

	return newHistory, nil
}

/*
 *
 */
func saveHistory(history *History) error {
	dir, err := getConfigDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "history.json")

	b, err := json.Marshal(history)
	if err != nil {
		return err
	}

	return os.WriteFile(path, b, os.ModePerm)
}

/*
 *
 */
func getConfigDir() (string, error) {

	dir := os.Getenv("HBRE2IP_PATH")
	if dir == "" {
		exePath, err := os.Executable()
		if err != nil {
			return "", err
		}
		dir = filepath.Dir(exePath)
	}

	return dir, nil
}
