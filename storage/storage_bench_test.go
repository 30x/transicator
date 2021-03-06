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
package storage

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

const (
	benchDBDir  = "./benchdata"
	largeDBDir  = "./benchlargedata"
	cleanDBDir  = "./cleanlargedata"
	parallelism = 100
)

var largeInit = &sync.Once{}
var cleanInit = &sync.Once{}
var largeDB, cleanDB DB
var largeScopes, cleanScopes []string
var largeScopeNames, cleanScopeNames []string
var purgePoint time.Time

func BenchmarkInserts(b *testing.B) {
	db, err := Open(benchDBDir)
	if err != nil {
		b.Fatalf("Error on open: %s\n", err)
	}
	defer func() {
		db.Close()
		db.Delete()
	}()

	scopes, _ := makeScopeList(100, 10000, 1000, b.N)
	b.Logf("Created %d scopes\n", len(scopes))
	b.Logf("Running %d insert iterations\n", b.N)
	b.ResetTimer()
	doInserts(db, scopes, b.N, 1)
}

func BenchmarkBatch10Inserts(b *testing.B) {
	db, err := Open(benchDBDir)
	if err != nil {
		b.Fatalf("Error on open: %s\n", err)
	}
	defer func() {
		db.Close()
		db.Delete()
	}()

	scopes, _ := makeScopeList(100, 10000, 1000, b.N)
	b.Logf("Created %d scopes\n", len(scopes))
	b.Logf("Running %d insert iterations\n", b.N)
	b.ResetTimer()
	doInserts(db, scopes, b.N, 10)
}

func BenchmarkSequence0To100WithMetadata(b *testing.B) {
	largeInit.Do(func() {
		initLargeDB(b)
	})

	if b.N > len(largeScopes) {
		b.Fatalf("Too many iterations: %d\n", b.N)
	}
	b.Logf("Reading %d sequences\n", b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		scope := largeScopeNames[rand.Intn(len(largeScopeNames))]
		entries, _, _, err := largeDB.Scan(
			[]string{scope}, 0, 0, 100, nil)
		if err != nil {
			b.Fatalf("Error on read: %s\n", err)
		}
		if len(entries) == 0 {
			b.Fatal("Expected at least one entry")
		}
	}
}

func BenchmarkSequenceAfterEnd(b *testing.B) {
	largeInit.Do(func() {
		initLargeDB(b)
	})

	if b.N > len(largeScopes) {
		b.Fatalf("Too many iterations: %d\n", b.N)
	}
	b.Logf("Reading %d sequences after end\n", b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		scope := largeScopeNames[rand.Intn(len(largeScopeNames))]
		entries, _, _, err := largeDB.Scan([]string{scope}, uint64(len(largeScopes)+1), 0, 100, nil)
		if err != nil {
			b.Fatalf("Error on read: %s\n", err)
		}
		if len(entries) != 0 {
			b.Fatalf("Expected no entries, got %d\n", len(entries))
		}
	}
}

func BenchmarkSequence0To100WithMetadataParallel(b *testing.B) {
	largeInit.Do(func() {
		initLargeDB(b)
	})
	b.Logf("Reading %d sequences in %d goroutines\n", b.N, parallelism)
	b.SetParallelism(parallelism)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			scope := largeScopeNames[rand.Intn(len(largeScopeNames))]
			entries, _, _, err := largeDB.Scan(
				[]string{scope}, 0, 0, 100, nil)
			if err != nil {
				b.Fatalf("Error on read: %s\n", err)
			}
			if len(entries) == 0 {
				b.Fatal("Expected at least one entry")
			}
		}
	})
}

func BenchmarkSequence0To100WithMetadataAfterClean(b *testing.B) {
	largeInit.Do(func() {
		initLargeDB(b)
	})

	cleanInit.Do(func() {
		purgeRecords(b, cleanDB)
	})

	b.Logf("Reading %d sequences\n", b.N)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		scope := cleanScopeNames[rand.Intn(len(cleanScopeNames))]
		_, _, _, err := cleanDB.Scan(
			[]string{scope}, 0, 0, 100, nil)
		if err != nil {
			b.Fatalf("Error on read: %s\n", err)
		}
	}
}

func purgeRecords(b *testing.B, db DB) {
	b.Logf("Purging about half the records...\n")
	purged, err := db.Purge(purgePoint)
	if err != nil {
		b.Fatalf("Error on purge: %s\n", err)
	}
	b.Logf("Cleaned %d\n", purged)
}

func initLargeDB(b *testing.B) {
	largeScopes, largeScopeNames = makeScopeList(100, 10000, 1000, 1)
	largeDB, _ = initDB(b, largeDBDir, largeScopes)
	cleanScopes, cleanScopeNames = makeScopeList(100, 10000, 1000, 1)
	cleanDB, purgePoint = initDB(b, cleanDBDir, cleanScopes)
}

func initDB(b *testing.B, dir string, insertScopes []string) (DB, time.Time) {
	db, err := Open(dir)
	if err != nil {
		b.Fatalf("Error on open: %s\n", err)
	}

	b.Logf("Inserting %d records\n", len(insertScopes))
	purgePoint := doInserts(db, insertScopes, len(insertScopes), 100)
	return db, purgePoint
}

func doInserts(db DB, scopes []string, iterations, batchSize int) time.Time {
	var seq uint64
	var purgePoint time.Time

	i := 0
	for i < iterations {
		var batch []Entry
		for b := 0; b < batchSize && i < iterations; b++ {
			seq++
			bod := []byte(fmt.Sprintf("seq-%d", seq))
			nb := Entry{
				Scope: scopes[i],
				LSN:   seq,
				Index: 0,
				Data:  bod,
			}
			batch = append(batch, nb)
			if i == (iterations / 2) {
				// Record half way along so that we can purge half the data
				purgePoint = time.Now()
			}
			i++
		}
		err := db.PutBatch(batch)
		if err != nil {
			panic(fmt.Sprintf("Fatal error on batch insert: %s\n", err))
		}
	}
	return purgePoint
}

var _ = Describe("Bench checks", func() {
	It("Permuted scope list", func() {
		sl, _ := makeScopeList(0, 0, 0, 0)
		Expect(sl).Should(BeEmpty())
		sl, sn := makeScopeList(100, 10000, 1000, 200000)
		Expect(len(sl)).Should(BeNumerically(">=", 1000))
		Expect(len(sn)).Should(Equal(100))
	})
})

func makeScopeList(numScopes, stddev, mean, minSize int) ([]string, []string) {
	var rawScopes []string
	var scopeNames []string

	for len(rawScopes) < minSize {
		for sc := 0; sc < numScopes; sc++ {
			scopeName := fmt.Sprintf("Scope-%d", sc)
			scopeNames = append(scopeNames, scopeName)
			rv := math.Abs(rand.NormFloat64()*float64(stddev) + float64(mean))
			count := int(rv)
			for cc := 0; cc < count; cc++ {
				rawScopes = append(rawScopes, scopeName)
			}
		}
	}

	permuted := make([]string, len(rawScopes))
	pix := rand.Perm(len(permuted))

	for i, p := range pix {
		permuted[i] = rawScopes[p]
	}

	return permuted, scopeNames
}
