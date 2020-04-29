package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
)

const (
	expTime       = 60
	longURL       = "https://www.baidu.com"
	shortlink     = "IFHzaO"
	shortlinkInfo = `{"url": "https://www.baidu.com", "expiration_in_minutes": 1}`
)

type storageMock struct {
	mock.Mock
}

var app App
var mockR *storageMock

func (s *storageMock) Shorten(url string, exp int64) (string, error) {
	args := s.Called(url, exp)
	return args.String(0), args.Error(1)
}

func (s *storageMock) ShortlinkInfo(eid string) (interface{}, error) {
	args := s.Called(eid)
	return args.String(0), args.Error(1)
}

func (s *storageMock) Unshorten(eid string) (string, error) {
	args := s.Called(eid)
	return args.String(0), args.Error(1)
}

func init() {
	app = App{}
	mockR = new(storageMock)
	app.Initialize(&Env{S:mockR})
}

func TestCreateShortlink(t *testing.T) {
	var jsonStr = []byte(`{
		"url":"https://www.baidu.com",
		"expiration_in_minutes": 60}`)
	req, err := http.NewRequest("POST", "/api/shorten", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal("Should able to create request.", err)
	}
	req.Header.Set("Content-Type","application/json")
	mockR.On("Shorten",longURL,int64(expTime)).Return(shortlink,nil).Once()

	rw := httptest.NewRecorder()
	app.Router.ServeHTTP(rw, req)
	if rw.Code != http.StatusCreated {
		t.Fatalf("Excepted recive %d, get %d", http.StatusCreated, rw.Code)
	}

	resp := struct {
		shortlink string `json:"shortlink"`
	}{}
	if err := json.NewDecoder(rw.Body).Decode(&resp); err != nil {
		t.Fatal("Should decode the response")
	}

	if resp.shortlink != shortlink {
		t.Fatalf("Except receive %s. Got %s", shortlink, resp.shortlink)
	}
}

func TestRedirect(t *testing.T) {
	r := fmt.Sprintf("/%s", shortlink)
	req, err := http.NewRequest("GET", r,nil)
	if err != nil {
		t.Fatalf("Should be able to create a request")
	}
	mockR.On("Unshorten", shortlink).Return(longURL,nil).Once()
	rw := httptest.NewRecorder()
	app.Router.ServeHTTP(rw,req)

	if rw.Code != http.StatusTemporaryRedirect {
		t.Fatalf("Except receive %d. Got %d", http.StatusTemporaryRedirect, rw.Code)
	}
}
