package types

import "fmt"

type Owner struct {
	TotalTopics uint64 `json:"total_topics"`
}

func NewOwner() Owner {
	return Owner{}
}

func (o Owner) IncreaseTotalTopics() Owner {
	return Owner{TotalTopics: o.TotalTopics + 1}
}

func (o Owner) String() string {
	return fmt.Sprintf(`Owner:
	TotalTopics: %d`,
		o.TotalTopics,
	)
}
