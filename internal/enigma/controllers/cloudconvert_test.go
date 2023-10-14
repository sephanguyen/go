package controllers

import (
	"encoding/json"
	"testing"
)

func TestParseCloudConvertWebhookPayload(t *testing.T) {
	payload := `
{"event":"job.finished","job":{"id":"9599ced2-e526-410f-9277-f713c082b85a","tag":null,"status":"finished","created_at":"2021-06-29T07:37:09+00:00","started_at":"2021-06-29T07:37:09+00:00","ended_at":"2021-06-29T07:37:16+00:00","tasks":[{"id":"60120f19-af7e-4869-927f-bc2d6983f173","name":"export-01F9BB6JY1B2DQC9K5E83P5908","job_id":"9599ced2-e526-410f-9277-f713c082b85a","status":"finished","credits":0,"code":null,"message":null,"percent":100,"operation":"export\/google-cloud-storage","result":{"files":[{"filename":"01EM0K92PVQ1R5SF4 20JM33RQ70K-1.png","dir":"assignment\/2021-06-29\/images\/01F9BB6JY1B2DQC9K5E83P5908\/"},{"filename":"01EM0K92PVQ1R5SF4 20JM33RQ70K-2.png","dir":"assignment\/2021-06-29\/images\/01F9BB6JY1B2DQC9K5E83P5908\/"},{"filename":"01EM0K92PVQ1R5SF4 20JM33RQ70K-3.png","dir":"assignment\/2021-06-29\/images\/01F9BB6JY1B2DQC9K5E83P5908\/"},{"filename":"01EM0K92PVQ1R5SF4 20JM33RQ70K-4.png","dir":"assignment\/2021-06-29\/images\/01F9BB6JY1B2DQC9K5E83P5908\/"},{"filename":"01EM0K92PVQ1R5SF4 20JM33RQ70K-5.png","dir":"assignment\/2021-06-29\/images\/01F9BB6JY1B2DQC9K5E83P5908\/"}]},"created_at":"2021-06-29T07:37:09+00:00","started_at":"2021-06-29T07:37:13+00:00","ended_at":"2021-06-29T07:37:16+00:00","retry_of_task_id":null,"copy_of_task_id":null,"user_id":47853587,"priority":10,"host_name":"kirstin","storage":null,"depends_on_task_ids":["f168ba3f-6a37-40f4-a900-48cb31516fc3"],"links":{"self":"https:\/\/api.cloudconvert.com\/v2\/tasks\/60120f19-af7e-4869-927f-bc2d6983f173"}},{"id":"f168ba3f-6a37-40f4-a900-48cb31516fc3","name":"convert-01F9BB6JY1B2DQC9K5E83P5908","job_id":"9599ced2-e526-410f-9277-f713c082b85a","status":"finished","credits":1,"code":null,"message":null,"percent":100,"operation":"convert","engine":"mupdf","engine_version":"1.17.0","result":{"files":[{"filename":"01EM0K92PVQ1R5SF4 20JM33RQ70K-1.png","size":210396},{"filename":"01EM0K92PVQ1R5SF4 20JM33RQ70K-2.png","size":314827},{"filename":"01EM0K92PVQ1R5SF4 20JM33RQ70K-3.png","size":266105},{"filename":"01EM0K92PVQ1R5SF4 20JM33RQ70K-4.png","size":226436},{"filename":"01EM0K92PVQ1R5SF4 20JM33RQ70K-5.png","size":176137}]},"created_at":"2021-06-29T07:37:09+00:00","started_at":"2021-06-29T07:37:10+00:00","ended_at":"2021-06-29T07:37:13+00:00","retry_of_task_id":null,"copy_of_task_id":null,"user_id":47853587,"priority":10,"host_name":"kirstin","storage":null,"depends_on_task_ids":["c277da79-c43f-411c-9cfe-65146f5d13d2"],"links":{"self":"https:\/\/api.cloudconvert.com\/v2\/tasks\/f168ba3f-6a37-40f4-a900-48cb31516fc3"}},{"id":"c277da79-c43f-411c-9cfe-65146f5d13d2","name":"import-01F9BB6JY1B2DQC9K5E83P5908","job_id":"9599ced2-e526-410f-9277-f713c082b85a","status":"finished","credits":0,"code":null,"message":null,"percent":100,"operation":"import\/url","result":{"files":[{"filename":"01EM0K92PVQ1R5SF4 20JM33RQ70K.pdf","size":363039}]},"created_at":"2021-06-29T07:37:09+00:00","started_at":"2021-06-29T07:37:09+00:00","ended_at":"2021-06-29T07:37:10+00:00","retry_of_task_id":null,"copy_of_task_id":null,"user_id":47853587,"priority":10,"host_name":"kirstin","storage":null,"depends_on_task_ids":[],"links":{"self":"https:\/\/api.cloudconvert.com\/v2\/tasks\/c277da79-c43f-411c-9cfe-65146f5d13d2"}}],"links":{"self":"https:\/\/api.cloudconvert.com\/v2\/jobs\/9599ced2-e526-410f-9277-f713c082b85a"}}}
`
	req := &cloudConvertJobData{}
	if err := json.Unmarshal([]byte(payload), req); err != nil {
		t.Errorf("json.Unmarshal: %v", err)
	}

	expectedFiles := []string{
		"assignment/2021-06-29/images/01F9BB6JY1B2DQC9K5E83P5908/01EM0K92PVQ1R5SF4%2020JM33RQ70K-1.png",
		"assignment/2021-06-29/images/01F9BB6JY1B2DQC9K5E83P5908/01EM0K92PVQ1R5SF4%2020JM33RQ70K-2.png",
		"assignment/2021-06-29/images/01F9BB6JY1B2DQC9K5E83P5908/01EM0K92PVQ1R5SF4%2020JM33RQ70K-3.png",
		"assignment/2021-06-29/images/01F9BB6JY1B2DQC9K5E83P5908/01EM0K92PVQ1R5SF4%2020JM33RQ70K-4.png",
		"assignment/2021-06-29/images/01F9BB6JY1B2DQC9K5E83P5908/01EM0K92PVQ1R5SF4%2020JM33RQ70K-5.png",
	}

	convertedFiles := parseCloudConvertConvertedFiles(req)
	for i, file := range convertedFiles {
		if file != expectedFiles[i] {
			t.Errorf("unexpected converted file: %q", file)
		}
	}
}
