package purge

import (
	"os"
	"os/signal"
	"purge-manager/config"
	"purge-manager/utils"
	"sync"
	"syscall"
	"time"

	"github.com/gorhill/cronexpr"
	log "github.com/sirupsen/logrus"
)

func StartPurge(softDelete bool) {
	for _, pc := range config.Purges() {
		p := NewPurgeSet(pc, softDelete)
		tt := utils.StartTimeTracker()
		log.Info("purgeset '", p.Name, "' : start.")
		p.Start()
		log.Info("purgeset '", p.Name, "' : done in ", tt.ElapsedHuman(), ".")
	}
}

func StartPurgeCron(softDelete bool) {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	var wg sync.WaitGroup
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for _, pc := range config.Purges() {
		wg.Add(1)
		go func() {
			p := NewPurgeSet(pc, softDelete)
			cron := pc.Cron
			for {
				expr := cronexpr.MustParse(cron)
				nextTime := expr.Next(time.Now())
				log.Info("purgeset '", p.Name, "' : Next execution at ", nextTime)
				duration := time.Until(nextTime)
				timer := time.NewTimer(duration)
				select {
				case <-timer.C:
					tt := utils.StartTimeTracker()
					log.Info("purgeset '", p.Name, "' : start.")
					p.Start()
					log.Info("purgeset '", p.Name, "' : done in ", tt.ElapsedHuman(), ".")
				case <-done:
					wg.Done()
					return
				}
			}
		}()
	}
	<-sigs
	close(done)
	wg.Wait()
}
