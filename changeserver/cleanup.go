/*
Copyright 2016 The Transicator Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"time"

	log "github.com/Sirupsen/logrus"
)

type cleaner struct {
	s        *server
	maxAge   time.Duration
	stopChan chan bool
}

func (s *server) startCleanup(maxAge time.Duration) {
	c := &cleaner{
		s:        s,
		maxAge:   maxAge,
		stopChan: make(chan bool, 1),
	}
	s.cleaner = c
	go c.run()
}

func (c *cleaner) stop() {
	c.stopChan <- true
}

func (c *cleaner) run() {
	tick := time.NewTicker(cleanupDelay(c.maxAge))
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			c.performCleanup()
		case <-c.stopChan:
			return
		}
	}
}

func (c *cleaner) performCleanup() {
	cleanupAge := time.Now().Add(-c.maxAge)
	log.Debugf("Cleaning up data records since before %v", cleanupAge)

	// Get the current last sequence
	_, _, lastSeq, err := c.s.db.Scan(nil, 0, 0, 0, nil)
	if err != nil {
		log.Errorf("Error after preparing for cleanup: %s", err)
		return
	}

	// Always insert a dummy record before cleaning up. That way when the
	// database is empty, we'll still have the last change number in there.
	err = c.s.db.Put(internalScope, lastSeq.LSN, lastSeq.Index, nil)
	if err != nil {
		log.Errorf("Error after preparing for cleanup: %s", err)
		return
	}

	// Now we can do the cleanup knowing that there will still be one record
	// so we can keep track of the highest sequence that we processed.
	cleanupCount, err := c.s.db.Purge(cleanupAge)

	if err != nil {
		log.Errorf("Error after cleaning up %d records: %s", cleanupCount, err)
	} else if cleanupCount > 0 {
		log.Infof("Purged %d old records from the database", cleanupCount)
	}
}

/*
cleanupDelay selects how often to run the cleanup task based
on the duration.
*/
func cleanupDelay(maxAge time.Duration) time.Duration {
	if maxAge <= time.Minute {
		return time.Second
	}
	if maxAge < 5*time.Minute {
		return 5 * time.Second
	}
	if maxAge < time.Hour {
		return 5 * time.Minute
	}
	return time.Hour
}
