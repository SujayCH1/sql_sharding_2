import { useParams, useNavigate } from "react-router-dom"
import { useShardInfo } from "./useShardInfo"
import { ShardInfoView } from "./ShardInfoView"

export default function ShardInfoPage() {
  const { shardId } = useParams()
  const navigate = useNavigate()

  const state = useShardInfo(shardId ?? "", navigate)

  if (!shardId) {
    return <div>Invalid shard</div>
  }

  return <ShardInfoView {...state} />
}
