import { useEffect, useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"

import { GetSchemaHistory } from "../../../../wailsjs/go/main/App"
import type { repository } from "../../../../wailsjs/go/models"

type ProjectSchema = repository.ProjectSchema

type Props = {
  projectId: string
}

function formatDate(date?: string) {
  if (!date) return "—"
  return new Date(date).toLocaleString()
}

function stateBadge(state: string) {
  switch (state) {
    case "applied":
      return <Badge className="bg-green-600">Applied</Badge>
    case "committed":
      return <Badge className="bg-blue-600">Committed</Badge>
    case "failed":
      return <Badge variant="destructive">Failed</Badge>
    default:
      return <Badge variant="secondary">{state}</Badge>
  }
}

export default function SchemaHistory({ projectId }: Props) {
  const [schemas, setSchemas] = useState<ProjectSchema[]>([])
  const [openDDL, setOpenDDL] = useState<string | null>(null)

  useEffect(() => {
    GetSchemaHistory(projectId).then((result) => {
      setSchemas(result ?? [])
    })
  }, [projectId])

  return (
    <Card>
      <CardHeader>
        <CardTitle>Schema History</CardTitle>
      </CardHeader>

      <CardContent>
        {schemas.length === 0 ? (
          <div className="text-sm text-muted-foreground">
            No schema history.
          </div>
        ) : (
          <div className="space-y-4 text-sm">
            {schemas.map((schema) => (
              <div
                key={schema.id}
                className="rounded-md border p-3 space-y-2"
              >
                {/* Header */}
                <div className="flex items-center justify-between">
                  <div className="font-medium">
                    Version v{schema.version}
                  </div>
                  {stateBadge(schema.state)}
                </div>

                {/* Timestamps */}
                <div className="text-xs text-muted-foreground space-y-1">
                  <div>Created: {formatDate(schema["created _at"])}</div>
                  <div>Committed: {formatDate(schema.commited_at)}</div>
                </div>

                {/* Error */}
                {schema.error_message && (
                  <div className="text-xs text-red-600">
                    ⚠ {schema.error_message}
                  </div>
                )}

                {/* DDL Toggle */}
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() =>
                    setOpenDDL(
                      openDDL === schema.id ? null : schema.id
                    )
                  }
                >
                  {openDDL === schema.id ? "Hide DDL" : "View DDL"}
                </Button>

                {openDDL === schema.id && (
                  <pre className="bg-muted p-2 rounded text-xs overflow-x-auto">
                    {schema.ddl_sql}
                  </pre>
                )}
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
