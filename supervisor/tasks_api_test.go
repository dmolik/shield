package supervisor_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/starkandwayne/shield/supervisor"
	"net/http"

	// sql drivers
	_ "github.com/mattn/go-sqlite3"
)

var _ = Describe("/v1/tasks API", func() {
	var API http.Handler

	// running
	TASK1 := `9211dc07-6c39-4028-997c-fdc4bbf7c5de`
	JOB1 := `07d8475d-0d62-4b59-9f05-8d2173226ad1`

	//
	TASK2 := `b7c35c17-b61b-4541-abea-d93d1837f971`
	ARCHIVE2 := `df4625db-52be-4b55-9ea8-f10214a041bf`
	JOB2 := `0f61e2e1-9293-438f-b2f7-c69a382010ec`

	TASK3 := `524753f0-4f24-4b63-929c-026d20cf07b1`
	JOB3 := `5f04aef7-69cc-40e1-9736-4b3ee4caef50`

	BeforeEach(func() {
		data, err := Database(
			// need a job
			`INSERT INTO jobs (uuid) VALUES ("`+JOB1+`")`,
			`INSERT INTO jobs (uuid) VALUES ("`+JOB2+`")`,
			`INSERT INTO jobs (uuid) VALUES ("`+JOB3+`")`,

			// need an archive
			`INSERT INTO archives (uuid) VALUES ("`+ARCHIVE2+`")`,

			// need a running task
			`INSERT INTO tasks (uuid, owner, op, job_uuid,
				status, started_at, log)
				VALUES (
					"`+TASK1+`", "system", "backup", "`+JOB1+`",
					"running", "2015-04-15 06:00:01", "this is the log"
				)`,

			// need a completed task
			`INSERT INTO tasks (uuid, owner, op, job_uuid, archive_uuid,
				status, started_at, stopped_at, log)
				VALUES (
					"`+TASK2+`", "joe", "restore", "`+JOB2+`", "`+ARCHIVE2+`",
					"done", "2015-04-10 17:35:01", "2015-04-10 18:19:45", "restore complete"
				)`,

			// need a canceled task
			`INSERT INTO tasks (uuid, owner, op, job_uuid,
				status, started_at, stopped_at, log)
				VALUES (
					"`+TASK3+`", "joe", "backup", "`+JOB3+`",
					"canceled", "2015-04-18 19:12:05", "2015-04-18 19:13:55", "cancel!"
				)`,
		)
		Ω(err).ShouldNot(HaveOccurred())
		API = TaskAPI{Data: data}
	})

	It("should retrieve all tasks, sorted properly", func() {
		res := GET(API, "/v1/tasks")
		Ω(res.Code).Should(Equal(200))
		Ω(res.Body.String()).Should(MatchJSON(`[
				{
					"uuid": "` + TASK3 + `",
					"owner": "joe",
					"type": "backup",
					"job_uuid": "` + JOB3 + `",
					"archive_uuid": "",
					"status": "canceled",
					"started_at": "2015-04-18 19:12:05",
					"stopped_at": "2015-04-18 19:13:55",
					"log": "cancel!"
				},
				{
					"uuid": "` + TASK1 + `",
					"owner": "system",
					"type": "backup",
					"job_uuid": "` + JOB1 + `",
					"archive_uuid": "",
					"status": "running",
					"started_at": "2015-04-15 06:00:01",
					"stopped_at": "",
					"log": "this is the log"
				},
				{
					"uuid": "` + TASK2 + `",
					"owner": "joe",
					"type": "restore",
					"job_uuid": "` + JOB2 + `",
					"archive_uuid": "` + ARCHIVE2 + `",
					"status": "done",
					"started_at": "2015-04-10 17:35:01",
					"stopped_at": "2015-04-10 18:19:45",
					"log": "restore complete"
				}
			]`))
		Ω(res.Code).Should(Equal(200))
	})

	It("should retrieve tasks based on status", func() {
		res := GET(API, "/v1/tasks?status=done")
		Ω(res.Code).Should(Equal(200))
		Ω(res.Body.String()).Should(MatchJSON(`[
				{
					"uuid": "` + TASK2 + `",
					"owner": "joe",
					"type": "restore",
					"job_uuid": "` + JOB2 + `",
					"archive_uuid": "` + ARCHIVE2 + `",
					"status": "done",
					"started_at": "2015-04-10 17:35:01",
					"stopped_at": "2015-04-10 18:19:45",
					"log": "restore complete"
				}
			]`))
		Ω(res.Code).Should(Equal(200))
	})

	It("can cancel tasks", func() {
		res := GET(API, "/v1/tasks?status=running")
		Ω(res.Code).Should(Equal(200))
		Ω(res.Body.String()).Should(MatchJSON(`[
				{
					"uuid": "` + TASK1 + `",
					"owner": "system",
					"type": "backup",
					"job_uuid": "` + JOB1 + `",
					"archive_uuid": "",
					"status": "running",
					"started_at": "2015-04-15 06:00:01",
					"stopped_at": "",
					"log": "this is the log"
				}
			]`))

		res = DELETE(API, "/v1/task/"+TASK1)
		Ω(res.Code).Should(Equal(200))
		Ω(res.Body.String()).Should(MatchJSON(`{"ok":"canceled"}`))

		res = GET(API, "/v1/tasks?state=running")
		Ω(res.Code).Should(Equal(200))
	})

	It("ignores other HTTP methods", func() {
		for _, method := range []string{"PUT", "DELETE", "PATCH", "OPTIONS", "TRACE"} {
			NotImplemented(API, method, "/v1/tasks", nil)
		}

		for _, method := range []string{"GET", "HEAD", "POST", "PATCH", "OPTIONS", "TRACE"} {
			NotImplemented(API, method, "/v1/tasks/sub/requests", nil)
			NotImplemented(API, method, "/v1/task/sub/requests", nil)
			NotImplemented(API, method, "/v1/task/5981f34c-ef58-4e3b-a91e-428480c68100", nil)
		}
	})

	It("ignores malformed UUIDs", func() {
		for _, id := range []string{"malformed-uuid-01234", "(abcdef-01234-56-789)"} {
			NotImplemented(API, "PUT", fmt.Sprintf("/v1/task/%s", id), nil)
		}
	})
})