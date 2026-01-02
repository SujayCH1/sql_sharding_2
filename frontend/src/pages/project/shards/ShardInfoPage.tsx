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
  FetchConnectionInfo,
  AddConnection,
  UpdateConnection,
} from "../../../../wailsjs/go/main/App"

import { repository } from "../../../../wailsjs/go/models"

// ================= TYPES =================

type AdminStatus = "active" | "inactive"

// UI-only form model (IMPORTANT)
type ShardConnectionForm = {
  host: string
  port: number
  database_name: string
  username: string
  password: string
}

export default function ShardInfoPage() {
  const shardId = useParams().shardId as string
  const navigate = useNavigate()

  // ================= STATE =================

  const [adminStatus, setAdminStatus] = useState<AdminStatus>("inactive")
  const [loadingStatus, setLoadingStatus] = useState(true)

  const [connection, setConnection] =
    useState<repository.ShardConnection | null>(null)

  const [connDialogOpen, setConnDialogOpen] = useState(false)

  const [form, setForm] = useState<ShardConnectionForm>({
    host: "",
    port: 5432,
    database_name: "",
    username: "",
    password: "",
  })

  const isActive = adminStatus === "active"

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

  // ================= LOAD CONNECTION =================

  useEffect(() => {
    if (!shardId) return

    async function loadConnection() {
      try {
        const conn = await FetchConnectionInfo(shardId)
        setConnection(conn)

        // map DB model → form model
        setForm({
          host: conn.host,
          port: conn.port,
          database_name: conn.database_name,
          username: conn.username,
          password: "", // never prefill password
        })
      } catch {
        setConnection(null)
      }
    }

    loadConnection()
  }, [shardId])

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
    }
  }

  async function handleSaveConnection() {
    if (!shardId) return

    const payload: repository.ShardConnection = {
      shard_id: shardId,
      host: form.host,
      port: form.port,
      database_name: form.database_name,
      username: form.username,
      password: form.password,
      created_at: connection?.created_at ?? "",
      updated_at: connection?.updated_at ?? "",
    }

    if (connection) {
      await UpdateConnection(payload)
    } else {
      await AddConnection(payload)
    }

    setConnection(payload)
    setConnDialogOpen(false)
  }

  // ================= RENDER =================

  return (
    <div className="p-6 space-y-6 max-w-6xl">
      <Button variant="ghost" onClick={() => navigate(-1)}>
        ← Back
      </Button>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">

        {/* LEFT */}
        <div className="space-y-4">
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
                <span className="text-muted-foreground">Status</span>
                <span className="capitalize">{adminStatus}</span>
              </div>
            </CardContent>
          </Card>

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
                    <span>{connection.database_name}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">User</span>
                    <span>{connection.username}</span>
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

        {/* RIGHT */}
        <div className="space-y-4">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-base">Actions</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col gap-2">

              <AlertDialog>
                <AlertDialogTrigger asChild>
                  <Button disabled={loadingStatus}>
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

              <AlertDialog>
                <AlertDialogTrigger asChild>
                  <Button variant="destructive">
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

          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-base">Connection Settings</CardTitle>
              <Button
                size="sm"
                variant="outline"
                disabled={isActive}
                onClick={() => setConnDialogOpen(true)}
              >
                {connection ? "Change" : "Add"}
              </Button>
            </CardHeader>
          </Card>
        </div>
      </div>

      {/* CONNECTION DIALOG */}
      <AlertDialog open={connDialogOpen} onOpenChange={setConnDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {connection ? "Update Connection" : "Add Connection"}
            </AlertDialogTitle>
          </AlertDialogHeader>

          <div className="space-y-3">
            <div>
              <Label>Host</Label>
              <Input
                value={form.host}
                onChange={(e) =>
                  setForm({ ...form, host: e.target.value })
                }
              />
            </div>

            <div>
              <Label>Port</Label>
              <Input
                value={form.port}
                onChange={(e) =>
                  setForm({ ...form, port: Number(e.target.value) })
                }
              />
            </div>

            <div>
              <Label>Database</Label>
              <Input
                value={form.database_name}
                onChange={(e) =>
                  setForm({ ...form, database_name: e.target.value })
                }
              />
            </div>

            <div>
              <Label>Username</Label>
              <Input
                value={form.username}
                onChange={(e) =>
                  setForm({ ...form, username: e.target.value })
                }
              />
            </div>

            <div>
              <Label>Password</Label>
              <Input
                type="password"
                onChange={(e) =>
                  setForm({ ...form, password: e.target.value })
                }
              />
            </div>
          </div>

          <div className="flex justify-end gap-2 mt-4">
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleSaveConnection}>
              Save
            </AlertDialogAction>
          </div>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
