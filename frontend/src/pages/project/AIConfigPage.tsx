import { useState } from "react"
import { useParams } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
// import { UpsertAIConfig } from "../../wailsjs/go/main/App"
import { UpsertAIConfig } from "../../../wailsjs/go/main/App"

export default function AiConfigPage() {
    const { projectId } = useParams()

    const [provider, setProvider] = useState("openai")
    const [apiKey, setApiKey] = useState("")
    const [model, setModel] = useState("")
    const [loading, setLoading] = useState(false)
    const [message, setMessage] = useState("")

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault()
        if (!projectId) return

        setLoading(true)
        setMessage("")

        try {
            await UpsertAIConfig({
                ProjectID: projectId,
                Provider: provider,
                APIKey: apiKey,
                Model: model,
            })

            setMessage("Configuration saved successfully")
        } catch (err) {
            console.error(err)
            setMessage("Failed to save config")
        } finally {
            setLoading(false)
        }
    }

    return (
        <div className="space-y-6 max-w-md">
            <div>
                <h2 className="text-lg font-semibold">AI Configuration</h2>
                <p className="text-sm text-muted-foreground">
                    Configure LLM provider for shard key inference
                </p>
            </div>

            <form onSubmit={handleSubmit} className="space-y-4">

                {/* Provider */}
                <div className="space-y-2">
                    <Label>Provider</Label>
                    <select
                        value={provider}
                        onChange={(e) => setProvider(e.target.value)}
                        className="w-full border rounded-md px-3 py-2 bg-background"
                    >
                        <option value="openai">OpenAI</option>
                        <option value="grok">Grok</option>
                        <option value="ollama">Ollama (Local)</option>
                    </select>
                </div>

                {/* API Key */}
                {provider !== "ollama" && (
                    <div className="space-y-2">
                        <Label>API Key</Label>
                        <Input
                            type="password"
                            placeholder="sk-..."
                            value={apiKey}
                            onChange={(e) => setApiKey(e.target.value)}
                            required
                        />
                    </div>
                )}

                {/* Model */}
                <div className="space-y-2">
                    <Label>Model</Label>
                    <Input
                        placeholder="gpt-4o-mini / grok-1"
                        value={model}
                        onChange={(e) => setModel(e.target.value)}
                        required
                    />
                </div>

                {/* Submit */}
                <Button type="submit" disabled={loading}>
                    {loading ? "Saving..." : "Save Configuration"}
                </Button>

                {message && (
                    <p className="text-sm text-muted-foreground">{message}</p>
                )}
            </form>
        </div>
    )
}