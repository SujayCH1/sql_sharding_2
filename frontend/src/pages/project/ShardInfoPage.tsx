import { useEffect, useState } from "react"
import { useParams, useNavigate } from "react-router-dom"

import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog"

import {
  ActivateShard,
  DeactivateShard,
  DeleteShard,
  FetchShardStatus,
} from "../../../wailsjs/go/main/App"

// ================= TYPES =================

type AdminStatus = "active" | "inactive"

type ConnectionInfo = {
  host: string
  port: string
  database: string
  username: string
  ssl: boolean
  updatedAt: string
}

export default function ShardInfoPage() {
  const shardId = useParams().shardId as string
  const navigate = useNavigate()

  // ================= STATE =================

  const [adminStatus, setAdminStatus] = useState<AdminStatus>("inactive")
  const [loadingStatus, setLoadingStatus] = useState(true)

  const [connection, setConnection] = useState<ConnectionInfo | null>(null)
  const [connDialogOpen, setConnDialogOpen] = useState(false)

  // ================= LOAD SHARD STATUS =================

  useEffect(() => {
    if (!shardId) return

    async function loadStatus() {
      try {
        const status = await FetchShardStatus(shardId)
        setAdminStatus(status as AdminStatus)
      } finally {
        setLoadingStatus(false)
      }
    }

    loadStatus()
  }, [shardId])

  const isActive = adminStatus === "active"

  // ================= ACTION HANDLERS =================

  async function toggleShardStatus() {
    if (!shardId) return

    if (isActive) {
      await DeactivateShard(shardId)
      setAdminStatus("inactive")
    } else {
      await ActivateShard(shardId)
      setAdminStatus("active")
    }
  }

  async function handleDeleteShard() {
  if (!shardId) return

  const result = await DeleteShard(shardId)

  if (result === "CANNOT_DELETE_ACTIVE_SHARD") {
    alert("Deactivate the shard before deleting it.")
    return
  }

  if (result === "DELETED") {
    navigate(-1)
    return
  }
}


  // ================= RENDER =================

  return (
    <div className="p-6 space-y-6 max-w-6xl">
      <Button variant="ghost" onClick={() => navigate(-1)}>
        ← Back
      </Button>

      {/* ================= MAIN ================= */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">

        {/* ========== LEFT: INFO ========== */}
        <div className="space-y-4">

          {/* Shard Overview */}
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-base">Shard Overview</CardTitle>
            </CardHeader>
            <CardContent className="text-sm space-y-1">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Shard ID</span>
                <span>{shardId}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Admin Status</span>
                <span className="capitalize">
                  {loadingStatus ? "Loading..." : adminStatus}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Created</span>
                <span>—</span>
              </div>
              <div className="pt-2 border-t mt-2 flex justify-between">
                <span className="text-muted-foreground">Ping</span>
                <span>Unknown</span>
              </div>
            </CardContent>
          </Card>

          {/* Connection Info */}
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-base">Connection</CardTitle>
            </CardHeader>
            <CardContent className="text-sm space-y-1">
              {connection ? (
                <>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Host</span>
                    <span>{connection.host}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">Database</span>
                    <span>{connection.database}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">User</span>
                    <span>{connection.username}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">SSL</span>
                    <span>{connection.ssl ? "Yes" : "No"}</span>
                  </div>
                  <div className="text-xs text-muted-foreground pt-2">
                    Updated {connection.updatedAt}
                  </div>
                </>
              ) : (
                <div className="text-muted-foreground">
                  No connection configured
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        {/* ========== RIGHT: ACTIONS ========== */}
        <div className="space-y-4">

          {/* Actions */}
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-base">Actions</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col gap-2">

              {/* Activate / Deactivate */}
              <AlertDialog>
                <AlertDialogTrigger asChild>
                  <Button size="lg" disabled={loadingStatus}>
                    {isActive ? "Deactivate Shard" : "Activate Shard"}
                  </Button>
                </AlertDialogTrigger>
                <AlertDialogContent>
                  <AlertDialogHeader>
                    <AlertDialogTitle>
                      {isActive ? "Deactivate shard?" : "Activate shard?"}
                    </AlertDialogTitle>
                  </AlertDialogHeader>
                  <div className="flex justify-end gap-2">
                    <AlertDialogCancel>Cancel</AlertDialogCancel>
                    <AlertDialogAction onClick={toggleShardStatus}>
                      Confirm
                    </AlertDialogAction>
                  </div>
                </AlertDialogContent>
              </AlertDialog>

              {/* Delete */}
              <AlertDialog>
                <AlertDialogTrigger asChild>
                  <Button variant="destructive" size="sm">
                    Delete Shard
                  </Button>
                </AlertDialogTrigger>
                <AlertDialogContent>
                  <AlertDialogHeader>
                    <AlertDialogTitle>Delete shard?</AlertDialogTitle>
                  </AlertDialogHeader>
                  <div className="flex justify-end gap-2">
                    <AlertDialogCancel>Cancel</AlertDialogCancel>
                    <AlertDialogAction
                      className="bg-destructive"
                      onClick={handleDeleteShard}
                    >
                      Delete
                    </AlertDialogAction>
                  </div>
                </AlertDialogContent>
              </AlertDialog>

            </CardContent>
          </Card>

          {/* Connection Settings */}
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-base">Connection Settings</CardTitle>
              <Button
                size="sm"
                variant="outline"
                onClick={() => setConnDialogOpen(true)}
              >
                {connection ? "Change" : "Add"}
              </Button>
            </CardHeader>
          </Card>
        </div>
      </div>

      {/* ================= FUTURE ================= */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-base">Shard Data & Analytics</CardTitle>
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground">
          Coming soon
        </CardContent>
      </Card>

      {/* ================= CONNECTION DIALOG ================= */}
      <AlertDialog open={connDialogOpen} onOpenChange={setConnDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {connection ? "Change Connection" : "Add Connection"}
            </AlertDialogTitle>
          </AlertDialogHeader>

          <div className="space-y-3">
            <div>
              <Label>Host</Label>
              <Input placeholder="localhost" />
            </div>
            <div>
              <Label>Port</Label>
              <Input placeholder="5432" />
            </div>
            <div>
              <Label>Database</Label>
              <Input placeholder="app_db" />
            </div>
            <div>
              <Label>Username</Label>
              <Input placeholder="app_user" />
            </div>
            <div>
              <Label>Password</Label>
              <Input type="password" />
            </div>
          </div>

          <div className="flex justify-end gap-2 mt-4">
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                // TEMP — backend persistence later
                setConnection({
                  host: "localhost",
                  port: "5432",
                  database: "app_db",
                  username: "app_user",
                  ssl: false,
                  updatedAt: new Date().toLocaleString(),
                })
                setConnDialogOpen(false)
              }}
            >
              Save
            </AlertDialogAction>
          </div>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
