package api

import (
	"encoding/json"

	"github.com/linkxzhou/gongx/packages/server"
)

func AccountLogin(c server.Context) interface{} {
	r := map[string]interface{}{
		"data": map[string]interface{}{
			"created": []interface{}{
				map[string]interface{}{
					"_id":           "634f73721e1d2d44d065cca0",
					"name":          "测试1",
					"created_by":    "634f7332ccacb8e66b529f48",
					"appid":         "u3nnf8",
					"status":        "running",
					"collaborators": []interface{}{},
					"runtime": map[string]interface{}{
						"image": "docker.io/lafyun/app-service:0.8.11",
					},
					"buckets": []interface{}{
						map[string]interface{}{
							"name":  "data",
							"mode":  "public-read",
							"quota": 1073741824,
						},
						map[string]interface{}{
							"name":  "public",
							"mode":  "public-read-write",
							"quota": 1073741824,
						},
					},
					"packages":   []interface{}{},
					"created_at": "2022-10-19T03:48:01.680Z",
					"updated_at": "2022-10-25T01:58:50.882Z",
					"spec": map[string]interface{}{
						"_id":      "634f73711e1d2d44d065cc9f",
						"appid":    "u3nnf8",
						"start_at": "2022-10-19T03:48:01.680Z",
						"end_at":   "2099-12-31T00:00:00.000Z",
						"enabled":  true,
						"spec": map[string]interface{}{
							"_id":               "627103c09e4d81d1bb1ff591",
							"name":              "starter",
							"label":             "starter",
							"request_cpu":       5,
							"request_memory":    67108864,
							"limit_cpu":         500,
							"limit_memory":      134217728,
							"database_capacity": 1073741824,
							"storage_capacity":  3221225472,
							"bandwith":          10485760,
							"out_traffic":       2147483648,
							"priority":          0,
							"enabled":           true,
							"created_at":        "2022-05-03T10:28:16.424Z",
							"updated_at":        "2022-05-03T10:28:16.424Z",
						},
						"created_at": "2022-10-19T03:48:01.682Z",
						"updated_at": "2022-10-19T03:48:01.682Z",
					},
				},
			},
			"joined": []interface{}{},
		},
	}

	if r1, err := json.Marshal(r); err != nil {
		return nil
	} else {
		return string(r1)
	}
}

func AccountProfile(c server.Context) interface{} {
	r := map[string]interface{}{
		"code": 0,
		"data": map[string]interface{}{
			"_id":        "634f7332ccacb8e66b529f48",
			"username":   "zhoulv",
			"phone":      "17724490919",
			"quota":      map[string]interface{}{"app_count": 1},
			"name":       "zhoulv",
			"created_at": "2022-10-19T03:46:58.422Z",
			"updated_at": "2022-10-19T03:46:58.422Z",
		},
	}

	if r1, err := json.Marshal(r); err != nil {
		return nil
	} else {
		return string(r1)
	}
}