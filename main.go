package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

type DataSet struct {
	XMLName xml.Name `xml:"root"`
	Rows    []struct {
		ID            int    `xml:"id"`
		Guid          string `xml:"guid"`
		IsActive      bool   `xml:"isActive"`
		Balance       string `xml:"balance"`
		Picture       string `xml:"picture"`
		Age           int    `xml:"age"`
		EyeColor      string `xml:"eyeColor"`
		FirstName     string `xml:"first_name"`
		LastName      string `xml:"last_name"`
		Gender        string `xml:"gender"`
		Company       string `xml:"company"`
		Email         string `xml:"email"`
		Phone         string `xml:"phone"`
		Address       string `xml:"address"`
		About         string `xml:"about"`
		Registered    string `xml:"registered"`
		FavoriteFruit string `xml:"favoriteFruit"`
	} `xml:"row"`
}

func main() {
	// s := httptest.NewServer(http.HandlerFunc(SearchServer))
	// s.Start()
	f, err := os.Open("dataset.xml")
	if err != nil {
		panic(err)
	}
	xmlDecoder := xml.NewDecoder(f)
	ds := DataSet{}
	err = xmlDecoder.Decode(&ds)
	if err != nil {
		panic(err)
	}
	for i := 0; i < len(ds.Rows); i++ {
		fmt.Println(ds.Rows[i].FirstName + " " + ds.Rows[i].LastName)
		time.Sleep(time.Second * 1)
	}
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
		return
	}

	// parsing request params
	req := SearchRequest{}

	req.Query = r.URL.Query().Get("query")

	req.OrderField = r.URL.Query().Get("order_field")

	ord, err := strconv.Atoi(r.URL.Query().Get("order_by"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "error: %s\n", err.Error())
		return
	}
	req.OrderBy = ord

	lim, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "error: %s\n", err.Error())
		return
	}
	req.Limit = lim

	off, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "error: %s\n", err.Error())
		return
	}
	req.Offset = off

	// на основе полученных параметров нужно теперь отдать результат клиенту
	if req.Query == "Name" {
		// делаем strings.Contains() по firstы_name + last_name
		// и отдаем результат обратно клиенту
	}
	f, err := os.Open("dataset.xml")
	if err != nil {
		log.Fatalln(err.Error())
	}
	data := DataSet{}
	err = xml.NewDecoder(f).Decode(&data)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func makeSort(ds *DataSet, order_field string) {
	switch order_field {
	case "ID":
		sort.Slice(ds.Rows, func(i int, j int) bool { return ds.Rows[i].ID < ds.Rows[j].ID })
	case "Age":
		sort.Slice(ds.Rows, func(i int, j int) bool { return ds.Rows[i].Age < ds.Rows[j].Age })
	case "Name":
		sort.Slice(ds.Rows, func(i int, j int) bool {
			return ds.Rows[i].FirstName+ds.Rows[i].LastName < ds.Rows[j].FirstName+ds.Rows[j].LastName
		})
	}

}
