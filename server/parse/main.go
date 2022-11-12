package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"learn-grpc/config"
	"learn-grpc/services/parse"
	"log"
	"net"
	"net/http"
	"sync"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

const totalPages = 50
const url = "https://gorest.co.in/public/v1/posts?page="

type parseData struct {
	Meta struct {
		Pagination struct {
			Total int `json:"total"`
			Pages int `json:"pages"`
			Page  int `json:"page"`
			Limit int `json:"limit"`
			Links struct {
				Previous string `json:"previous"`
				Current  string `json:"current"`
				Next     string `json:"next"`
			} `json:"links"`
		} `json:"pagination"`
	} `json:"meta"`
	Data []Post `json:"data"`
}

type Post struct {
	Id     int    `json:"id"`
	UserId int    `json:"user_id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

type parseServer struct {
	parse.UnimplementedParseServiceServer
}

var db *sql.DB

func initDB(cfg *config.Config) error {
	d, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s", cfg.DB_Host, cfg.DB_Port, cfg.DB_Username, cfg.DB_Name, cfg.DB_Password, cfg.DB_SSLMode))
	if err != nil {
		return err
	}
	if err := d.Ping(); err != nil {
		return err
	}
	db = d
	return nil
}

func createPostTable() error {
	query := `CREATE TABLE post(
		id int,
		user_id int,
		title text,
		body text
	);`
	if _, err := db.Exec(query); err != nil {
		return err
	}
	return nil
}

func main() {
	cfg, err := config.InitConfig("./config.json")
	if err != nil {
		log.Fatal(err)
	}
	initDB(cfg)
	createPostTable()

	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatal(err)
	}
	srv := grpc.NewServer()
	parse.RegisterParseServiceServer(srv, &parseServer{})
	log.Printf("starting parse server at %v", lis.Addr())
	if err := srv.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func (s *parseServer) Parse(ctx context.Context, in *parse.EmptyRequest) (*parse.Status, error) {
	client := http.Client{}
	var wg sync.WaitGroup
	var mu sync.Mutex
	wg.Add(totalPages)
	for i := 1; i <= totalPages; i++ {
		go func(i int) {
			defer wg.Done()
			resp, err := client.Get(fmt.Sprintf("%s%d", url, i))
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()
			var data parseData
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				log.Fatal(err)
			}
			mu.Lock()
			if err := insertToPostTable(data.Data); err != nil {
				log.Fatal(err)
			}
			mu.Unlock()
		}(i)
	}
	wg.Wait()
	return &parse.Status{Info: "Ok"}, nil
}

func insertToPostTable(posts []Post) error {
	query := "INSERT INTO post(id, user_id, title, body) VALUES ($1, $2, $3, $4);"
	for _, post := range posts {
		if _, err := db.Exec(query, post.Id, post.UserId, post.Title, post.Body); err != nil {
			return err
		}
	}
	return nil
}
