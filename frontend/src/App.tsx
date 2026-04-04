import { Route, Routes } from 'react-router';

import { CampaignCreatePage } from './pages/CampaignCreatePage';
import { CampaignListPage } from './pages/CampaignListPage';
import { CampaignPlayPage } from './pages/CampaignPlayPage';

function App() {
  return (
    <Routes>
      <Route path="/" element={<CampaignListPage />} />
      <Route path="/new" element={<CampaignCreatePage />} />
      <Route path="/play/:id" element={<CampaignPlayPage />} />
    </Routes>
  );
}

export default App;
