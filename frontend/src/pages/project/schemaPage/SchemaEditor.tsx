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
} from "../../../../wailsjs/go/main/App"

import type { repository } from "../../../../wailsjs/go/models"

type ProjectSchema = repository.ProjectSchema

type Props = {
  projectId: string 
}

export default function SchemaEditor({ projectId }: Props) {
  const [loading, setLoading] = useState(true)
  const [projectStatus, setProjectStatus] = useState("")
  const [currentSchema, setCurrentSchema] =
    useState<ProjectSchema | null>(null)

  const [ddl, setDDL] = useState("")
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    async function load() {
      try {
        const status = await FetchProjectStatus(projectId)
        setProjectStatus(status)

        try {
          const schema = await GetCurrentSchema(projectId)
          setCurrentSchema(schema)
          setDDL(schema.ddl_sql ?? "")
        } catch {
          setCurrentSchema(null)
          setDDL("")
        }
      } finally {
        setLoading(false)
      }
    }

    load()
  }, [projectId])

  const isProjectActive = projectStatus === "active"
  const isDraft = currentSchema?.state === "draft"

  const isLocked =
    currentSchema?.state === "pending" ||
    currentSchema?.state === "applying" ||
    isProjectActive

  async function handleSaveDraft() {
    if (!ddl.trim()) return
    setSaving(true)
    try {
      const schema = await CreateSchemaDraft(projectId, ddl)
      setCurrentSchema(schema)
    } finally {
      setSaving(false)
    }
  }

  async function handleCommit() {
    if (!currentSchema) return
    setSaving(true)
    try {
      await CommitSchemaDraft(projectId, currentSchema.id)
      const updated = await GetCurrentSchema(projectId)
      setCurrentSchema(updated)
    } finally {
      setSaving(false)
    }
  }

  async function handleDiscard() {
    if (!currentSchema) return
    setSaving(true)
    try {
      await DeleteSchemaDraft(currentSchema.id)
      setCurrentSchema(null)
      setDDL("")
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return <div className="text-sm text-muted-foreground">Loading schema…</div>
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle>Schema DDL</CardTitle>
        {currentSchema && (
          <Badge variant="outline">{currentSchema.state}</Badge>
        )}
      </CardHeader>

      <CardContent className="space-y-4">
        <Textarea
          value={ddl}
          onChange={(e) => setDDL(e.target.value)}
          className="min-h-[220px] font-mono"
          disabled={isLocked}
        />

        <div className="flex gap-2">
          {!currentSchema && (
            <Button onClick={handleSaveDraft} disabled={!ddl.trim() || saving}>
              Save Draft
            </Button>
          )}

          {isDraft && (
            <>
              <Button onClick={handleCommit} disabled={saving}>
                Commit Schema
              </Button>
              <Button variant="outline" onClick={handleDiscard} disabled={saving}>
                Discard Draft
              </Button>
            </>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
