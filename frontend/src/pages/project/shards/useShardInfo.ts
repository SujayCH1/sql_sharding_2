import { useEffect, useState } from "react"
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

type AdminStatus = "active" | "inactive"

type ShardConnectionForm = {
  host: string
  port: number
  database_name: string
  username: string
  password: string
}

export function useShardInfo(
  shardId: string,
  // navigate: (delta: number) => void
) {
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

  // -------- LOAD SHARD STATUS --------
  useEffect(() => {
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

  // -------- LOAD CONNECTION --------
  useEffect(() => {
    async function loadConnection() {
      try {
        const conn = await FetchConnectionInfo(shardId)
        setConnection(conn)

        setForm({
          host: conn.host,
          port: conn.port,
          database_name: conn.database_name,
          username: conn.username,
          password: "",
        })
      } catch {
        setConnection(null)
      }
    }
    loadConnection()
  }, [shardId])

  // -------- ACTIONS --------
  async function toggleShardStatus() {
    if (isActive) {
      await DeactivateShard(shardId)
      setAdminStatus("inactive")
    } else {
      await ActivateShard(shardId)
      setAdminStatus("active")
    }
  }

  // const [deleteError, setDeleteError] = useState<string | null>(null)

  // async function handleDeleteShard() {
  //   try {
  //     const result = await DeleteShard(shardId)

  //     if (result == "CANNOT_DELETE_ACTIVE_SHARD") {
  //       setDeleteError("Deactivate the shard before deleting it.")
  //       return
  //     }

  //     if (result === "DELETED") {
  //       navigate(-1)
  //       return
  //     }

  //     setDeleteError("Unable to delete shard.")
  //   } catch {
  //     setDeleteError("Unable to delete shard.")
  //   }

  // }

  async function handleDeleteShard(): Promise<string> {
    return await DeleteShard(shardId)
  }


  async function handleSaveConnection() {
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

  return {
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
  }
}
