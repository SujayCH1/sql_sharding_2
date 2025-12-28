export interface Project {
  id: string
  name: string
  databaseType: "Postgres" | "MySQL"
  status: "ACTIVE" | "PAUSED" | "ERROR"
  createdAt: string
}

export interface CreateProjectInput {
  name: string
  databaseType: "Postgres" | "MySQL"
  connectionString: string
  description?: string
}
