import { useState } from "react"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

import type { CreateProjectInput, Project } from "@/types/project"

interface Props {
  onCreate: (project: Project) => void
}

const initialFormState: CreateProjectInput = {
  name: "",
  databaseType: "Postgres",
  connectionString: "",
  shardCount: 4,
  description: "",
}


export function CreateProjectDialog({ onCreate }: Props) {
  const [open, setOpen] = useState(false)

  const [form, setForm] = useState<CreateProjectInput>(initialFormState)

  function update<K extends keyof CreateProjectInput>(
    key: K,
    value: CreateProjectInput[K]
  ) {
    setForm(prev => ({ ...prev, [key]: value }))
  }

  function handleCreate() {
    const newProject: Project = {
      id: crypto.randomUUID(),
      name: form.name,
      databaseType: form.databaseType,
      shardCount: form.shardCount,
      status: "ACTIVE",
      createdAt: new Date().toISOString(),
    }

    onCreate(newProject)
    setForm(initialFormState)
    setOpen(false) // close dialog
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>+ New Project</Button>
      </DialogTrigger>

      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Create New Project</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          <div>
            <Label>Project Name</Label>
            <Input
              value={form.name}
              onChange={e => update("name", e.target.value)}
            />
          </div>

          <div>
            <Label>Database Type</Label>
            <Select
              value={form.databaseType}
              onValueChange={v =>
                update("databaseType", v as CreateProjectInput["databaseType"])
              }
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="Postgres">PostgreSQL</SelectItem>
                <SelectItem value="MySQL">MySQL</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div>
            <Label>Connection String</Label>
            <Input
              type="password"
              value={form.connectionString}
              onChange={e => update("connectionString", e.target.value)}
            />
          </div>

          <div>
            <Label>Initial Shard Count</Label>
            <Input
              type="number"
              min={1}
              value={form.shardCount}
              onChange={e => update("shardCount", Number(e.target.value))}
            />
          </div>

          <div>
            <Label>Description</Label>
            <Textarea
              value={form.description}
              onChange={e => update("description", e.target.value)}
            />
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <Button variant="secondary" onClick={() => setOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreate}>Create Project</Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
