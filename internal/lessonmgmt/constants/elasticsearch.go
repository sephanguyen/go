package constants

const (
	LessonIndexName    = "lesson"
	LessonIndexMapping = `{
		"mappings": {
		  "properties":{
			"start_time":{
				"type":"date"
			},
			"end_time":{
				"type":"date"
			},
			"teaching_method":{
				"type":"keyword"
			},
			"teaching_medium":{
				"type":"keyword"
			},
			"lesson_id":{
				"type":"keyword"
			},
			"location_id":{
				"type":"keyword"
			},
			"class_id":{
				"type":"keyword"
			},
			"course_id":{
				"type":"keyword"
			},
			"resource_path":{
				"type":"keyword"
			},
			"lesson_members":{
				"properties":{
				  "id":{
					  "type":"keyword"
				  },
				  "name":{
					  "type":"text"
				  },
				  "current_grade":{
					  "type":"integer"
				  },
				  "course_id":{
					  "type":"keyword"
				  }
				}
			},
			"created_at":{
				"type":"date"
			},
			"deleted_at":{
				"type":"date"
			},
			"updated_at":{
				"type":"date"
			}
		  }
		}
	  }`
)
