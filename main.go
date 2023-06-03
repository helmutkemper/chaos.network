package main

import (
	"encoding/hex"
	"fmt"
	"github.com/helmutkemper/chaos/networkdelay"
	"log"
	"math/rand"
	"os"
	"strconv"
)

//var localAddr = flag.String("l", ":27016", "local address")
//var remoteAddr = flag.String("r", "10.0.0.2:27017", "remote address")

type ParserFunc struct{}

func (e ParserFunc) Parser(data []byte, direction string) (dataSize int, err error) {
	if len(data) == 0 {
		fmt.Printf("direction: %v\nempty\n\n", direction)
		return
	}

	fmt.Printf("direction: %v\n%v\n", direction, hex.Dump(data))
	dataSize = len(data)
	return
}

type ParserChaosFunc struct{}

func (e ParserChaosFunc) Parser(data []byte, direction string) (dataSize int, err error) {
	if len(data) == 0 {
		fmt.Printf("direction: %v\nempty\n\n", direction)
		return
	}

	if changeRate != -1.0 && rand.Float64() <= changeRate {
		addr := rand.Intn(len(data) - 1)
		newValue := byte(rand.Intn(255))
		data[addr] = newValue
	}

	fmt.Printf("direction: %v\n%v\n", direction, hex.Dump(data))
	dataSize = len(data)
	return
}

var changeRate float64

func main() {
	var err error

	var bufferSize int64
	var minDelay, maxDelay int64

	var bufferSizeString = os.Getenv("CHAOS_NETWORK_BUFFER_SIZE")
	var minDelayString = os.Getenv("CHAOS_NETWORK_MIN_DELAY")
	var maxDelayString = os.Getenv("CHAOS_NETWORK_MAX_DELAY")
	var changeRateString = os.Getenv("CHAOS_NETWORK_DATA_CHANGE_RATE")

	var localPortString = os.Getenv("CHAOS_NETWORK_LOCAL_PORT")
	var remoteContainerString = os.Getenv("CHAOS_NETWORK_REMOTE_CONTAINER")

	if bufferSize, err = strconv.ParseInt(bufferSizeString, 10, 64); err != nil {
		bufferSize = 32 * 1024
	}

	if minDelay, err = strconv.ParseInt(minDelayString, 10, 64); err != nil {
		minDelay = 100
	}

	if maxDelay, err = strconv.ParseInt(maxDelayString, 10, 64); err != nil {
		maxDelay = 600
	}

	if changeRate, err = strconv.ParseFloat(changeRateString, 64); err != nil {
		// interval is between 0.0 and 1.0. So 100.0 will never happen
		changeRate = -1.0
	}

	if changeRate == 0 {
		changeRate = -1.0
		log.Println("CHAOS_NETWORK_DATA_CHANGE_RATE: zero is't allowed")
	}

	if localPortString == "" {
		log.Println("CHAOS_NETWORK_LOCAL_PORT: must be set. eg.\":27016\"")
		os.Exit(1)
	}

	if remoteContainerString == "" {
		log.Println("CHAOS_NETWORK_REMOTE_CONTAINER: must be set. eg.\"delete_mongo_0:27017\"")
		os.Exit(1)
	}

	log.Printf("CHAOS_NETWORK_BUFFER_SIZE: %v", bufferSize)
	log.Printf("CHAOS_NETWORK_MIN_DELAY: %v milliseconds", minDelay)
	log.Printf("CHAOS_NETWORK_MAX_DELAY: %v milliseconds", maxDelay)
	log.Printf("CHAOS_NETWORK_DATA_CHANGE_RATE: %v", changeRate)
	log.Printf("CHAOS_NETWORK_LOCAL_PORT: \"%v\"", localPortString)
	log.Printf("CHAOS_NETWORK_REMOTE_CONTAINER: \"%v\"", remoteContainerString)

	fmt.Printf("\nListening: %v\nProxying: %v\n\n", localPortString, remoteContainerString)

	var p networkdelay.ParserInterface = &ParserFunc{}

	var proxy networkdelay.Proxy
	proxy.SetBufferSize(int(bufferSize))
	proxy.SetParserFunction(p)
	proxy.SetDelayMillisecond(int(minDelay), int(maxDelay))

	err = proxy.Proxy(localPortString, remoteContainerString)
	if err != nil {
		panic(err)
	}
}
