import { Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import BillList from './pages/BillList'
import BillDetail from './pages/BillDetail'
import EmailAction from './pages/EmailAction'

function App() {
  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<BillList />} />
        <Route path="bill/:id" element={<BillDetail />} />
        <Route path="email/:billId" element={<EmailAction />} />
      </Route>
    </Routes>
  )
}

export default App
