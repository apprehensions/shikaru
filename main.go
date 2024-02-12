package main

import (
	"context"
	"flag"
	"strings"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gopkg.in/yaml.v3"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

type Filter struct {
	Match []string `yaml:"match"`
	Send  string   `yaml:"send"`
}

type Config struct {
	Token   string   `yaml:"token"`
	Filters []Filter `yaml:"filters"`
}

func main() {
	configPath := flag.String("config", "shikaru.yaml", "config file path")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalln("load config:", err)
	}

	s := state.New("Bot " + cfg.Token)
	s.AddIntents(gateway.IntentGuildMessages)
	s.AddHandler(func(c *gateway.MessageCreateEvent) {
		for _, f := range cfg.Filters {
			for _, m := range f.Match {
				if !strings.Contains(c.Content, m) {
					return
				}
			}

			log.Println(c.Author.Username, "sent", c.Content)

			_, err := s.SendMessageReply(c.ChannelID, f.Send, c.ID)
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
	f, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(f, &cfg); err != nil {
		return nil, err
	}

	log.Printf("Config: %+v", cfg)

	return &cfg, nil
}
