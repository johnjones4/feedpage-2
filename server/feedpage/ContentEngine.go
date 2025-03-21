package feedpage

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/mmcdole/gofeed"
)

type post struct {
	title       string
	source      string
	timestamp   time.Time
	description string
	url         string
}

type ContentEngine struct {
	urls  []string
	posts []post
	lock  sync.RWMutex
	Log   *slog.Logger
}

var policy = bluemonday.StrictPolicy()

func (e *ContentEngine) Init() error {
	path := os.Getenv("URLS_PATH")
	if path == "" {
		return errors.New("missing env var URLS_PATH")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &e.urls)
}

func (e *ContentEngine) Start(ctx context.Context) error {
	e.reloadAll()
	tick := time.Tick(time.Minute * 5)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tick:
			e.reloadAll()
		}
	}
}

func (e *ContentEngine) GetPosts() []post {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return e.posts
}

func (e *ContentEngine) reloadAll() {
	e.Log.Info("Reloading all posts")
	all := make([]post, 0)
	for _, url := range e.urls {
		posts, err := e.reload(url)
		if err != nil {
			e.Log.Error("error loading url", slog.String("url", url), slog.Any("error", err))
		}
		if len(posts) > 0 {
			all = append(all, posts...)
		}
	}
	slices.SortFunc(all, func(a post, b post) int {
		return b.timestamp.Compare(a.timestamp)
	})
	if len(all) > 1 {
		for i0 := range len(all) - 1 {
			i := i0 + 1
			if all[i].source == all[i-1].source {
			jloop:
				for j := i + 1; j < len(all); j++ {
					if all[i].source != all[j].source {
						f := all[i].source
						all[i].source = all[j].source
						all[j].source = f
						break jloop
					}
				}
			}
		}
	}
	e.Log.Info("Done reloading", slog.Int("len", len(all)))
	e.lock.Lock()
	e.posts = all
	e.lock.Unlock()
}

func (e *ContentEngine) reload(url string) ([]post, error) {
	e.Log.Info("Loading feed", slog.String("url", url))
	parsed, err := gofeed.NewParser().ParseURL(url)
	if err != nil {
		return nil, err
	}

	limit := time.Now().AddDate(0, -1, 0)
	out := make([]post, 0, len(parsed.Items))
	for _, item := range parsed.Items {
		if item.PublishedParsed.After(limit) {
			desc := cleanStr(item.Content)
			if desc == "" {
				desc = cleanStr(item.Description)
			}
			if len(desc) > 1000 {
				desc = desc[:1000] + " ..."
			}
			out = append(out, post{
				title:       cleanStr(item.Title),
				timestamp:   *item.PublishedParsed,
				source:      cleanStr(parsed.Title),
				description: desc,
				url:         item.Link,
			})
		}
	}
	return out, nil
}

func cleanStr(str string) string {
	return strings.TrimSpace(policy.Sanitize(str))
}
