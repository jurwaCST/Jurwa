package domain

// Status is the status of a report. The numbers match jurwa.common.v1.Status,
// but the domain does not import the generated proto code.
type Status int32

const (
	StatusUnspecified Status = 0
	Open              Status = 1
	Assigned          Status = 2
	InProgress        Status = 3
	Resolved          Status = 4
	Escalated         Status = 5 // set only by the system when the SLA window passes
)

var statusNames = map[Status]string{
	Open:       "OPEN",
	Assigned:   "ASSIGNED",
	InProgress: "IN_PROGRESS",
	Resolved:   "RESOLVED",
	Escalated:  "ESCALATED",
}

// Valid reports whether s is one of the five real statuses.
func (s Status) Valid() bool {
	_, ok := statusNames[s]
	return ok
}

// String returns the upper-case code, the same one stored in the database.
func (s Status) String() string {
	if name, ok := statusNames[s]; ok {
		return name
	}
	return "STATUS_UNSPECIFIED"
}
