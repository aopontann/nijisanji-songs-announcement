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
	raw, err := ioutil.ReadFile("./vtuberList.json")
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

func SeedOut(w http.ResponseWriter, _ *http.Request) {
	var (
		id string
		name string
	)
	var vtuberList jsonList
	rows, err := DB.Query("select id, name from vtubers")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &name)
		if err != nil {
			log.Fatal(err)
		}
		vtuberList.List = append(vtuberList.List, vtuberInfo{Id: id, Name: name})
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	file, _ := json.MarshalIndent(vtuberList, "", " ")
	_ = ioutil.WriteFile("vtuberList.json", file, 0644)
}
