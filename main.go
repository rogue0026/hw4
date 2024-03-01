package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type DatasetRow struct {
	XMLName       xml.Name  `xml:"row"`
	ID            int       `xml:"id"`
	Guid          string    `xml:"guid"`
	IsActive      bool      `xml:"isActive"`
	Balance       string    `xml:"balance"`
	Picture       string    `xml:"picture"`
	Age           int       `xml:"age"`
	EyeColor      string    `xml:"eyeColor"`
	FirstName     string    `xml:"first_name"`
	LastName      string    `xml:"last_name"`
	Gender        string    `xml:"gender"`
	Company       string    `xml:"company"`
	Email         string    `xml:"email"`
	Phone         string    `xml:"phone"`
	Address       string    `xml:"address"`
	About         string    `xml:"about"`
	Registered    time.Time `xml:"registered"`
	FavoriteFruit string    `xml:"favoriteFruit"`
}

type DataSet struct {
	XMLName xml.Name `xml:"root"`
	Rows    []DatasetRow
}

func main() {
	// s := httptest.NewServer(http.HandlerFunc(SearchServer))
	// s.Start()
	f, err := os.Open("dataset.xml")
	if err != nil {
		panic(err)
	}
	xmlDecoder := xml.NewDecoder(f)
	d := make([]DatasetRow, 0)
	for i := 0; i < len(d); i++ {
		err = xmlDecoder.Decode(&d)
		if err != nil {
			log.Fatalf("ERROR: %s\n", err.Error())
		}
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

	// search info in xml file
	// f, err := os.Open("dataset.xml")
	// if err != nil {
	// 	log.Fatalln(err.Error())
	// }
	// io.ReadAll()
}
