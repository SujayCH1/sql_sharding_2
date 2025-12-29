import { useEffect, useState } from "react"

import { ProjectCard } from "@/components/ProjectCard"
import { CreateProjectDialog } from "@/components/CreateProjectDialog"

import { ListProjects } from "../../wailsjs/go/main/App"
import type { repository } from "../../wailsjs/go/models"

type Project = repository.Project

export default function Home() {
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)

  // Load projects from backend (DB)
  async function loadProjects() {
    try {
      const data = await ListProjects()
      setProjects(data)
    } catch (err) {
      console.error("Failed to load projects", err)
    } finally {
      setLoading(false)
    }
  }

  // Load projects when Home mounts
  useEffect(() => {
    loadProjects()
  }, [])

  return (
    <div className="p-6 space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-semibold">Projects</h1>

        {/* Dialog no longer passes a project object */}
        <CreateProjectDialog onProjectCreated={loadProjects} />
      </div>

      {loading ? (
        <div className="text-sm text-muted-foreground">
          Loading projects...
        </div>
      ) : projects.length === 0 ? (
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
