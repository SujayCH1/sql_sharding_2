import { useEffect, useRef, useState } from "react"
import {
  NavLink,
  Outlet,
  useNavigate,
  useParams,
} from "react-router-dom"
import { ArrowLeft, ChevronUp, ChevronDown, Trash2 } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"

import { FetchProjectByID } from "../../wailsjs/go/main/App"
import type { repository } from "../../wailsjs/go/models"

import { EventsOn, EventsOff } from "../../wailsjs/runtime/runtime"

type Project = repository.Project

type LogEvent = {
  level: "info" | "warn" | "error"
  message: string
  source: string
  timestamp: string
  fields?: Record<string, string>
}

function SideNav({ projectId }: { projectId: string }) {
  const base = `/projects/${projectId}`

  const linkClass = ({ isActive }: { isActive: boolean }) =>
    `px-3 py-2 text-sm rounded-md transition ${
      isActive
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

      <NavLink to={`${base}/schema`} className={linkClass}>
        Schema
      </NavLink>
    </nav>
  )
}

export default function ProjectPage() {
  const navigate = useNavigate()
  const { projectId } = useParams()

  const [project, setProject] = useState<Project | null>(null)
  const [loading, setLoading] = useState(true)

  const [logs, setLogs] = useState<LogEvent[]>([])
  const [consoleOpen, setConsoleOpen] = useState(true)

  const consoleBodyRef = useRef<HTMLDivElement | null>(null)

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

  useEffect(() => {
    const handler = (event: LogEvent) => {
      setLogs(prev => [...prev, event])
    }

    EventsOn("log:event", handler)

    return () => {
      EventsOff("log:event")
    }
  }, [])

  // auto-scroll on new logs
  useEffect(() => {
    if (!consoleOpen) return
    const el = consoleBodyRef.current
    if (!el) return
    el.scrollTop = el.scrollHeight
  }, [logs, consoleOpen])

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
    <div className="p-6 space-y-6 pb-64">
      <Button
        variant="ghost"
        size="sm"
        onClick={() => navigate("/")}
        className="w-fit"
      >
        <ArrowLeft className="mr-2 h-4 w-4" />
        Back to Projects
      </Button>

      <div className="space-y-1">
        <h1 className="text-2xl font-semibold">{project.name}</h1>
        <p className="text-sm text-muted-foreground">
          {project.description || "No description provided"}
        </p>
      </div>

      <Separator />

      <div className="flex gap-6">
        <SideNav projectId={projectId} />

        <Separator orientation="vertical" className="h-auto" />

        <div className="flex-1">
          <Outlet />
        </div>
      </div>

      {/* Bottom Console */}
      <div
        className={`fixed bottom-0 left-0 right-0 border-t border-neutral-700 bg-neutral-900 transition-all ${
          consoleOpen ? "h-56" : "h-10"
        }`}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-3 h-10 border-b border-neutral-700 bg-neutral-800">
          <div className="text-sm font-medium text-neutral-200">
            Console
          </div>

          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="icon"
              onClick={() => setLogs([])}
              title="Clear logs"
            >
              <Trash2 className="h-4 w-4 text-neutral-300" />
            </Button>

            <Button
              variant="ghost"
              size="icon"
              onClick={() => setConsoleOpen(o => !o)}
            >
              {consoleOpen ? (
                <ChevronDown className="h-4 w-4 text-neutral-300" />
              ) : (
                <ChevronUp className="h-4 w-4 text-neutral-300" />
              )}
            </Button>
          </div>
        </div>

        {/* Body */}
        {consoleOpen && (
          <div
            ref={consoleBodyRef}
            className="h-[calc(100%-2.5rem)] overflow-auto px-3 py-2 font-mono text-xs space-y-1 text-neutral-100"
          >
            {logs.length === 0 && (
              <div className="text-neutral-500">
                No logs yet.
              </div>
            )}

            {logs.map((log, idx) => (
              <div key={idx} className="flex gap-2 flex-wrap">
                <span className="text-neutral-500">
                  {new Date(log.timestamp).toLocaleTimeString()}
                </span>

                <span
                  className={
                    log.level === "error"
                      ? "text-red-500"
                      : log.level === "warn"
                      ? "text-amber-400"
                      : "text-neutral-200"
                  }
                >
                  {log.level.toUpperCase()}
                </span>

                <span className="text-neutral-500">
                  [{log.source}]
                </span>

                <span>{log.message}</span>

                {log.fields &&
                  Object.keys(log.fields).length > 0 && (
                    <span className="text-neutral-400">
                      {Object.entries(log.fields)
                        .map(([k, v]) => `${k}=${v}`)
                        .join(" ")}
                    </span>
                  )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
