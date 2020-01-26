package main

import (
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/morus12/dht22"
)

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("TOPIC: %s\n", msg.Topic())
	log.Printf("MSG: %s\n", msg.Payload())
}

func main() {
	// mqtt.DEBUG = log.New(os.Stdout, "", 0)
	// mqtt.ERROR = log.New(os.Stdout, "", 0)
	opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883").SetClientID("pidht")
	opts.SetKeepAlive(2 * time.Second)
	opts.SetDefaultPublishHandler(f)
	opts.SetPingTimeout(1 * time.Second)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	log.Println("Connected to mqtt")

	sensor := dht22.New("GPIO_26")

	c1 := make(chan struct{}, 1)
	go func() {
		temperature, err := sensor.Temperature()
		if err != nil {
			log.Fatal(err)
		}
		log.Println(temperature - 1)
		humidity, err := sensor.Humidity()
		if err != nil {
			log.Fatal(err)
		}
		log.Println(humidity)
		if temperature < 0 || humidity < 0 {
			os.Exit(0)
		}
		token := c.Publish("sensors/node_bedroom/last", 0, true, fmt.Sprintf(`{"temperature": %0.2f,"humidity": %0.2f}`, temperature-1, humidity))
		token.Wait()
		c1 <- struct{}{}
	}()

	select {
	case <-c1:
		log.Println("done")
	case <-time.After(15 * time.Second):
		log.Println("timeout...")
		os.Exit(1)
	}

}
