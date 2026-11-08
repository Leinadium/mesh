package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"google.golang.org/protobuf/proto"
)

func tryMapReport(b []byte) bool {
	var data meshtastic.MapReport

	if err := proto.Unmarshal(b, &data); err != nil {
		return false
	}

	println("Map report!")
	println(data.LongName)
	println(data.LatitudeI)
	println(data.LongitudeI)
	return true
}

func tryServiceEnvelope(b []byte) bool {
	var data meshtastic.ServiceEnvelope

	if err := proto.Unmarshal(b, &data); err != nil {
		return false
	}

	if data.Packet != nil {
		decoded := data.Packet.GetDecoded()
		if decoded != nil {
			bytes := decoded.GetPayload()
			if bytes != nil {
				println("---------------")
				fmt.Printf("port num: %v\n", decoded.Portnum.String())

				if decoded.Portnum.String() == meshtastic.PortNum_POSITION_APP.String() {
					var pos meshtastic.Position
					if err := proto.Unmarshal(decoded.Payload, &pos); err != nil {
						fmt.Printf("couldn't decode position")
					} else {
						fmt.Printf("pos: alt=%d lat=%f lon=%f\n", *pos.Altitude, float32(*pos.LatitudeI)*1e-7, float32(*pos.LongitudeI)*1e-7)
					}
				}

				if decoded.Portnum.String() == meshtastic.PortNum_TELEMETRY_APP.String() {
					var tel meshtastic.Telemetry
					if err := proto.Unmarshal(decoded.Payload, &tel); err != nil {
						fmt.Printf("couldn't decode position")
					} else {
						fmt.Printf("telemetry: bat: %d\n", tel.GetDeviceMetrics().BatteryLevel)
					}
				}

				fmt.Printf("bytes: %x\n", bytes)
				fmt.Printf("string: %s\n", string(bytes))
			}
		} else {
			println("payload is nil")
		}
	} else {
		println("decoded is nil")
	}

	return true
}

func handle(msg mqtt.Message) {
	possibles := []func(b []byte) bool{
		tryServiceEnvelope,
		tryMapReport,
	}

	for _, p := range possibles {
		if p(msg.Payload()) {
			break
		}
	}
}

func Handle(client mqtt.Client, msg mqtt.Message) {
	go handle(msg)
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected to MQTT Broker")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection lost: %v\n", err)
}

func main() {
	broker := os.Getenv("BROKER")
	topic := os.Getenv("TOPIC")
	clientID := "go-mqtt"

	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetDefaultPublishHandler(Handle)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// Subscribe to the topic
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic: %s\n", topic)

	// Wait for interrupt signal to gracefully shutdown the subscriber
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	// Unsubscribe and disconnect
	fmt.Println("Unsubscribing and disconnecting...")
	client.Unsubscribe(topic)
	client.Disconnect(250)

	fmt.Println("Goroutine terminated, exiting...")
}
