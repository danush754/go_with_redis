package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

var client *redis.Client

func init() {
	fmt.Println("Trying to connect to the redis server")

	client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	err := client.Ping(context.Background()).Err()
	if err != nil {
		log.Fatalf("failed top connect to the redis server : %v", err)
	}
}

func main() {

	engine := html.New("./view", ".html")

	newApp := fiber.New(fiber.Config{
		Views: engine,
	})

	newApp.Static("/", "./public")

	newApp.Get("/", LoadStartingPage)

	newApp.Post("/get-url/:url", ShortenOriginalUrl)

	newApp.Listen(":8000")

}

func LoadStartingPage(ctx *fiber.Ctx) error {

	fmt.Println("loaded home page")

	return ctx.Render("homepage", fiber.Map{})
}

func ShortenOriginalUrl(ctx *fiber.Ctx) error {

	originalUrl := ctx.Params("url")

	if originalUrl == "" {
		fmt.Println("")
	}

	return nil
}
