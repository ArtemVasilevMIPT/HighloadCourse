package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const keyServerAddr = "serverAddr"

func get(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "GET" {
		resp := make(map[string][]string)
		data := strings.Split(time.Now().Format(time.DateOnly), "-")
		for i := 0; i < 10000; i += 1 {
			resp[strconv.Itoa(i)] = data
		}
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}
		w.Write(jsonResp)
	} else if r.Method == "POST" {
		name := r.PostFormValue("name")
		println("Name: ", name)
		resp := make(map[string]string)
		for i := 0; i < 10000; i += 1 {
			resp[strconv.Itoa(i)] = name
		}
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}
		w.Write(jsonResp)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", get)
	ctx, cancelCtx := context.WithCancel(context.Background())
	serverOne := &http.Server{Addr: ":3333", Handler: mux, BaseContext: func(l net.Listener) context.Context {
		ctx = context.WithValue(ctx, keyServerAddr, l.Addr().String())
		return ctx
	}}
	serverTwo := &http.Server{Addr: ":4444", Handler: mux, BaseContext: func(l net.Listener) context.Context {
		ctx = context.WithValue(ctx, keyServerAddr, l.Addr().String())
		return ctx
	}}
	go func() {
		err := serverOne.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("Server one closed\n")
		} else if err != nil {
			fmt.Printf("error starting server one: %s\n", err)
		}
		cancelCtx()
	}()
	go func() {
		err := serverTwo.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("Server two closed\n")
		} else if err != nil {
			fmt.Printf("error starting server two: %s\n", err)
		}
		cancelCtx()
	}()
	<-ctx.Done()
}
