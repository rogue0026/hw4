package main

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

// код писать тут

type UserInfo struct {
	XMLName       xml.Name `xml:"row"`
	Id            int      `xml:"id"`
	Guid          string   `xml:"guid"`
	IsActive      bool     `xml:"isActive"`
	Balance       string   `xml:"balance"`
	Picture       string   `xml:"picture"`
	Age           int      `xml:"age"`
	EyeColor      string   `xml:"eyeColor"`
	FirstName     string   `xml:"first_name"`
	LastName      string   `xml:"last_name"`
	Gender        string   `xml:"gender"`
	Company       string   `xml:"company"`
	Email         string   `xml:"email"`
	Phone         string   `xml:"phone"`
	Address       string   `xml:"address"`
	About         string   `xml:"about"`
	Registered    string   `xml:"registered"`
	FavoriteFruit string   `xml:"favoriteFruit"`
}

type Users struct {
	XMLName xml.Name   `xml:"root"`
	Info    []UserInfo `xml:"row"`
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("AccessToken")
	if token == "" {
		SendJSONErrResponse(w, http.StatusUnauthorized, SearchErrorResponse{Error: "Bad AccessToken"})
		return
	}
	searchRequest, errResp := ParseParams(r)
	if errResp != nil {
		SendJSONErrResponse(w, http.StatusBadRequest, *errResp)
		return
	}
	f, err := os.Open("dataset.xml")
	if err != nil {
		SendJSONErrResponse(w, http.StatusInternalServerError, SearchErrorResponse{Error: err.Error()})
		return
	}

	data, err := io.ReadAll(f)
	if err != nil {
		log.Fatalf("%s\n", err.Error())
	}
	allUsers := Users{}
	err = xml.Unmarshal(data, &allUsers)
	if err != nil {
		log.Fatalf("%s\n", err.Error())
	}

	MakeSort(allUsers.Info, searchRequest.OrderField, searchRequest.OrderBy)

	if searchRequest.Offset >= len(allUsers.Info) {
		allUsers.Info = make([]UserInfo, 0)
	} else {
		allUsers.Info = allUsers.Info[searchRequest.Offset:]
	}
	if searchRequest.Limit > len(allUsers.Info) {
		searchRequest.Limit = len(allUsers.Info)
		allUsers.Info = allUsers.Info[:searchRequest.Limit]
	} else {
		allUsers.Info = allUsers.Info[:searchRequest.Limit]
	}

	results := make([]User, 0, len(allUsers.Info))
	if searchRequest.Query == "" {
		for _, elem := range allUsers.Info {
			u := User{
				Id:     elem.Id,
				Name:   elem.FirstName + " " + elem.LastName,
				Age:    elem.Age,
				About:  elem.About,
				Gender: elem.Gender,
			}
			results = append(results, u)
		}
	} else {
		for _, elem := range allUsers.Info {
			if strings.Contains(elem.FirstName+elem.LastName, searchRequest.Query) || strings.Contains(elem.About, searchRequest.Query) {
				u := User{
					Id:     elem.Id,
					Name:   elem.FirstName + " " + elem.LastName,
					Age:    elem.Age,
					About:  elem.About,
					Gender: elem.Gender,
				}
				results = append(results, u)
			}
		}
	}
	SendJSONResponse(w, http.StatusOK, results)
}

func ParseParams(r *http.Request) (*SearchRequest, *SearchErrorResponse) {
	req := SearchRequest{}

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

	validOrderFields := map[string]struct{}{"id": {}, "age": {}, "name": {}}
	orderField := r.URL.Query().Get("order_field")
	if orderField == "" {
		orderField = "name"
	}
	if _, ok := validOrderFields[orderField]; !ok {
		return nil, &SearchErrorResponse{Error: "ErrorBadOrderField"}
	} else {
		req.OrderField = orderField
	}

	validOrderBy := map[string]struct{}{
		strconv.Itoa(OrderByAsIs): {},
		strconv.Itoa(OrderByAsc):  {},
		strconv.Itoa(OrderByDesc): {},
	}
	ord := r.URL.Query().Get("order_by")
	if _, ok := validOrderBy[ord]; !ok {
		return nil, &SearchErrorResponse{Error: "bad orderBy parameter"}
	}
	ordBy, err := strconv.Atoi(ord)
	if err != nil {
		log.Println("error while parsing order by param")
	}
	req.OrderBy = ordBy

	req.Query = r.URL.Query().Get("query")

	return &req, nil
}

func MakeSort(users []UserInfo, orderField string, orderBy int) {
	if orderBy == OrderByAsc { // в порядке возрастания
		switch orderField {
		case "id":
			sort.Slice(users, func(i int, j int) bool { return users[i].Id < users[j].Id })
		case "age":
			sort.Slice(users, func(i int, j int) bool { return users[i].Age < users[j].Age })
		case "name":
			sort.Slice(users, func(i int, j int) bool {
				return users[i].FirstName+users[i].LastName < users[j].FirstName+users[j].LastName
			})
		}
	} else if orderBy == OrderByDesc {
		switch orderField {
		case "id":
			sort.Slice(users, func(i int, j int) bool { return users[i].Id > users[j].Id })
		case "age":
			sort.Slice(users, func(i int, j int) bool { return users[i].Age > users[j].Age })
		case "name":
			sort.Slice(users, func(i int, j int) bool {
				return users[i].FirstName+users[i].LastName > users[j].FirstName+users[j].LastName
			})
		}
	}
}

