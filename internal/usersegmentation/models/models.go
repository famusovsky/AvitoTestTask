package models

type UserSegmentationDbProcessor interface {
	AddSegment(slug string) (int, error)
	DeleteSegment(slug string) error
	ModifyUser(id int, append []string, remove []string) error
	CheckupUser(id int) ([]string, error)
}

type Segment struct {
	Slug string `json:"slug"`
}

type User struct {
	Id int `json:"id"`
}

type UserModification struct {
	User
	Append []string `json:"append"`
	Remove []string `json:"remove"`
}
