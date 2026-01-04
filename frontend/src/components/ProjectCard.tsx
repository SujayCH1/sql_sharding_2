import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { MoreVertical } from "lucide-react"

import type { repository } from "../../wailsjs/go/models"

type Project = repository.Project

type Props = {
  project: Project
}

export function ProjectCard({ project }: Props) {
  const isActive = project.status === "active"

  return (
    <Card className="w-full max-w-sm">
      <CardHeader>
        <CardTitle className="flex justify-between items-center">
          <span className="truncate">{project.name}</span>

          <Button variant="ghost" size="icon">
            <MoreVertical className="h-4 w-4" />
          </Button>
        </CardTitle>
      </CardHeader>

      <CardContent className="space-y-3 text-sm">
        <Badge variant={isActive ? "default" : "secondary"}>
          {isActive ? "active" : "inactive"}
        </Badge>

        {project.description && (
          <p className="text-sm text-muted-foreground line-clamp-3">
            {project.description}
          </p>
        )}
      </CardContent>
    </Card>
  )
}

