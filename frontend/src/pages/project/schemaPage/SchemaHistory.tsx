import { useEffect, useState } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

import { GetSchemaHistory } from "../../../../wailsjs/go/main/App"
import type { repository } from "../../../../wailsjs/go/models"

type ProjectSchema = repository.ProjectSchema

type Props = {
  projectId: string
}

export default function SchemaHistory({ projectId }: Props) {
  const [schemas, setSchemas] = useState<ProjectSchema[]>([])

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
          <div className="space-y-2 text-sm">
            {schemas.map((schema) => (
              <div key={schema.id}>
                <span className="font-medium">v{schema.version}</span>{" "}
                <span className="text-muted-foreground">
                  ({schema.state})
                </span>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
