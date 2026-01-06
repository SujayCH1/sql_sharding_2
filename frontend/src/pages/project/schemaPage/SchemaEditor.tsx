import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Textarea } from "@/components/ui/textarea"
import { Badge } from "@/components/ui/badge"

import {
  CreateSchemaDraft,
  CommitSchemaDraft,
  DeleteSchemaDraft,
  GetCurrentSchema,
  FetchProjectStatus,
  ExecuteProjectSchema,
  RetrySchemaExecution,
} from "../../../../wailsjs/go/main/App"

import type { repository } from "../../../../wailsjs/go/models"

type ProjectSchema = repository.ProjectSchema

type Props = {
  projectId: string
}

export default function SchemaEditor({ projectId }: Props) {
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [executing, setExecuting] = useState(false)

  const [projectStatus, setProjectStatus] = useState("")
  const [currentSchema, setCurrentSchema] =
    useState<ProjectSchema | null>(null)

  const [ddl, setDDL] = useState("")
  const [isEditingDraft, setIsEditingDraft] = useState(false)

  /* ---------------------------------- */
  /* Load current schema + project      */
  /* ---------------------------------- */
  async function refresh() {
    const status = await FetchProjectStatus(projectId)
    setProjectStatus(status)

    try {
      const schema = await GetCurrentSchema(projectId)
      setCurrentSchema(schema)
      setDDL(schema.ddl_sql ?? "")
      setIsEditingDraft(schema.state === "draft")
    } catch {
      setCurrentSchema(null)
      setDDL("")
      setIsEditingDraft(true)
    }
  }

  useEffect(() => {
    refresh().finally(() => setLoading(false))
  }, [projectId])

  /* ---------------------------------- */
  /* Derived state                      */
  /* ---------------------------------- */
  const isProjectActive = projectStatus === "active"
  const schemaState = currentSchema?.state

  const isDraft = schemaState === "draft"
  const isPending = schemaState === "pending"
  const isApplying = schemaState === "applying"
  const isApplied = schemaState === "applied"
  const isFailed = schemaState === "failed"

  const canEditDDL =
    isEditingDraft &&
    !isPending &&
    !isApplying &&
    !isApplied

  const canCommit = !isProjectActive && isDraft
  const canExecute = isProjectActive && (isPending || isFailed)

  /* ---------------------------------- */
  /* Actions                            */
  /* ---------------------------------- */
  function handleStartNewDraft() {
    setDDL("")
    setCurrentSchema(null)
    setIsEditingDraft(true)
  }

  async function handleSaveDraft() {
    if (!ddl.trim()) return

    setSaving(true)
    try {
      await CreateSchemaDraft(projectId, ddl)
      await refresh()
      setIsEditingDraft(false)
    } finally {
      setSaving(false)
    }
  }

  async function handleCommit() {
    if (!currentSchema || !canCommit) return

    setSaving(true)
    try {
      await CommitSchemaDraft(projectId, currentSchema.id)
      await refresh()
    } finally {
      setSaving(false)
    }
  }

  async function handleDiscard() {
    if (!currentSchema) return

    setSaving(true)
    try {
      await DeleteSchemaDraft(currentSchema.id)
      await refresh()
      setIsEditingDraft(false)
    } finally {
      setSaving(false)
    }
  }

  async function handleExecute() {
    if (!canExecute) return

    setExecuting(true)
    try {
      await ExecuteProjectSchema(projectId)

      // clear editor after successful execution
      setDDL("")
      setCurrentSchema(null)
      setIsEditingDraft(false)

      await refresh()
    } finally {
      setExecuting(false)
    }
  }

  async function handleRetry() {
    if (!canExecute) return

    setExecuting(true)
    try {
      await RetrySchemaExecution(projectId)
      await refresh()
    } finally {
      setExecuting(false)
    }
  }

  /* ---------------------------------- */
  /* Render                             */
  /* ---------------------------------- */
  if (loading) {
    return (
      <div className="text-sm text-muted-foreground">
        Loading schema…
      </div>
    )
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle>Schema DDL</CardTitle>
        {currentSchema && (
          <Badge variant="outline">{schemaState}</Badge>
        )}
      </CardHeader>

      <CardContent className="space-y-4">
        <Textarea
          value={ddl}
          onChange={(e) => setDDL(e.target.value)}
          className="min-h-[220px] font-mono"
          disabled={!canEditDDL}
        />

        {isProjectActive && (
          <div className="text-sm text-yellow-600">
            Project is active. Schema commit is disabled.
          </div>
        )}

        <div className="flex flex-wrap gap-2">
          {/* Create new draft */}
          {isApplied && !isEditingDraft && (
            <Button onClick={handleStartNewDraft}>
              Create New Draft
            </Button>
          )}

          {/* Save draft */}
          {isEditingDraft && !isDraft && (
            <Button
              onClick={handleSaveDraft}
              disabled={!ddl.trim() || saving}
            >
              Save Draft
            </Button>
          )}

          {/* Draft actions */}
          {isDraft && (
            <>
              <Button
                onClick={handleCommit}
                disabled={saving || !canCommit}
              >
                Commit Schema
              </Button>

              <Button
                variant="outline"
                onClick={handleDiscard}
                disabled={saving}
              >
                Discard Draft
              </Button>
            </>
          )}

          {/* Execute */}
          {canExecute && isPending && (
            <Button onClick={handleExecute} disabled={executing}>
              {executing ? "Executing…" : "Execute Schema"}
            </Button>
          )}

          {/* Retry */}
          {canExecute && isFailed && (
            <Button
              variant="outline"
              onClick={handleRetry}
              disabled={executing}
            >
              {executing ? "Retrying…" : "Retry Execution"}
            </Button>
          )}
        </div>

        {isApplying && (
          <div className="text-sm text-muted-foreground">
            Schema execution in progress…
          </div>
        )}

        {isFailed && (
          <div className="text-sm text-red-600">
            Execution failed:{" "}
            {currentSchema?.error_message ?? "Unknown error"}
          </div>
        )}

        {isApplied && !isEditingDraft && (
          <div className="text-sm text-green-600">
            Schema applied successfully. Create a new draft to continue.
          </div>
        )}
      </CardContent>
    </Card>
  )
}
