package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
	"time"
    "os/exec"
)

func TestGetUrls(t *testing.T) {
    go build()
    time.Sleep(3*time.Second)
    go run()
    time.Sleep(3*time.Second)
	sum := 0
	for i := 0; i < 100; i++ {
		time.Sleep(21 * time.Millisecond)
		go get(t, "http://localhost:9100/en/")
        time.Sleep(21 * time.Millisecond)
		go get(t, "http://localhost:9100/fi/")
        time.Sleep(21 * time.Millisecond)
		go get(t, "http://localhost:9100/fi/category/Viihde")
        time.Sleep(21 * time.Millisecond)
		go get(t, "http://localhost:9100/fi/category/Talous")
        time.Sleep(21 * time.Millisecond)
		go get(t, "http://localhost:9100/fi/category/Ulkomaat")
        time.Sleep(21 * time.Millisecond)
        go get(t, "http://localhost:9100/fi/category/Digi")
        time.Sleep(21 * time.Millisecond)
        go get(t, "http://localhost:9100/en/category/Viihde")
        time.Sleep(21 * time.Millisecond)
		go get(t, "http://localhost:9100/en/category/Talous")
        time.Sleep(21 * time.Millisecond)
		go get(t, "http://localhost:9100/en/category/Ulkomaat")
        time.Sleep(21 * time.Millisecond)
        go get(t, "http://localhost:9100/en/category/Digi")
        time.Sleep(21 * time.Millisecond)
        go get(t, "http://localhost:9100/uutispuro_logo_small.gif")
        time.Sleep(21 * time.Millisecond)
        go get(t, "http://localhost:9100/like.png")
		sum += 12
	}
    exec.Command("killall uutispuroReader")
	fmt.Printf("Made %v GETs ", sum)
}

func build() {
    out, _ := exec.Command("/usr/local/bin/go","build").Output()
    fmt.Printf("%s", out)
    o, _ := exec.Command("ls","-la", "uutispuroReader").Output()
    fmt.Printf("%s", o)
}

func run() {
    o, _ := exec.Command("sh", "-c", "./uutispuroReader").Output()
    fmt.Printf("%s", o)
}

func get(t *testing.T, url string) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	res, err := client.Do(req)
    if err != nil {
        fmt.Println("Bad Response " + err.Error())
    } else {
    	data, err2 := ioutil.ReadAll(res.Body)
    	if err2 != nil {
    		fmt.Println(err2)
    	}
    	res.Body.Close()
    	body := string(data)
    	fmt.Println(url + " " + strconv.Itoa(res.StatusCode))
    	if res.StatusCode != http.StatusOK {
    		t.Fatalf("Non-expected status code: %v\n\tbody: %v, data:%s\n", http.StatusCreated, res.StatusCode, body)
    	}
    }
}