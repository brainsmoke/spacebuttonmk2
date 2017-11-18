package main

import (
	"time"
	"os"
	"strings"
	"net/http"
	"flag"
	"post6.net/spacestate/gpio"
)

var secretKey string

func init() {
	flag.StringVar(&secretKey, "key", "s3cr1t", "state change request key")
}

const statusLedGPIO = 21
const buttonGPIO = 20

const (
	Green = iota
	Red
)

type SpaceState int

const (
	Unknown SpaceState = iota
	Closed
	Open
)

var strState = []string { "unknown", "closed", "open" }

var blinkPattern = [3][]int {
	{Red,Green},
	{Red},
	{Green},
}

func showStatus(gpioPin int, stateChange chan SpaceState, quit chan int) {

	ledGPIO, err := gpio.OpenGPIO(gpioPin, gpio.Output)
	if err != nil {
		panic(err)
	}
	defer ledGPIO.Close();
	state := Unknown
	color := -1
	tick := time.NewTicker(time.Second / time.Duration(4))
	i := 0
	for {
		if color != blinkPattern[state][i] {
			color = blinkPattern[state][i]
			gpio.WriteGPIO(ledGPIO, color)
		}
		select {
			case <- tick.C:
				i += 1
				i %= len(blinkPattern[state])
			case state = <-stateChange:
				i %= len(blinkPattern[state])
			case <-quit:
				return
		}
	}
}

func buttonPoll(gpioPin int, buttonPress, quit chan int) {

	buttonGPIO, err := gpio.OpenGPIO(gpioPin, gpio.Input)
	if err != nil {
		panic(err)
	}
	defer buttonGPIO.Close();
	tick := time.NewTicker(time.Second / time.Duration(20))
	i := 0
	oldP := 1
	for {
		select {
			case <- tick.C:
			case <- quit:
				return
		}
		if i == 0 {
			p := gpio.ReadGPIO(buttonGPIO)
			if p == 0 && oldP == 1 {
				buttonPress <- 1
			}
			oldP = p
			i = 4
		} else {
			i -= 1
		}
	}
}

func getState(url string) SpaceState {
	os.Stdout.Write([]byte("doing request: "+url+"\n"))
	buf := make([]byte, 1024)
	resp, err := http.Get(url)
	if err != nil {

		return Unknown
	}
	_, err = resp.Body.Read(buf)
	s := strings.ToLower(string(buf))
	if strings.Contains(s, "open") {
		return Open
	} else if strings.Contains(s, "closed") {
		return Closed
	}
	return Unknown
}

func changeState(state SpaceState, stateChange chan SpaceState) {
	stateChange <- getState("https://techinc.nl/space/index.php?state="+strState[state]+"&key="+secretKey)
}

func fetchState(stateChange chan SpaceState, quit chan int) {
	tick := time.NewTicker(time.Second * time.Duration(30))
	for {
		state := getState("https://techinc.nl/space/spacestate")
		if state != Unknown {
			stateChange <- state
		}
		select {
			case <-tick.C:
			case <-quit:
				return
		}
	}
}

func main() {
	flag.Parse()
	stateChangeDisplay := make(chan SpaceState)
	stateChangeIn := make(chan SpaceState)
	buttonPress := make(chan int)
	quit := make(chan int)
	go showStatus(statusLedGPIO, stateChangeDisplay, quit)
	go buttonPoll(buttonGPIO, buttonPress, quit)
	go fetchState(stateChangeIn, quit)
	state := Unknown
	for {
		select {
			case state = <-stateChangeIn:
				stateChangeDisplay<- state
				os.Stdout.Write([]byte("state change: "+strState[state]+"\n"))
			case <-buttonPress:
				os.Stdout.Write([]byte("button pressed\n"))
				newState := Open
				if state == Open {
					newState = Closed
				}
				go changeState(newState, stateChangeIn)
		}
	}
}
