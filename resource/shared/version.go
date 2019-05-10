package shared

import "time"

type VersionList struct {
	Versions []Version `json:"items"`
}

type Version struct {
	Name    string    `json:"versionName"`
	Phase   string    `json:"phase"`
	Updated time.Time `json:"settingUpdatedAt"`
}
