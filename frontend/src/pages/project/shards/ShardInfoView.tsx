import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog"

import { repository } from "../../../../wailsjs/go/models"

type ShardConnectionForm = {
  host: string
  port: number
  database_name: string
  username: string
  password: string
}

type Props = {
  shardId: string
  adminStatus: "active" | "inactive"
  loadingStatus: boolean
  isActive: boolean
  connection: repository.ShardConnection | null
  connDialogOpen: boolean
  setConnDialogOpen: (v: boolean) => void
  form: ShardConnectionForm
  setForm: React.Dispatch<React.SetStateAction<ShardConnectionForm>>
  toggleShardStatus: () => void
  handleDeleteShard: () => Promise<string>
  handleSaveConnection: () => void
}

export function ShardInfoView({
  shardId,
  adminStatus,
  loadingStatus,
  isActive,
  connection,
  connDialogOpen,
  setConnDialogOpen,
  form,
  setForm,
  toggleShardStatus,
  handleDeleteShard,
  handleSaveConnection,
}: Props) {
  const [deleteError, setDeleteError] = useState<string | null>(null)

  async function onDeleteShard() {
    try {
      const result = await handleDeleteShard()

      if (result === "CANNOT_DELETE_ACTIVE_SHARD") {
        setDeleteError("Deactivate the shard before deleting it.")
        return
      }

      if (result === "DELETED") {
        history.back()
        return
      }

      setDeleteError("Unable to delete shard.")
    } catch {
      setDeleteError("Unable to delete shard.")
    }
  }



  return (
    <div className="p-6 space-y-6 max-w-6xl">
      <Button variant="ghost" onClick={() => history.back()}>
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
                  <Button variant="destructive">Delete Shard</Button>
                </AlertDialogTrigger>
                <AlertDialogContent>
                  <AlertDialogHeader>
                    <AlertDialogTitle>Delete shard?</AlertDialogTitle>
                  </AlertDialogHeader>
                  <div className="flex justify-end gap-2">
                    <AlertDialogCancel>Cancel</AlertDialogCancel>
                    <AlertDialogAction
                      className="bg-destructive"
                      onClick={onDeleteShard}
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

      {/* DELETE ERROR DIALOG */}
      <AlertDialog open={!!deleteError} onOpenChange={() => setDeleteError(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Action blocked</AlertDialogTitle>
            <AlertDialogDescription>
              {deleteError}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogAction onClick={() => setDeleteError(null)}>
              OK
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
