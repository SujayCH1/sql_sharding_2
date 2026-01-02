import { useEffect, useState } from "react"
import {
  NavLink,
  Outlet,
  useNavigate,
  useParams,
} from "react-router-dom"
import { ArrowLeft } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"

import { FetchProjectByID } from "../../wailsjs/go/main/App"
import type { repository } from "../../wailsjs/go/models"

type Project = repository.Project

function SideNav({ projectId }: { projectId: string }) {
  const base = `/projects/${projectId}`

  const linkClass = ({ isActive }: { isActive: boolean }) =>
    `px-3 py-2 text-sm rounded-md transition ${isActive
      ? "bg-muted font-medium"
      : "text-muted-foreground hover:text-foreground hover:bg-muted/50"
    }`

  return (
    <nav className="flex flex-col gap-1 w-40 shrink-0">
      <NavLink end to={base} className={linkClass}>
        Overview
      </NavLink>

      <NavLink to={`${base}/shards`} className={linkClass}>
        Shards
      </NavLink>
    </nav>
  )
}

export default function ProjectPage() {
  const navigate = useNavigate()
  const { projectId } = useParams()

  const [project, setProject] = useState<Project | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    async function loadProject() {
      if (!projectId) return

      try {
        const data = await FetchProjectByID(projectId)
        setProject(data)
      } catch (err) {
        console.error("Failed to load project", err)
      } finally {
        setLoading(false)
      }
    }

    loadProject()
  }, [projectId])

  if (loading) {
    return (
      <div className="p-6 text-sm text-muted-foreground">
        Loading project...
      </div>
    )
  }

  if (!project || !projectId) {
    return (
      <div className="p-6 space-y-4">
        <Button variant="ghost" size="sm" onClick={() => navigate("/")}>
          <ArrowLeft className="mr-2 h-4 w-4" />
          Back to Projects
        </Button>

        <div className="text-sm text-muted-foreground">
          Project not found.
        </div>
      </div>
    )
  }

  return (
    <div className="p-6 space-y-6">
      {/* Back button */}
      <Button
        variant="ghost"
        size="sm"
        onClick={() => navigate("/")}
        className="w-fit"
      >
        <ArrowLeft className="mr-2 h-4 w-4" />
        Back to Projects
      </Button>

      {/* Header */}
      <div className="space-y-1">
        <h1 className="text-2xl font-semibold">{project.name}</h1>
        <p className="text-sm text-muted-foreground">
          {project.description || "No description provided"}
        </p>
      </div>

      {/* Divider */}
      <Separator />

      {/* Nav + content */}
      <div className="flex gap-6">
        <SideNav projectId={projectId} />

        <Separator orientation="vertical" className="h-auto" />

        <div className="flex-1">
          <Outlet />
        </div>
      </div>

    </div>
  )
}
