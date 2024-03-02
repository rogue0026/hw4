package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
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

func main() {}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	// Разбираем параметры запроса
	searchRequest, errResp := parseParams(r)
	if errResp != nil {
		SendJSONErrRespone(w, http.StatusBadRequest, *errResp)
		return
	}
	// Если с параметрами все ок, то открывает наш xml-файл,
	// читаем из него данные и в зависимости от параметров запроса возвращаем клиенту
	f, err := os.Open("dataset.xml")
	if err != nil {
		errResp := SearchErrorResponse{Error: err.Error()}
		SendJSONErrRespone(w, http.StatusInternalServerError, errResp)
		return
	}
	data := DataSet{}
	err = xml.NewDecoder(f).Decode(&data)
	if err != nil {
		SendJSONErrRespone(w, http.StatusInternalServerError, SearchErrorResponse{Error: err.Error()})
		return
	}
	makeSort(&data, searchRequest.OrderField, searchRequest.OrderBy)
	data.Rows = data.Rows[searchRequest.Offset:searchRequest.Limit] // отсекаем лишние записи
	if searchRequest.Query == "" {                                  // вернуть все записи
		users := make([]User, 0, len(data.Rows))
		for i := 0; i < len(data.Rows); i++ {
			u := User{
				Id:     data.Rows[i].ID,
				Name:   fmt.Sprintf("%s %s", data.Rows[i].FirstName, data.Rows[i].LastName),
				Age:    data.Rows[i].Age,
				About:  data.Rows[i].About,
				Gender: data.Rows[i].Gender,
			}
			users = append(users, u)
		}
		SendJSONResponse(w, http.StatusOK, users)
	} else if searchRequest.Query == "Name" {

	}
}

func parseParams(r *http.Request) (*SearchRequest, *SearchErrorResponse) {
	req := SearchRequest{}

	req.Query = r.URL.Query().Get("query") // "Name", "About" or ""
	if req.Query != "Name" || req.Query != "About" || req.Query != "" {
		return nil, &SearchErrorResponse{Error: "bad query parameter"}
	}

	req.OrderField = r.URL.Query().Get("order_field")
	if req.OrderField != "ID" || req.OrderField != "Age" || req.OrderField != "Name" {
		return nil, &SearchErrorResponse{Error: ErrorBadOrderField}
	}

	ord := r.URL.Query().Get("order_by")
	if ord != strconv.Itoa(OrderByAsc) || ord != strconv.Itoa(OrderByAsIs) || ord != strconv.Itoa(OrderByDesc) {
		return nil, &SearchErrorResponse{Error: "bad orderBy parameter"}
	}
	req.OrderBy, _ = strconv.Atoi(ord) // не проверяем ошибку, потому что полученный параметр равен 0, 1 или -1 и он преобразуется без ошибки

	lim, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		return nil, &SearchErrorResponse{Error: "bad limit parameter"}
	}
	req.Limit = lim

	off, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		return nil, &SearchErrorResponse{Error: "bad offset parameter"}
	}
	req.Offset = off

	return &req, nil
}

func makeSort(ds *DataSet, orderField string, orderBy int) {
	if orderBy == OrderByAsc {
		switch orderField {
		case "ID":
			sort.Slice(ds.Rows, func(i int, j int) bool { return ds.Rows[i].ID < ds.Rows[j].ID })
		case "Age":
			sort.Slice(ds.Rows, func(i int, j int) bool { return ds.Rows[i].Age < ds.Rows[j].Age })
		case "Name":
			sort.Slice(ds.Rows, func(i int, j int) bool {
				return ds.Rows[i].FirstName+ds.Rows[i].LastName < ds.Rows[j].FirstName+ds.Rows[j].LastName
			})
		case "":
			sort.Slice(ds.Rows, func(i int, j int) bool {
				return ds.Rows[i].FirstName+ds.Rows[i].LastName < ds.Rows[j].FirstName+ds.Rows[j].LastName
			})
		}
	} else if orderBy == OrderByDesc {
		switch orderField {
		case "ID":
			sort.Slice(ds.Rows, func(i int, j int) bool { return ds.Rows[i].ID > ds.Rows[j].ID })
		case "Age":
			sort.Slice(ds.Rows, func(i int, j int) bool { return ds.Rows[i].Age > ds.Rows[j].Age })
		case "Name":
			sort.Slice(ds.Rows, func(i int, j int) bool {
				return ds.Rows[i].FirstName+ds.Rows[i].LastName > ds.Rows[j].FirstName+ds.Rows[j].LastName
			})
		case "":
			sort.Slice(ds.Rows, func(i int, j int) bool {
				return ds.Rows[i].FirstName+ds.Rows[i].LastName > ds.Rows[j].FirstName+ds.Rows[j].LastName
			})
		}
	}
}

func SendJSONErrRespone(w http.ResponseWriter, status int, response SearchErrorResponse) {
	js, err := json.Marshal(&response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(status)
	w.Write(js)
}

func SendJSONResponse(w http.ResponseWriter, status int, users []User) {
	js, err := json.MarshalIndent(&users, "", "    ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	w.Write(js)
}
