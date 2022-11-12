package main

import (
	"context"
	"database/sql"
	"fmt"
	"learn-grpc/config"
	"learn-grpc/services/post"
	"log"
	"net"
	"strings"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

type postServer struct {
	post.UnimplementedPostServiceServer
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

func main() {
	cfg, err := config.InitConfig("./config.json")
	if err != nil {
		log.Fatal(err)
	}
	initDB(cfg)

	lis, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatal(err)
	}
	srv := grpc.NewServer()
	post.RegisterPostServiceServer(srv, &postServer{})
	log.Printf("starting post server at %v", lis.Addr())
	if err := srv.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func (s *postServer) GetPosts(ctx context.Context, in *post.EmptyRequest) (*post.PostResponse, error) {
	var resp post.PostResponse
	query := "SELECT * FROM post;"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var post post.Post
		if err := rows.Scan(&post.Id, &post.UserId, &post.Title, &post.Body); err != nil {
			return nil, err
		}
		resp.Posts = append(resp.Posts, &post)
	}
	return &resp, nil
}

func (s *postServer) GetPost(ctx context.Context, in *post.PostRequest) (*post.Post, error) {
	var post post.Post
	query := "SELECT * FROM post WHERE id = $1;"
	if err := db.QueryRow(query, in.GetId()).Scan(&post.Id, &post.UserId, &post.Title, &post.Body); err != nil {
		return nil, err
	}
	return &post, nil
}

func (s *postServer) UpdatePost(ctx context.Context, in *post.UpdatePostRequest) (*post.Status, error) {
	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argId := 1
	if in.GetTitle() != "" {
		setValues = append(setValues, fmt.Sprintf("title=$%d", argId))
		args = append(args, in.GetTitle())
		argId++
	}
	if in.GetBody() != "" {
		setValues = append(setValues, fmt.Sprintf("body=$%d", argId))
		args = append(args, in.GetBody())
		argId++
	}
	setQuery := strings.Join(setValues, ", ")
	updateQuery := fmt.Sprintf("UPDATE post SET %s WHERE id = $%d;", setQuery, argId)
	args = append(args, in.GetId())
	if _, err := db.Exec(updateQuery, args...); err != nil {
		return nil, err
	}
	return &post.Status{Info: "Updated"}, nil
}

func (s *postServer) DeletePost(ctx context.Context, in *post.PostRequest) (*post.Status, error) {
	query := "DELETE FROM post WHERE id = $1"
	if _, err := db.Exec(query, in.GetId()); err != nil {
		return nil, err
	}
	return &post.Status{Info: "Deleted"}, nil
}
