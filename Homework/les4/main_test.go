package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

//Для пост запроса: curl -F 'file=@testfile.txt' http://localhost:8082/upload
//Вывести список файлов: curl -i http://localhost:8082/
//Вывести список с фильтром по расширению файла: curl -i http://localhost:8082?ext=txt

func TestUploadHandler(t *testing.T) {

	file, err := os.Open("testfile.txt")
	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filepath.Base(file.Name()))
	io.Copy(part, file)
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok!")
	}))
	defer ts.Close()

	uploadHandler := &UploadHandler{
		HostAddr:  ts.URL,
		UploadDir: "upload",
	}
	uploadHandler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `testfile.txt`
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

	// remove testfile
	path := uploadHandler.UploadDir
	if !strings.HasPrefix(path, "/") {
		currentDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		path = filepath.Join(currentDir, path)
	}
	err = os.Remove(filepath.Join(path, "testfile.txt"))
	if err != nil {
		panic(err)
	}

}

func TestBrowseHandler(t *testing.T) {

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dir = filepath.Join(dir, "upload")

	// Create files for test
	f1, err := os.Create(filepath.Join(dir, "testfile.txt"))
	if err != nil {
		panic(err)
	}
	_, err = f1.WriteString("testfile.txt")
	if err != nil {
		panic(err)
	}
	f1.Close()

	f2, err := os.Create(filepath.Join(dir, "file.yml"))
	if err != nil {
		panic(err)
	}
	_, err = f2.WriteString("file.yml")
	if err != nil {
		panic(err)
	}
	f2.Close()

	f3, err := os.Create(filepath.Join(dir, "testfile3"))
	if err != nil {
		panic(err)
	}
	_, err = f3.WriteString("testfile3")
	if err != nil {
		panic(err)
	}
	f3.Close()

	// case 1
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	browseHandler := &BrowseHandler{
		BrowseDir: "upload",
	}
	browseHandler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := []FileInfo{
		{"file.yml", "yml", 8},
		{"testfile.txt", "txt", 12},
		{"testfile3", "", 9},
	}
	data := []FileInfo{}

	err = json.NewDecoder(rr.Body).Decode(&data)
	if err != nil {
		log.Println(err)
	}
	if !reflect.DeepEqual(data, expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			data, expected)
	}

	// case 2
	req, _ = http.NewRequest(http.MethodGet, "?ext=txt", nil)
	rr = httptest.NewRecorder()
	browseHandler.ServeHTTP(rr, req)
	expected = []FileInfo{
		{"testfile.txt", "txt", 12},
	}
	data = []FileInfo{}

	err = json.NewDecoder(rr.Body).Decode(&data)
	if err != nil {
		log.Println(err)
	}
	if !reflect.DeepEqual(data, expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			data, expected)
	}

	// case 3
	req, _ = http.NewRequest(http.MethodGet, "?ext=yml", nil)
	rr = httptest.NewRecorder()
	browseHandler.ServeHTTP(rr, req)
	expected = []FileInfo{
		{"file.yml", "yml", 8},
	}
	data = []FileInfo{}

	err = json.NewDecoder(rr.Body).Decode(&data)
	if err != nil {
		log.Println(err)
	}
	if !reflect.DeepEqual(data, expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			data, expected)
	}

	// remove testfile
	path := browseHandler.BrowseDir
	if !strings.HasPrefix(path, "/") {
		currentDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		path = filepath.Join(currentDir, path)
	}
	err = os.Remove(filepath.Join(path, "testfile.txt"))
	if err != nil {
		panic(err)
	}
	err = os.Remove(filepath.Join(path, "file.yml"))
	if err != nil {
		panic(err)
	}
	err = os.Remove(filepath.Join(path, "testfile3"))
	if err != nil {
		panic(err)
	}

}
