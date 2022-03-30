package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type vtuberInfo struct {
	Id string `json:"id"`
	Name string `json:"name"`
}

type jsonList struct {
	List []vtuberInfo `json"list"`
}

func Seed(w http.ResponseWriter, _ *http.Request) {
	var vtuberList jsonList
	raw, err := ioutil.ReadFile("./sample_VtuberList.json")
    if err != nil {
        fmt.Println(err.Error())
        os.Exit(1)
    }
	json.Unmarshal(raw, &vtuberList)
	stmt, err := DB.Prepare("INSERT INTO vtubers(id, name) VALUES(?,?)")
	if err != nil {
		log.Fatal(err)
	}
	for _, vtuber := range vtuberList.List {
		fmt.Printf("name=%s\n", vtuber.Name)
		_, err = stmt.Exec(vtuber.Id, vtuber.Name)
		if err != nil {
			log.Fatal(err)
		}
	}
}
