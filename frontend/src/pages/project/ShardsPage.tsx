import { useEffect, useState } from "react"
import { useParams, useNavigate } from "react-router-dom"

import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

import { ListShards, AddShard } from "../../../wailsjs/go/main/App"

type Shard = {
  id: string
  shard_index: number
  status: string
  created_at: string
}

export default function ShardsPage() {
  const { projectId } = useParams()
  const navigate = useNavigate()
  const [shards, setShards] = useState<Shard[]>([])

  async function loadShards() {
    if (!projectId) return
    const data = await ListShards(projectId)
    setShards(data)
  }

  async function handleAddShard() {
    if (!projectId) return
    await AddShard(projectId)
    loadShards()
  }

  useEffect(() => {
    loadShards()
  }, [projectId])

  return (
    <div className="p-6 space-y-4">
      <div className="flex justify-between items-center">
        <Button variant="ghost" onClick={() => navigate(-1)}>
          ← Back
        </Button>

        <Button onClick={handleAddShard}>+ Add Shard</Button>
      </div>

      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Index</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Created</TableHead>
            <TableHead className="text-right">Info</TableHead>
          </TableRow>
        </TableHeader>

        <TableBody>
          {shards.map((shard) => (
            <TableRow key={shard.id}>
              <TableCell>{shard.shard_index}</TableCell>
              <TableCell className="capitalize">{shard.status}</TableCell>
              <TableCell>
                {new Date(shard.created_at).toLocaleString()}
              </TableCell>
              <TableCell className="text-right">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() =>
                    navigate(
                      `/projects/${projectId}/shards/${shard.id}`
                    )
                  }
                >
                  Info
                </Button>
              </TableCell>
            </TableRow>
          ))}

          {shards.length === 0 && (
            <TableRow>
              <TableCell colSpan={4} className="text-center text-muted-foreground">
                No shards found
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </div>
  )
}
