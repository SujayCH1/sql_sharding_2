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
  ExecuteProjectSchema,
  RetrySchemaExecution,
  GetSchemaCapabilities,
  UpdateProjectSchemaDraft,
} from "../../../../wailsjs/go/main/App"

import type { repository } from "../../../../wailsjs/go/models"

type ProjectSchema = repository.ProjectSchema

type SchemaCapabilities = {
  can_create_draft: boolean
  can_edit_draft: boolean
  can_commit: boolean
  can_execute: boolean
  can_retry: boolean
  reason?: string
}

type Props = {
  projectId: string
}

export default function SchemaEditor({ projectId }: Props) {
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [executing, setExecuting] = useState(false)

  const [schema, setSchema] = useState<ProjectSchema | null>(null)
  const [ddl, setDDL] = useState("")
  const [caps, setCaps] = useState<SchemaCapabilities | null>(null)

  /* ---------------------------------- */
  /* Load state                         */
  /* ---------------------------------- */
  async function refresh() {
    const [schemaRes, capsRes] = await Promise.all([
      GetCurrentSchema(projectId).catch(() => null),
      GetSchemaCapabilities(projectId),
    ])

    setSchema(schemaRes)
    setCaps(capsRes)
    setDDL(schemaRes?.ddl_sql ?? "")
  }

  useEffect(() => {
    refresh().finally(() => setLoading(false))
  }, [projectId])

  /* ---------------------------------- */
  /* Actions                            */
  /* ---------------------------------- */

  async function handleCreateDraft() {
    setSaving(true)
    try {
      await CreateSchemaDraft(projectId, "")
      await refresh()
    } finally {
      setSaving(false)
    }
  }

async function handleSaveDraft() {
  if (!ddl.trim() || !schema) return

  setSaving(true)
  try {
    await UpdateProjectSchemaDraft(projectId, schema.id, ddl) 
    await refresh()
  } finally {
    setSaving(false)
  }
}


  async function handleCommit() {
    if (!schema) return

    setSaving(true)
    try {
      await CommitSchemaDraft(projectId, schema.id)
      await refresh()
    } finally {
      setSaving(false)
    }
  }

  async function handleDiscard() {
    if (!schema) return

    setSaving(true)
    try {
      await DeleteSchemaDraft(schema.id)
      await refresh()
    } finally {
      setSaving(false)
    }
  }

  async function handleExecute() {
    setExecuting(true)
    try {
      await ExecuteProjectSchema(projectId)
      await refresh()
    } finally {
      setExecuting(false)
    }
  }

  async function handleRetry() {
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

  if (loading || !caps) {
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
        {schema && <Badge variant="outline">{schema.state}</Badge>}
      </CardHeader>

      <CardContent className="space-y-4">
        <Textarea
          value={ddl}
          onChange={(e) => setDDL(e.target.value)}
          className="min-h-[220px] font-mono"
          disabled={!caps.can_edit_draft}
        />

        {caps.reason && (
          <div className="text-sm text-muted-foreground">
            {caps.reason}
          </div>
        )}

        <div className="flex flex-wrap gap-2">
          {/* Create Draft */}
          {caps.can_create_draft && (
            <Button
              onClick={handleCreateDraft}
              disabled={saving}
            >
              Create New Draft
            </Button>
          )}

          {/* Save Draft */}
          {caps.can_edit_draft && (
            <Button
              onClick={handleSaveDraft}
              disabled={!ddl.trim() || saving}
            >
              Save Draft
            </Button>
          )}

          {/* Commit */}
          {caps.can_commit && (
            <Button
              onClick={handleCommit}
              disabled={saving}
            >
              Commit Schema
            </Button>
          )}

          {/* Discard */}
          {caps.can_edit_draft && schema && (
            <Button
              variant="outline"
              onClick={handleDiscard}
              disabled={saving}
            >
              Discard Draft
            </Button>
          )}

          {/* Execute */}
          {caps.can_execute && (
            <Button
              onClick={handleExecute}
              disabled={executing}
            >
              {executing ? "Executing…" : "Execute Schema"}
            </Button>
          )}

          {/* Retry */}
          {caps.can_retry && (
            <Button
              variant="outline"
              onClick={handleRetry}
              disabled={executing}
            >
              {executing ? "Retrying…" : "Retry Execution"}
            </Button>
          )}
        </div>

        {schema?.state === "applying" && (
          <div className="text-sm text-muted-foreground">
            Schema execution in progress…
          </div>
        )}

        {schema?.state === "failed" && (
          <div className="text-sm text-red-600">
            Execution failed:{" "}
            {schema.error_message ?? "Unknown error"}
          </div>
        )}

        {schema?.state === "applied" && (
          <div className="text-sm text-green-600">
            Schema applied successfully.
          </div>
        )}
      </CardContent>
    </Card>
  )
}
