import { useState } from "react"

import { ProjectCard } from "@/components/ProjectCard"
import { CreateProjectDialog } from "@/components/CreateProjectDialog"

import type { Project } from "@/types/project"

export default function Home() {
  const [projects, setProjects] = useState<Project[]>([])

  function handleCreateProject(project: Project) {
    setProjects(prev => [...prev, project])
  }

  return (
    <div className="p-6 space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-semibold">Projects</h1>
        <CreateProjectDialog onCreate={handleCreateProject} />
      </div>

      {projects.length === 0 ? (
        <div className="text-sm text-muted-foreground">
          No projects yet. Create your first project.
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {projects.map(project => (
            <ProjectCard key={project.id} project={project} />
          ))}
        </div>
      )}
    </div>
  )
}
