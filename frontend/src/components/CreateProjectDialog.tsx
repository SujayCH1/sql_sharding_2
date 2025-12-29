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

import { CreateProject } from "../../wailsjs/go/main/App"

type CreateProjectForm = {
  name: string
  description: string
}

type Props = {
  onProjectCreated: () => void | Promise<void>
}

const initialFormState: CreateProjectForm = {
  name: "",
  description: "",
}

export function CreateProjectDialog({ onProjectCreated }: Props) {
  const [open, setOpen] = useState(false)
  const [form, setForm] = useState<CreateProjectForm>(initialFormState)
  const [loading, setLoading] = useState(false)

  function update<K extends keyof CreateProjectForm>(
    key: K,
    value: CreateProjectForm[K]
  ) {
    setForm(prev => ({ ...prev, [key]: value }))
  }

  async function handleCreate() {
    if (!form.name.trim()) return

    try {
      setLoading(true)

      // ✅ correct backend call
      await CreateProject(form.name, form.description)

      // ✅ refresh projects from DB
      await onProjectCreated()

      // ✅ reset & close
      setForm(initialFormState)
      setOpen(false)
    } catch (err) {
      console.error("Failed to create project", err)
    } finally {
      setLoading(false)
    }
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
              placeholder="My Sharded Database"
            />
          </div>

          <div>
            <Label>Description</Label>
            <Textarea
              value={form.description}
              onChange={e => update("description", e.target.value)}
              placeholder="Optional description"
            />
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <Button
              variant="secondary"
              onClick={() => setOpen(false)}
              disabled={loading}
            >
              Cancel
            </Button>

            <Button
              onClick={handleCreate}
              disabled={loading || !form.name.trim()}
            >
              {loading ? "Creating..." : "Create Project"}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
