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
package snapshotserver

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Schema tests", func() {
	It("Verify schema table pickup that have _change_selector column", func() {
		tx, err := db.Begin()
		Expect(err).Should(Succeed())
		defer tx.Commit()

		s, err := enumeratePgTables(tx)
		Expect(err).Should(Succeed())
		Expect(s["public.snapshot_test"]).ShouldNot(BeNil())
		Expect(s["not.found"]).Should(BeNil())

		st := s["public.snapshot_test"]
		Expect(st.schema).Should(Equal("public"))
		Expect(st.name).Should(Equal("snapshot_test"))
		Expect(st.columns[0].name).Should(Equal("id"))
		Expect(st.columns[1].name).Should(Equal("bool"))
		Expect(st.columns[2].name).Should(Equal("chars"))
		Expect(st.columns[3].name).Should(Equal("varchars"))
		Expect(st.columns[4].name).Should(Equal("int"))
		Expect(st.columns[5].name).Should(Equal("smallint"))
		Expect(st.columns[6].name).Should(Equal("bigint"))
		Expect(st.columns[7].name).Should(Equal("float"))
		Expect(st.columns[8].name).Should(Equal("double"))
		Expect(st.columns[9].name).Should(Equal("date"))
		Expect(st.columns[10].name).Should(Equal("time"))
		Expect(st.columns[11].name).Should(Equal("blob"))
		Expect(st.columns[12].name).Should(Equal("_change_selector"))
		Expect(st.columns[0].typid).Should(Equal(1043))
		Expect(st.columns[0].primaryKey).Should(BeTrue())
		Expect(st.columns[14].name).Should(Equal("timestampp"))
		Expect(st.columns[14].typid).Should(Equal(1114))
		Expect(st.columns[14].primaryKey).Should(BeFalse())
		Expect(st.primaryKeys[0]).Should(Equal("id"))
		Expect(st.hasSelector).Should(BeTrue())
		fmt.Fprintf(GinkgoWriter, "SQL: %s\n", makeSqliteTableSQL(st))

		a := s["public.app"]
		Expect(a).ShouldNot(BeNil())
		Expect(a.columns[0].name).Should(Equal("org"))
		Expect(a.columns[1].name).Should(Equal("id"))
		Expect(a.columns[2].name).Should(Equal("dev_id"))
		Expect(a.columns[3].name).Should(Equal("display_name"))
		Expect(a.columns[4].name).Should(Equal("name"))
		Expect(a.columns[5].name).Should(Equal("created_at"))
		Expect(a.columns[6].name).Should(Equal("created_by"))
		Expect(a.columns[7].name).Should(Equal("_change_selector"))
		Expect(a.columns[0].primaryKey).Should(BeFalse())
		Expect(a.columns[7].primaryKey).Should(BeTrue())
		Expect(a.primaryKeys[0]).Should(Equal("id"))
		Expect(a.primaryKeys[1]).Should(Equal("_change_selector"))
		fmt.Fprintf(GinkgoWriter, "SQL: %s\n", makeSqliteTableSQL(a))

		a = s["public.developer"]
		Expect(a).ShouldNot(BeNil())
		Expect(a.columns[0].name).Should(Equal("org"))
		Expect(a.columns[1].name).Should(Equal("id"))
		Expect(a.columns[2].name).Should(Equal("username"))
		Expect(a.columns[3].name).Should(Equal("firstname"))
		Expect(a.columns[4].name).Should(Equal("created_at"))
		Expect(a.columns[5].name).Should(Equal("created_by"))
		Expect(a.columns[6].name).Should(Equal("_change_selector"))
		Expect(a.columns[0].typid).Should(Equal(1043))
		Expect(a.columns[1].primaryKey).Should(BeTrue())
		Expect(a.columns[6].primaryKey).Should(BeTrue())
		fmt.Fprintf(GinkgoWriter, "SQL: %s\n", makeSqliteTableSQL(a))

		a = s["transicator_tests.schema_table"]
		Expect(a).ShouldNot(BeNil())
		Expect(a.columns[0].name).Should(Equal("id"))
		Expect(a.columns[1].name).Should(Equal("created_at"))
		Expect(a.columns[2].name).Should(Equal("_change_selector"))
		Expect(a.columns[0].primaryKey).Should(BeTrue())
		Expect(a.columns[0].typid).Should(Equal(1043))

		a = s["public.deployment_history"]
		Expect(a).Should(BeNil())

		a = s["public.deployment_history2"]
		Expect(a).Should(BeNil())

		a = s["public.DATA_SCOPE"]
		Expect(a).Should(BeNil())

		a = s["public.APID_CLUSTER"]
		Expect(a).Should(BeNil())

	})

	It("Check timestamp format", func() {
		// Just make sure that timestamp formatting works the way that we expect
		now := time.Now().UTC()
		fmt.Fprintf(GinkgoWriter, "Time is now %s\n", now)

		nowStr := now.Format(sqliteTimestampFormat)
		fmt.Fprintf(GinkgoWriter, "          = %s\n", nowStr)

		nowParsed, err := time.Parse(sqliteTimestampFormat, nowStr)

		Expect(err).Should(Succeed())
		Expect(nowParsed.UnixNano()).Should(Equal(now.UnixNano()))
	})
})
