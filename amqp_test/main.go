package main

import (
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
	log "gitlab.com/orgname/go-logger"
	drv_cons "gitlab.com/orgname/go_consumers/drivers"
	producers "gitlab.com/orgname/go_producers"
	drv_prod "gitlab.com/orgname/go_producers/drivers"
)

const (
	Topic    = "blocks"
	FilePath = "app.log"
)

var cRcv int

// var prodChan chan bool
// var consChan chan bool

func main() {
	logger := logrus.New()

	filePath := os.Args[1]

	file, err := os.Open(filePath)
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer file.Close()

	amqpURI := "amqp://guest:guest@localhost:5672"

	cons := drv_cons.NewAmqpConsumer(amqpURI, logger)
	defer cons.Close()

	cTx := 0
	cRcv = 0
	defer func() {
		fmt.Println(cTx)
		fmt.Println(cRcv)
	}()

	producer, err := drv_prod.NewAmqpProducer(amqpURI, logger)
	if err != nil {
		logger.Fatal("Unable connect to RabbitMQ. ", err.Error())
	}

	bytesChan := make(chan []byte)
	// prodChan = make(chan bool)
	// consChan = make(chan bool)

	go func() {
		readFile(file, bytesChan)
		close(bytesChan)
	}()

	go func() {
		err := cons.Consume(Topic, "amqp_test", handler, logger)
		if err != nil {
			logger.Fatal(err.Error())
		}
		// consChan <- true
	}()

	for bytes := range bytesChan {
		cTx++
		// line := string(bytes)

		// fmt.Println(line)
		// logger.Fatal("STOP")

		// task, err := producers.NewProducerTask(line)
		task, err := producers.NewProducerTask(bytes)
		if err != nil {
			logger.Fatal(err.Error())
		}
		_, err = producer.Produce(Topic, task)
		if err != nil {
			logger.Fatal(err.Error())
		}
		// if cTx == 10 {
		// 	break
		// }
	}

	// <-prodChan
	// <-consChan
}

func readFile(reader io.Reader, bytesChan chan []byte) {
	bytes := make([]byte, 1024)
	// bytes := make([]byte, 20)
	for {
		_, err := reader.Read(bytes)
		if err != nil {
			break
		}
		// fmt.Printf("%v\n", bytes)

		bytesChan <- (bytes)
	}
	// prodChan <- true
}

func handler(body []byte, msgID string, logger log.Logger) error {
	// fmt.Print(string(body))
	cRcv++
	return nil
}
