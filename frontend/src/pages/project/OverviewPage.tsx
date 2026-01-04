import { useEffect, useState } from "react"
import { useParams } from "react-router-dom"

import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import {
  AlertDialog,
  AlertDialogTrigger,
  AlertDialogContent,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogCancel,
  AlertDialogAction,
} from "@/components/ui/alert-dialog"

import {
  Activateproject,
  Deactivateproject,
  FetchProjectByID,
} from "../../../wailsjs/go/main/App"

import type { repository } from "../../../wailsjs/go/models"

type Project = repository.Project

export default function OverviewPage() {
  const { projectId } = useParams<{ projectId: string }>()
  const [project, setProject] = useState<Project | null>(null)
  const [loading, setLoading] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)

  useEffect(() => {
    if (!projectId) return
    FetchProjectByID(projectId).then(setProject)
  }, [projectId])

  if (!project) {
    return <div className="text-sm text-muted-foreground">Loading...</div>
  }

  const isActive = project.status === "active"

  const handleConfirm = async () => {
    if (!projectId) return

    setLoading(true)
    setErrorMessage(null)

    try {
      if (isActive) {
        await Deactivateproject(projectId)
      } else {
        await Activateproject(projectId)
      }

      const updated = await FetchProjectByID(projectId)
      setProject(updated)

    } catch (err: unknown) {
      const message = err?.toString?.() || ""

      if (message.includes("another project is already active")) {
        setErrorMessage(
          "Another project is already active. Please deactivate it first."
        )
      } else if (message.includes("All shards are not active")) {
        setErrorMessage(
          "All shards must be active before this project can be activated."
        )
      } else {
        setErrorMessage("Operation failed. Please try again.")
        console.error("Project activation error:", err)
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-4">
      {errorMessage && (
        <div className="rounded-md border border-destructive/50 bg-destructive/10 p-3 text-sm text-destructive">
          {errorMessage}
        </div>
      )}

      <div className="flex items-center gap-3">
        <Badge variant={isActive ? "default" : "secondary"}>
          {isActive ? "active" : "inactive"}
        </Badge>

        <AlertDialog>
          <AlertDialogTrigger asChild>
            <Button
              disabled={loading}
              variant={isActive ? "destructive" : "default"}
            >
              {isActive ? "Deactivate Project" : "Activate Project"}
            </Button>
          </AlertDialogTrigger>

          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>
                {isActive ? "Deactivate project?" : "Activate project?"}
              </AlertDialogTitle>

              <AlertDialogDescription>
                {isActive ? (
                  <>
                    Deactivating this project will disconnect all shard
                    connections and stop routing queries.
                  </>
                ) : (
                  <>
                    Activating this project will make it the only active
                    project and initialize all shard connections.
                  </>
                )}
              </AlertDialogDescription>
            </AlertDialogHeader>

            <AlertDialogFooter>
              <AlertDialogCancel disabled={loading}>
                Cancel
              </AlertDialogCancel>
              <AlertDialogAction
                onClick={handleConfirm}
                disabled={loading}
              >
                {isActive ? "Deactivate" : "Activate"}
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>

      <p className="text-sm text-muted-foreground">
        Project activation controls shard availability and runtime connections.
      </p>
    </div>
  )
}
