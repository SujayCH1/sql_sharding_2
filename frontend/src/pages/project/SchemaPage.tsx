import { useParams } from "react-router-dom"
import SchemaEditor from "./schemaPage/SchemaEditor"
import SchemaHistory from "./schemaPage/SchemaHistory"

export default function SchemaPage() {
  const { projectId } = useParams<{ projectId: string }>()

  if (!projectId) {
    return (
      <div className="text-sm text-muted-foreground">
        Invalid project.
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <SchemaEditor projectId={projectId} />
      <SchemaHistory projectId={projectId} />
    </div>
  )
}
