package main

import (
	"fmt"
	"html/template"
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

	ShutdownServer = false
	return l, nil
}

func mainJSHandler(w http.ResponseWriter, r *http.Request) {
	data, _ := Asset("js/main.js")
	fmt.Fprint(w, string(data))
}

func authTWHandler(w http.ResponseWriter, r *http.Request) {
	data, _ := Asset("view/auth.html")
	t, _ := template.New("auth").Parse(string(data))
	t.Execute(w, nil)
}

func accessTokenHandler(w http.ResponseWriter, r *http.Request) {
	accessToken := r.FormValue("access_token")
	data, _ := Asset("view/at.html")
	t, _ := template.New("at").Parse(string(data))
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
