package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

var ShutdownServer bool = false

func StartHttpServer() (l net.Listener, err error) {
	l, err = net.Listen("tcp", fmt.Sprintf(":%d", 5454))
	if err != nil {
		return l, err
	}

	route := http.NewServeMux()

	route.HandleFunc("/", authTWHandler)
	route.HandleFunc("/access_token", accessTokenHandler)
	route.HandleFunc("/js/main.js", mainJSHandler)

	go func() {
		http.Serve(l, route)
	}()

	//srv := &http.Server{Addr: ":5454"}
	//
	//http.HandleFunc("/", authTWHandler)
	//http.HandleFunc("/access_token", accessTokenHandler)
	//http.HandleFunc("/js/main.js", mainJSHandler)
	//
	//go func() {
	//	if err := srv.ListenAndServe(); err != nil {
	//		log.Printf("Httpserver: ListenAndServe() error: %s", err)
	//	}
	//}()
	ShutdownServer = false
	return l, nil
}

func mainJSHandler(w http.ResponseWriter, r *http.Request) {
	file, _ := ioutil.ReadFile("js/main.js")
	fmt.Fprint(w, string(file))
}

func authTWHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("./view/auth.html")
	t.Execute(w, nil)
}

func accessTokenHandler(w http.ResponseWriter, r *http.Request) {
	accessToken := r.FormValue("access_token")
	t, _ := template.ParseFiles("./view/at.html")
	t.Execute(w, nil)
	fmt.Println("AccessToken: ", accessToken)

	dataBase, err := initDB()
	if err != nil {
		log.Panic(err)
	}
	dataBase.InsertAccessToken(accessToken)
	dataBase.db.Close()
	ShutdownServer = true
}
