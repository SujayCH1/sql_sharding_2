package main

import (
	"database/sql"
	"strings"
)

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
// to check if the shard isdeactivated before it is deleted
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

// to check if the project is inactive before committing in the schema
func (a *App) checkIfProjectInactive(projectID string) (bool, error) {
	status, err := a.FetchProjectStatus(projectID)
	if err != nil {
		return false, err
	}

	if status == "inactive" {
		return true, nil
	}

	return false, nil
}

// to check if the schema is in draft state before commiting
func (a *App) checkIfSchemaDraft(schemaID string) (bool, error) {
	status, err := a.GetProjectSchemaStatus(schemaID)
	if err != nil {
		return false, err
	}

	if status == "draft" {
		return true, nil
	}

	return false, nil
}

// to check if any schema is already pending or applying for a project
func (a *App) checkIfSchemaInFlight(projectID string) (bool, error) {

	schemas, err := a.GetSchemaHistory(projectID)
	if err != nil {
		return false, err
	}

	for _, schema := range schemas {
		if schema.State == "pending" || schema.State == "applying" {
			return true, nil
		}
	}

	return false, nil
}

// to check if DDL is destructive after first committed schema
func (a *App) checkIfDDLDestructive(projectID string, ddlSQL string) (bool, error) {

	appliedSchema, err := a.GetCurrentSchema(projectID)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	if appliedSchema == nil {
		return false, nil
	}

	lowered := strings.ToLower(ddlSQL)

	if strings.Contains(lowered, "drop table") ||
		strings.Contains(lowered, "drop column") ||
		strings.Contains(lowered, "truncate") ||
		strings.Contains(lowered, "alter table") && strings.Contains(lowered, " drop ") {
		return true, nil
	}

	return false, nil
}

// to ensure only DDL statements are present
func (a *App) checkIfOnlyDDL(ddlSQL string) bool {

	lowered := strings.ToLower(ddlSQL)

	disallowed := []string{
		"insert ",
		"update ",
		"delete ",
		"select ",
		"merge ",
	}

	for _, keyword := range disallowed {
		if strings.Contains(lowered, keyword) {
			return false
		}
	}

	return true
}
