package types

func (o Owner) IncreaseTotalTopics() Owner {
	return Owner{
		TotalTopics: o.TotalTopics + 1,
	}
}