func SendJSONErrResponse(w http.ResponseWriter, httpStatus int, response SearchErrorResponse) {
	js, err := json.Marshal(&response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	} else {
		w.WriteHeader(httpStatus)
		w.Write(js)
	}
}

func SendJSONResponse(w http.ResponseWriter, httpStatus int, users []User) {
	js, err := json.MarshalIndent(&users, "", "    ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		w.WriteHeader(httpStatus)
		w.Write(js)
	}
}

func TestMakeSort(t *testing.T) {
	type testCase struct {
		caseName     string
		inputData    []UserInfo
		expectedData []UserInfo
		OrderBy      int
		OrderField   string
	}
	userData := []UserInfo{
		{
			Id:        9,
			FirstName: "Ivan",
			LastName:  "Ivanov",
			Age:       23,
		},
		{
			Id:        3,
			FirstName: "Vasiliy",
			LastName:  "Semibratov",
			Age:       38,
		},
		{
			Id:        6,
			FirstName: "Petr",
			LastName:  "Levandovskiy",
			Age:       27,
		},
	}

	testCases := []testCase{
		{
			caseName:  "sort by id in asc order",
			inputData: userData,
			expectedData: []UserInfo{
				{Id: 3, FirstName: "Vasiliy", LastName: "Semibratov", Age: 38},
				{Id: 6, FirstName: "Petr", LastName: "Levandovskiy", Age: 27},
				{Id: 9, FirstName: "Ivan", LastName: "Ivanov", Age: 23},
			},
			OrderBy:    OrderByAsc,
			OrderField: "id",
		},
		{
			caseName:  "sort by age in asc order",
			inputData: userData,
			expectedData: []UserInfo{
				{Id: 9, FirstName: "Ivan", LastName: "Ivanov", Age: 23},
				{Id: 6, FirstName: "Petr", LastName: "Levandovskiy", Age: 27},
				{Id: 3, FirstName: "Vasiliy", LastName: "Semibratov", Age: 38},
			},
			OrderBy:    OrderByAsc,
			OrderField: "age",
		},
		{
			caseName:  "sort by name in asc order",
			inputData: userData,
			expectedData: []UserInfo{
				{Id: 9, FirstName: "Ivan", LastName: "Ivanov", Age: 23},
				{Id: 6, FirstName: "Petr", LastName: "Levandovskiy", Age: 27},
				{Id: 3, FirstName: "Vasiliy", LastName: "Semibratov", Age: 38},
			},
			OrderBy:    OrderByAsc,
			OrderField: "name",
		},
		{
			caseName:  "sort by age in desc order",
			inputData: userData,
			expectedData: []UserInfo{
				{Id: 3, FirstName: "Vasiliy", LastName: "Semibratov", Age: 38},
				{Id: 6, FirstName: "Petr", LastName: "Levandovskiy", Age: 27},
				{Id: 9, FirstName: "Ivan", LastName: "Ivanov", Age: 23},
			},
			OrderBy:    OrderByDesc,
			OrderField: "age",
		},
	}

	for _, c := range testCases {
		MakeSort(c.inputData, c.OrderField, c.OrderBy)
		for i := 0; i < len(c.expectedData); i++ {
			if c.expectedData[i] != c.inputData[i] {
				t.Errorf("test failed: case %s\n", c.caseName)
			}
		}
	}
}

func TestParseParams(t *testing.T) {

	setParams := func(p *url.Values, req *SearchRequest) {
		p.Add("limit", strconv.Itoa(req.Limit))
		p.Add("offset", strconv.Itoa(req.Offset))
		p.Add("query", req.Query)
		p.Add("order_field", req.OrderField)
		p.Add("order_by", strconv.Itoa(req.OrderBy))
	}
	goodTestParams := SearchRequest{
		Limit:      10,
		Offset:     0,
		Query:      "some_query",
		OrderField: "age",
		OrderBy:    OrderByAsc,
	}

	params := url.Values{}
	setParams(&params, &goodTestParams)
	r := httptest.NewRequest(http.MethodGet, "http://localhost:9090/search?"+params.Encode(), nil)
	_, errResp := ParseParams(r)
	if errResp != nil {
		t.Errorf("test failed on good query params")
	}

	testParamsWithBadOrderField := SearchRequest{
		Limit:      10,
		Offset:     0,
		OrderField: "about",
		OrderBy:    OrderByAsc,
	}

	params = url.Values{}
	setParams(&params, &testParamsWithBadOrderField)

	r = httptest.NewRequest(http.MethodGet, "http://localhost:9909/search?"+params.Encode(), nil)
	_, errResp = ParseParams(r)
	if errResp != nil {
		ok := errResp.Error == ErrorBadOrderField
		if ok {
			t.Errorf("test failed on bad order field parameter")
		}
	}
}

