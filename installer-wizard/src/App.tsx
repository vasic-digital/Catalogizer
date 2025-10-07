import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import WizardLayout from './components/layout/WizardLayout'
import WelcomeStep from './components/wizard/WelcomeStep'
import ProtocolSelectionStep from './components/wizard/ProtocolSelectionStep'
import NetworkScanStep from './components/wizard/NetworkScanStep'
import SMBConfigurationStep from './components/wizard/SMBConfigurationStep'
import FTPConfigurationStep from './components/wizard/FTPConfigurationStep'
import NFSConfigurationStep from './components/wizard/NFSConfigurationStep'
import WebDAVConfigurationStep from './components/wizard/WebDAVConfigurationStep'
import LocalConfigurationStep from './components/wizard/LocalConfigurationStep'
import ConfigurationManagementStep from './components/wizard/ConfigurationManagementStep'
import SummaryStep from './components/wizard/SummaryStep'
import { WizardProvider } from './contexts/WizardContext'
import { ConfigurationProvider } from './contexts/ConfigurationContext'

function App() {
  return (
    <ConfigurationProvider>
      <WizardProvider>
        <Router>
          <div className="min-h-screen bg-background">
              <Routes>
                 <Route path="/" element={<WizardLayout />}>
                   <Route index element={<WelcomeStep />} />
                   <Route path="protocol" element={<ProtocolSelectionStep />} />
                   <Route path="scan" element={<NetworkScanStep />} />
                   <Route path="configure-smb" element={<SMBConfigurationStep />} />
                   <Route path="configure-ftp" element={<FTPConfigurationStep />} />
                   <Route path="configure-nfs" element={<NFSConfigurationStep />} />
                   <Route path="configure-webdav" element={<WebDAVConfigurationStep />} />
                   <Route path="configure-local" element={<LocalConfigurationStep />} />
                   <Route path="manage" element={<ConfigurationManagementStep />} />
                   <Route path="summary" element={<SummaryStep />} />
                 </Route>
              </Routes>
          </div>
        </Router>
      </WizardProvider>
    </ConfigurationProvider>
  )
}

export default App