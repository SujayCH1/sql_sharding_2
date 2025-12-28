import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { MoreVertical } from "lucide-react"

import type { Project } from "@/types/project"


export function ProjectCard({ project }: { project: Project }) {
  return (
    <Card className="w-full max-w-sm">
      <CardHeader>
        <CardTitle className="flex justify-between items-center">
          {project.name}
          <Button variant="ghost" size="icon">
            <MoreVertical className="h-4 w-4" />
          </Button>
        </CardTitle>
      </CardHeader>

      <CardContent className="space-y-2 text-sm">
        <div>DB: {project.databaseType}</div>
        <div>Shards: {project.shardCount}</div>

        <Badge variant={
          project.status === "ACTIVE" ? "default" :
          project.status === "PAUSED" ? "secondary" : "destructive"
        }>
          {project.status}
        </Badge>
      </CardContent>
    </Card>
  )
}