/*
	кейсы
	3. превысить время ожидания ответа
	5. отправить 500 ответ клиенту (можно создать простой сервер, который будет просто пятисотить)
	6. отправить левый json с ошибкой
	7. отправить на сервер левое поле для сортировки записей
	8. отправить левый json с результатом
	9. отправить просто ошибку с 400 ответом
*/

func TestFindUsersForBadLimit(t *testing.T) {

	cl := SearchClient{
		AccessToken: "asdfasdf:123123",
	}

	srchReq := SearchRequest{
		Limit: -10,
	}
	_, err := cl.FindUsers(srchReq)
	if err == nil {
		t.Errorf("test for bad limid failed")
	}
}

func TestFindUsersForBadOffset(t *testing.T) {

	c := SearchClient{
		AccessToken: "asdfasdkljfhalksdhjf",
	}
	r := SearchRequest{
		Offset: -10,
	}
	_, err := c.FindUsers(r)
	if err == nil {
		t.Errorf("test for bad offset failed")
	}
}

func TestFindUsersBadAccessToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	srchReq := SearchRequest{
		Limit:      5,
		Offset:     0,
		Query:      "",
		OrderField: "id",
		OrderBy:    OrderByAsc,
	}
	cl := SearchClient{
		URL: ts.URL,
	}
	_, err := cl.FindUsers(srchReq)
	if err == nil {
		t.Errorf("test for bad access token failed")
	}
}

func TestFindUsersTimeout(t *testing.T) {
	longTimeHandler := func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second * 3)
		SearchServer(w, r)
	}
	ts := httptest.NewServer(http.HandlerFunc(longTimeHandler))
	req := SearchRequest{
		Limit:      10,
		Offset:     0,
		OrderField: "name",
		OrderBy:    OrderByAsc,
		Query:      "",
	}
	cl := SearchClient{
		AccessToken: "some token",
		URL:         ts.URL,
	}
	_, err := cl.FindUsers(req)
	if err == nil {
		t.Error("test for timeout failed")
	}
}

func TestFindUsersNormalRequest(t *testing.T) {
	goodRequest := SearchRequest{
		Limit:      10,
		Offset:     0,
		OrderField: "name",
		OrderBy:    OrderByAsIs,
		Query:      "",
	}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	cl := SearchClient{
		AccessToken: "some token",
		URL:         ts.URL,
	}
	_, err := cl.FindUsers(goodRequest)
	if err != nil {
		t.Error("test for normal request failed")
	}
}

func TestFindUsersForInternalServerError(t *testing.T) {
	goodRequest := SearchRequest{
		Limit:      10,
		Offset:     0,
		OrderField: "name",
		OrderBy:    OrderByAsIs,
		Query:      "",
	}

	internalErHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}
	ts := httptest.NewServer(http.HandlerFunc(internalErHandler))
	cl := SearchClient{
		AccessToken: "some token",
		URL:         ts.URL,
	}
	_, err := cl.FindUsers(goodRequest)
	if err == nil {
		t.Error("test for internal server error failed")
	}
}

func TestFindUsersBadRequest(t *testing.T) {
	searchRequest := SearchRequest{
		Limit:      5,
		Offset:     0,
		OrderField: "badfield",
		Query:      "",
		OrderBy:    OrderByAsc,
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	cl := SearchClient{
		AccessToken: "some token",
		URL:         ts.URL,
	}

	_, err := cl.FindUsers(searchRequest)
	if err != nil {
		t.Log(err.Error())
	}
}

func TestFindUsersForBadJSONResponse(t *testing.T) {
	erJSONHandler := func(w http.ResponseWriter, r *http.Request) {
		js, _ := json.Marshal("some value")
		w.Write(js)
	}
	ts := httptest.NewServer(http.HandlerFunc(erJSONHandler))
	cl := SearchClient{
		AccessToken: "some token",
		URL:         ts.URL,
	}
	req := SearchRequest{
		Limit:      10,
		Offset:     0,
		Query:      "somequery",
		OrderField: "id",
		OrderBy:    OrderByAsc,
	}
	_, err := cl.FindUsers(req)
	if err == nil {
		t.Error("test for bad json failed")
	}
}

func TestFindUsers(t *testing.T) {
	tCases := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			"bad limit",
			TestFindUsersForBadLimit,
		},
		{
			"bad offset",
			TestFindUsersForBadOffset,
		},
		{
			"bad access token",
			TestFindUsersBadAccessToken,
		},
		{
			"timeout",
			TestFindUsersTimeout,
		},
		{
			"normal request",
			TestFindUsersNormalRequest,
		},
		{
			"internal server error",
			TestFindUsersForInternalServerError,
		},
		{
			"bad order field param",
			TestFindUsersBadRequest,
		},
		{
			"bad json response",
			TestFindUsersForBadJSONResponse,
		},
	}

	for _, curTest := range tCases {
		ok := t.Run(curTest.name, curTest.testFunc)
		if !ok {
			t.Errorf("testing failed on %s", curTest.name)
		}
	}
}
