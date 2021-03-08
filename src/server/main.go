package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type counters struct {
	sync.Mutex
	view  int
	click int
}

//this is a map of 4 counters for each content category. Views and clicks will be stored here until uploaded to the store
var counterMap = map[string]counters{
	"sports":        counters{},
	"entertainment": counters{},
	"business":      counters{},
	"education":     counters{},
}

var content = []string{"sports", "entertainment", "business", "education"}

//this is the counter store
var counterStore = map[string]counters{}

//this is just a variable to lock to modify counters with
var c = counters{}

//global variable to monitor status requests within 10 second window
var statusRequests = 0

//the last request time made to the status page, initialized as 20 seconds ago, so first request will automatically reinitialize this variable
//(becuase isallowed() reinitializes this variable if a request is made and the last request was more than 10 seconds ago)
var lastRefresh = time.Now().Unix() - 20

//function that is executed after every time counters are dumped into the store.
func makeCounters() {
	counterMap = map[string]counters{
		"sports":        counters{},
		"entertainment": counters{},
		"business":      counters{},
		"education":     counters{},
	}
}

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to EQ Works 😎")

}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	data := content[rand.Intn(len(content))]

	//pointer to counter of interest
	c = counterMap[data]

	fmt.Fprintln(w, "You are viewing")
	fmt.Fprintln(w, data)
	c.Lock()
	//increments counter here
	c.view++

	c.Unlock()
	//places new value of counter back into map
	counterMap[data] = c

	err := processRequest(r)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(400)
		return
	}

	// simulate random click call
	if rand.Intn(100) < 50 {
		processClick(data)

	}
}

func processRequest(r *http.Request) error {
	time.Sleep(time.Duration(rand.Int31n(50)) * time.Millisecond)
	return nil
}

func processClick(data string) error {
	//same logic as view, put counter in pointer, increment, put back.
	var c = counterMap[data]
	c.Lock()
	c.click++
	c.Unlock()
	counterMap[data] = c
	return nil
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	if !isAllowed() {
		w.WriteHeader(429)
		return
	}
}

func isAllowed() bool {
	// time windows are 10 seconds long, you are allowed 10 requests per window. every 10 seconds, a new window is opened. If your last request was more than 10 seconds ago, a new window starts
	// the goal is to make it so that spamming requests 10 times in a row blocks the request.
	if lastRefresh < (time.Now().Unix() - 10) {
		lastRefresh = time.Now().Unix()
		statusRequests = 1
		return true
	}
	if statusRequests < 10 {
		statusRequests++
		return true
	}
	return false
}

func uploadCounters() error {
	//here, every 5 seconds, all contents of the counter map are moved to the counter store
	//the counter store will store all data about views and clicks in each category for each 5 second interval.
	for true {
		time.Sleep(5000 * time.Millisecond)
		var i = 0
		for i < 4 {
			counterStore[content[i]+time.Now().String()] = counterMap[content[i]]
			i++
		}
		makeCounters()

	}

	return nil
}

func main() {
	go uploadCounters()
	http.HandleFunc("/", welcomeHandler)
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/stats/", statsHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
