package main

import (
	"context"
	"encoding/json"
	"learn-grpc/services/parse"
	"learn-grpc/services/post"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

type handler struct {
	parseClient parse.ParseServiceClient
	postClient  post.PostServiceClient
}

func main() {
	parseConn, err := grpc.Dial("localhost:8081", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatal(err)
	}
	defer parseConn.Close()
	parseClient := parse.NewParseServiceClient(parseConn)

	postConn, err := grpc.Dial("localhost:8082", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatal(err)
	}
	defer parseConn.Close()
	postClient := post.NewPostServiceClient(postConn)

	h := handler{
		parseClient: parseClient,
		postClient:  postClient,
	}

	mux := mux.NewRouter()
	mux.HandleFunc("/parse", h.parseEndpoint).Methods(http.MethodGet)
	mux.HandleFunc("/posts", h.getAllPostsEndpoint).Methods(http.MethodGet)
	mux.HandleFunc("/post/{id}", h.getPostByIdEndpoint).Methods(http.MethodGet)
	mux.HandleFunc("/post/{id}", h.deletePostByIdEndpoint).Methods(http.MethodDelete)
	mux.HandleFunc("/post/{id}", h.updatePostByIdEndpoint).Methods(http.MethodPut)
	srv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	log.Printf("starting api gateway on http://localhost%s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func (h *handler) parseEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute)
	defer cancel()
	status, err := h.parseClient.Parse(ctx, &parse.EmptyRequest{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *handler) getAllPostsEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute)
	defer cancel()
	resp, err := h.postClient.GetPosts(ctx, &post.EmptyRequest{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *handler) getPostByIdEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute)
	defer cancel()
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	post, err := h.postClient.GetPost(ctx, &post.PostRequest{Id: int32(id)})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(post); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *handler) deletePostByIdEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute)
	defer cancel()
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	status, err := h.postClient.DeletePost(ctx, &post.PostRequest{Id: int32(id)})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type updatePost struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (h *handler) updatePostByIdEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx, cancel := context.WithTimeout(r.Context(), time.Minute)
	defer cancel()
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	var update updatePost
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	updateReq := &post.UpdatePostRequest{
		Id:    int32(id),
		Title: &update.Title,
		Body:  &update.Body,
	}
	status, err := h.postClient.UpdatePost(ctx, updateReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
