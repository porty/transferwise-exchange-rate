package main

import (
	"context"

	"github.com/joho/godotenv"
	"github.com/porty/transferwise-exchange-rate/p"
)

func main() {
	_ = godotenv.Load(".env")

	if err := p.HelloPubSub(context.Background(), p.PubSubMessage{
		Data: []byte("hello"),
	}); err != nil {
		panic(err)
	}
}
