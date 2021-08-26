package command

import (
	"bulletin/cache"
	"bulletin/feed"
	"bulletin/log"
	"flag"
	"fmt"
	corelog "log"
	"time"
)

const ComposeCommandName = "compose"

type ComposeCommand struct {
	Cache *cache.Cache
}

var referenceTime time.Time

func init() {
	t, err := time.Parse(time.RFC3339, "2000-01-03T00:00:00Z") // monday
	if err != nil {
		corelog.Fatal(err)
	}
	referenceTime = t
}

func (c *ComposeCommand) Execute(args []string) error {
	now := time.Now()
	opts, err := getComposeOptions(args)
	if err != nil {
		return fmt.Errorf("compose: %s", err)
	}
	interval := time.Duration(opts.intervalDays) * 24 * time.Hour
	intervalStart := getNearestInterval(referenceTime, interval, now)
	articles, err := c.Cache.GetArticles()
	if err != nil {
		return err
	}
	var filteredArticles []feed.Article
	for _, a := range articles {
		if a.Updated.After(intervalStart) && !a.Updated.After(intervalStart.Add(interval)) {
			log.Debugf("Accept %s, %s", a.Id, a.Updated)
			filteredArticles = append(filteredArticles, a)
		} else {
			log.Debugf("Drop %s, %s", a.Id, a.Updated)
		}
	}
	formatted, err := feed.FormatHtml(opts.intervalDays, now, filteredArticles)
	if err != nil {
		return err
	}
	fmt.Println(formatted)
	return nil
}

func getComposeOptions(args []string) (composeOptions, error) {
	var options composeOptions
	fs := flag.NewFlagSet(ComposeCommandName, flag.ContinueOnError)
	fs.IntVar(&options.intervalDays, "days", 7, "time range of the articles in DAYS")
	err := fs.Parse(args)
	return options, err
}

type composeOptions struct {
	intervalDays int
}

func getNearestInterval(reference time.Time, interval time.Duration, now time.Time) time.Time {
	n := now.Sub(reference) / interval
	d := (n - 1) * interval
	return reference.Add(d)
}
