package main

// helper to ActivateProject in app.go
// to see if there are any other active projects before shard activation
func (a *App) checkAllProjectsInactive() (bool, error) {

	projects, err := a.ListProjects()
	if err != nil {
		return false, err
	}

	for _, project := range projects {
		if project.Status == "active" {
			return false, nil
		}
	}

	return true, nil
}

// helper func to ActivateProject in app.go
// to see if all project shards are active before project activation
func (a *App) checkAllShardsActive(projectID string) (bool, error) {

	shards, err := a.ListShards(projectID)
	if err != nil {
		return false, err
	}

	if len(shards) == 0 {
		return false, nil
	}

	for _, shard := range shards {
		if shard.Status != "active" {
			return false, nil
		}
	}

	return true, nil
}

// helper func to DeleteShard in app.gp
// to check if the shard isdeactovated before it is deleted
func (a *App) checkIfShardInactive(shardID string) (bool, error) {

	status, err := a.FetchShardStatus(shardID)
	if err != nil {
		return false, err
	}

	if status == "active" {
		return false, nil
	}

	return true, nil
}
