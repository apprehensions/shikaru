package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/BurntSushi/toml"
)

type Config struct {
	Token   string   `toml:"token"`
	Filters []struct {
		Pattern string `toml:"pattern"`
		Message string `toml:"message"`
	} `toml:"filter"`
}

type Filter struct {
	Pattern *regexp.Regexp
	Message string
}

func main() {
	configPath := flag.String("config", "shikaru.toml", "config file path")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalln("load config:", err)
	}

	var fs []Filter
	for _, f := range cfg.Filters {
		if f.Message == "" {
			log.Fatalf("filter %s message is empty", f)
		}

		r, err := regexp.Compile(f.Pattern)
		if err != nil {
			log.Fatalf("filter %s compile: %s", f, err)
		}

		fs = append(fs, Filter{
			Pattern: r,
			Message: f.Message,
		})
	}

	s := state.New("Bot " + cfg.Token)
	s.AddIntents(gateway.IntentGuildMessages)
	s.AddHandler(func(c *gateway.MessageCreateEvent) {
		for _, f := range fs {
			if !f.Pattern.MatchString(c.Content) {
				continue
			}

			log.Println(c.Author.Username, "sent", c.Content)

			_, err := s.SendMessageReply(c.ChannelID, f.Message, c.ID)
			if err != nil {
				log.Fatalf("send message reply %s: %s", c.ID, err)
			}
		}
	})

	if err := s.Open(context.Background()); err != nil {
		log.Fatalln("cannot open:", err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	log.Println("Shikaru is now running. Send TERM/INT to exit.")
	<-sc

	s.Close()
}

func loadConfig(name string) (*Config, error) {
	var cfg Config
	
	if _, err := toml.DecodeFile(name, &cfg); err != nil {
		return nil, err
	}

	log.Printf("Config: %+v", cfg)

	return &cfg, nil
}
