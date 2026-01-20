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
  FetchShardKeys,
  RecomputeKeys,
  ReplaceShardKeys,
} from "../../../wailsjs/go/main/App"

import type { repository } from "../../../wailsjs/go/models"

type ShardKeyUI = {
  table_name: string
  shard_key_column: string
  is_manual_override: boolean
}

type Project = repository.Project

export default function OverviewPage() {
  const { projectId } = useParams<{ projectId: string }>()

  const [project, setProject] = useState<Project | null>(null)

  const [loading, setLoading] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)

  // ---------- Shard keys ----------
  const [shardKeys, setShardKeys] = useState<ShardKeyUI[]>([])
  const [keysLoading, setKeysLoading] = useState(false)
  const [keysError, setKeysError] = useState<string | null>(null)
  const [dirty, setDirty] = useState(false)

  const [recomputing, setRecomputing] = useState(false)


  useEffect(() => {
    if (!projectId) return
    FetchProjectByID(projectId).then(setProject)
  }, [projectId])


  const loadShardKeys = async () => {
    if (!projectId) return
    setKeysLoading(true)
    setKeysError(null)

    try {
      const keys = await FetchShardKeys(projectId)

      // Map backend models → UI DTOs
      const uiKeys: ShardKeyUI[] = keys.map(k => ({
        table_name: k.table_name,
        shard_key_column: k.shard_key_column,
        is_manual_override: k.is_manual_override,
      }))

      setShardKeys(uiKeys)
      setDirty(false)
    } catch (err) {
      console.error(err)
      setKeysError("Failed to fetch shard keys.")
    } finally {
      setKeysLoading(false)
    }
  }

  useEffect(() => {
    loadShardKeys()
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
        console.error(err)
      }
    } finally {
      setLoading(false)
    }
  }


  const handleRecomputeShardKeys = async () => {
    if (!projectId) return

    setRecomputing(true)
    setKeysError(null)

    try {
      await RecomputeKeys(projectId)
      await loadShardKeys()
    } catch (err) {
      console.error(err)
      setKeysError("Failed to recompute shard keys.")
    } finally {
      setRecomputing(false)
    }
  }


  const handleSaveShardKeys = async () => {
    if (!projectId) return

    const payload: repository.ShardKeyRecord[] = shardKeys.map(k => ({
      TableName: k.table_name,
      ShardKeyColumn: k.shard_key_column,
      IsManual: k.is_manual_override,
    }))

    try {
      await ReplaceShardKeys(projectId, payload)
      setDirty(false)
      await loadShardKeys()
    } catch (err) {
      console.error(err)
      setKeysError("Failed to save shard keys.")
    }
  }

  return (
    <div className="space-y-6">

      {/* ---------- Errors ---------- */}
      {errorMessage && (
        <div className="rounded-md border border-destructive/50 bg-destructive/10 p-3 text-sm text-destructive">
          {errorMessage}
        </div>
      )}

      {keysError && (
        <div className="rounded-md border border-destructive/50 bg-destructive/10 p-3 text-sm text-destructive">
          {keysError}
        </div>
      )}

      {/* ---------- Project status ---------- */}
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
                {isActive
                  ? "Deactivating disconnects all shards and stops routing."
                  : "Activating initializes all shard connections."}
              </AlertDialogDescription>
            </AlertDialogHeader>

            <AlertDialogFooter>
              <AlertDialogCancel disabled={loading}>
                Cancel
              </AlertDialogCancel>
              <AlertDialogAction onClick={handleConfirm} disabled={loading}>
                {isActive ? "Deactivate" : "Activate"}
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>

      {/* ---------- Shard key inference ---------- */}
      <div className="space-y-3">
        <AlertDialog>
          <AlertDialogTrigger asChild>
            <Button variant="outline" disabled={recomputing}>
              Recompute Shard Keys
            </Button>
          </AlertDialogTrigger>

          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Recompute shard keys?</AlertDialogTitle>
              <AlertDialogDescription>
                Keys will be regenerated from schema. Manual overrides will be preserved.
              </AlertDialogDescription>
            </AlertDialogHeader>

            <AlertDialogFooter>
              <AlertDialogCancel disabled={recomputing}>
                Cancel
              </AlertDialogCancel>
              <AlertDialogAction
                onClick={handleRecomputeShardKeys}
                disabled={recomputing}
              >
                Recompute
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>

        {/* ---------- Keys table ---------- */}
        {keysLoading && (
          <p className="text-sm text-muted-foreground">Loading shard keys…</p>
        )}

        {!keysLoading && shardKeys.length === 0 && (
          <p className="text-sm text-muted-foreground">
            No shard keys inferred yet.
          </p>
        )}

        {shardKeys.length > 0 && (
          <>
            <table className="w-full border rounded-md text-sm">
              <thead className="bg-muted">
                <tr>
                  <th className="p-2 text-left">Table</th>
                  <th className="p-2 text-left">Shard Key Column</th>
                  <th className="p-2 text-left">Type</th>
                </tr>
              </thead>

              <tbody>
                {shardKeys.map((key, idx) => (
                  <tr key={key.table_name} className="border-t">
                    <td className="p-2 font-mono">{key.table_name}</td>

                    <td className="p-2">
                      <input
                        className="w-full rounded border px-2 py-1"
                        value={key.shard_key_column}
                        onChange={(e) => {
                          const updated = [...shardKeys]
                          updated[idx] = {
                            ...updated[idx],
                            shard_key_column: e.target.value,
                            is_manual_override: true,
                          }
                          setShardKeys(updated)
                          setDirty(true)
                        }}
                      />
                    </td>

                    <td className="p-2">
                      {key.is_manual_override ? (
                        <Badge>manual</Badge>
                      ) : (
                        <Badge variant="secondary">auto</Badge>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>

            <Button
              onClick={handleSaveShardKeys}
              disabled={!dirty}
            >
              Save Shard Keys
            </Button>
          </>
        )}
      </div>
    </div>
  )
}
