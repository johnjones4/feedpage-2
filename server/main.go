package main

import (
	"context"
	"log/slog"
	"main/feedpage"
	"sync"
)

func main() {
	log := slog.Default()

	ce := &feedpage.ContentEngine{
		Log: log,
	}

	api := &feedpage.API{
		ContentEngine: ce,
		Log:           log,
	}

	err := ce.Init()
	if err != nil {
		log.Error("error initializing content engine", slog.Any("error", err))
		return
	}

	startables := []feedpage.Startable{
		ce,
		api,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	for _, s := range startables {
		wg.Add(1)
		go func() {
			err := s.Start(ctx)
			if err != nil {
				log.Error("error starting service", slog.Any("error", err))
				cancel()
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
