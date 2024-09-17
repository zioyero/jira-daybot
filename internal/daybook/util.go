package daybook

func TasksByStatus(tasks []*Task) map[string][]*Task {
	byStatus := make(map[string][]*Task)
	for _, task := range tasks {
		byStatus[task.Status] = append(byStatus[task.Status], task)
	}
	return byStatus
}
