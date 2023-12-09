package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

const (
	rabbitMQHost     = "localhost:5672"
	rabbitMQExchange = "tags"
	rabbitMQQueue    = "tags"
)

func main() {
	conn, err := amqp.Dial("amqp://" + rabbitMQHost)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		rabbitMQExchange,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Declare and bind the queue
	_, err = ch.QueueDeclare(
		rabbitMQQueue,
		true,
		false,
		false,
		false,
		nil)
	if err != nil {
		log.Fatal(err)
	}

	err = ch.QueueBind(
		rabbitMQQueue,
		"",
		rabbitMQExchange,
		false,
		nil)
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()

	r.GET("/metrics", func(c *gin.Context) {
		h := promhttp.Handler()
		h.ServeHTTP(c.Writer, c.Request)
	})
	r.POST("/tags", func(c *gin.Context) {
		var tags []map[string]interface{}
		err := c.BindJSON(&tags)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		for _, tag := range tags {
			message, err := json.Marshal(tag)
			if err != nil {
				log.Printf("Failed to marshal tag: %v", err)
				continue
			}

			err = ch.Publish(
				rabbitMQExchange,
				"",
				false,
				false,
				amqp.Publishing{
					ContentType: "application/json",
					Body:        message,
				},
			)
			if err != nil {
				log.Printf("Failed to publish message to RabbitMQ: %v", err)
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Tags successfully sent to RabbitMQ"})
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	log.Printf("Server started on port %d", server.Addr)
	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
}
