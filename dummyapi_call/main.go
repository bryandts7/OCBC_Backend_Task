package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

type Posts struct {
	Data []Post `json:"data"`
}

type Post struct {
	Text        string   `json:"text"`
	Tags        []string `json:"tags"`
	PublishDate string   `json:"publishDate"`
	Likes       int      `json:"likes"`
	Owner       struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	} `json:"owner"`
}

type Users struct {
	Data []User `json:"data"`
}

type User struct {
	Id        string `json:"id"`
	Title     string `json:"title"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Gender    string `json:"gender"`
}

const (
	apiBaseURL = "https://dummyapi.io/data/v1"
	appID      = "6671681cbcf0d740072e040b"
)

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "A simple CLI app to scrape data from Dummy API",
}

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Scrape data from Dummy API",
	Run:   runWorker,
}

func main() {
	rootCmd.AddCommand(workerCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runWorker(cmd *cobra.Command, args []string) {
	var wg sync.WaitGroup
	client := resty.New().SetHeader("app-id", appID)

	for i := 1; i <= 10; i++ {
		wg.Add(1)
		getUsers(client, i, &wg)

		wg.Add(1)
		go getPosts(client, i, &wg)

	}
	wg.Wait()
}

func getUsers(client *resty.Client, page int, wg *sync.WaitGroup) {
	defer wg.Done()
	resp, err := client.R().
		SetQueryParam("page", fmt.Sprintf("%d", page)).
		SetQueryParam("limit", "10").
		Get(fmt.Sprintf("%s/user", apiBaseURL))
	if err != nil {
		log.Printf("Error fetching users on page %d: %v\n", page, err)
		return
	}

	var result Users
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		log.Printf("Error unmarshalling users response on page %d: %v\n", page, err)
		return
	}

	var userWg sync.WaitGroup
	for _, user := range result.Data {
		userWg.Add(1)
		go getUserDetails(client, user.Id, &userWg)
	}
	userWg.Wait()
}

func getUserDetails(client *resty.Client, userId string, userWg *sync.WaitGroup) {
	defer userWg.Done()
	resp, err := client.R().
		Get(fmt.Sprintf("%s/user/%s", apiBaseURL, userId))
	if err != nil {
		log.Printf("Error fetching user %s: %v\n", userId, err)
		return
	}

	var user User
	err = json.Unmarshal(resp.Body(), &user)
	if err != nil {
		log.Printf("Error unmarshalling user response on userId %s: %v\n", userId, err)
		return
	}

	fmt.Printf("User name %s %s %s %s %s\n", user.Title, user.FirstName, user.LastName, user.Email, user.Gender)

}

func getPosts(client *resty.Client, page int, wg *sync.WaitGroup) {
	defer wg.Done()
	resp, err := client.R().
		SetQueryParam("page", fmt.Sprintf("%d", page)).
		SetQueryParam("limit", "10").
		Get(fmt.Sprintf("%s/post", apiBaseURL))
	if err != nil {
		log.Printf("Error fetching posts on page %d: %v\n", page, err)
	}

	var result Posts

	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		log.Printf("Error unmarshalling posts response on page %d: %v\n", page, err)
	}

	for _, post := range result.Data {
		fmt.Printf("Posted by %s %s:\n\n%s\n\nLikes %d Tags %v\nDate posted %s\n",
			post.Owner.FirstName, post.Owner.LastName, post.Text, post.Likes, post.Tags, post.PublishDate)
	}
}
