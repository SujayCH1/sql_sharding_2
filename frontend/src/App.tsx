
import { BrowserRouter, Routes, Route } from "react-router-dom"
import Home from "./pages/HomePage"
import ProjectPage from "./pages/ProjectPage"
import OverviewPage from "./pages/project/OverviewPage"
// import ShardsPage from "./pages/project/shards/ShardsPage"
import ShardsPage from "./pages/project/ShardsPage"
import ShardInfoPage from "./pages/project/shards/ShardInfoPage"
import SchemaPage from "./pages/project/SchemaPage"

function App() {
  return (
    // <>
    //   <Home></Home>
    // </>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home/>} />
        <Route path="/projects/:projectId" element={<ProjectPage />} > 
          <Route index element={<OverviewPage />} />
          <Route path="shards" element={<ShardsPage />} />
          <Route path="schema" element={<SchemaPage />} />
          <Route
            path="/projects/:projectId/shards/:shardId"
            element={<ShardInfoPage />}
          />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App
