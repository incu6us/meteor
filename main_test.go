package main

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "fmt"
)

func TestRunFunc(t *testing.T)  {
    req, err := http.NewRequest("GET", "/api/task/run/test", nil)
    if err != nil {
        t.Fatal(err)
    }

    recorder := httptest.NewRecorder()
    handler := http.HandlerFunc(RunFunc)
    handler.ServeHTTP(recorder, req)

    fmt.Println(recorder.Body.String())
    //defer req.Body.Close()

    //data, err := ioutil.ReadAll(req.Body)
    //if err != nil {
    //    t.Error(err)
    //}

    //fmt.Print(string(data))
}
