export interface Project {
  id: string
  name: string
  databaseType: "Postgres" | "MySQL"
  shardCount: number
  status: "ACTIVE" | "PAUSED" | "ERROR"
  createdAt: string
}

export interface CreateProjectInput {
  name: string
  databaseType: "Postgres" | "MySQL"
  connectionString: string
  shardCount: number
  description?: string
}
