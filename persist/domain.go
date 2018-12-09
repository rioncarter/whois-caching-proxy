package persist

import "time"

type Domain struct{
	uid           int64
	Name          string
	RegisteredRaw string
	Registered    string
	RegisteredDate	time.Time
}
