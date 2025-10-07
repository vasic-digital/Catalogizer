import { useState } from 'react'
import { Button } from '../ui/Button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/Card'
import { useWizard } from '../../contexts/WizardContext'
import { useConfiguration } from '../../contexts/ConfigurationContext'
import {
  Server,
  FileText,
  HardDrive,
  Globe,
  Folder,
  CheckCircle
} from 'lucide-react'

interface ProtocolOption {
  id: string
  name: string
  description: string
  icon: React.ReactNode
  features: string[]
}

const protocolOptions: ProtocolOption[] = [
  {
    id: 'smb',
    name: 'SMB/CIFS',
    description: 'Windows file sharing protocol for network drives',
    icon: <Server className="h-8 w-8" />,
    features: ['Network discovery', 'Share browsing', 'Authentication', 'Domain support']
  },
  {
    id: 'ftp',
    name: 'FTP',
    description: 'File Transfer Protocol for remote file access',
    icon: <FileText className="h-8 w-8" />,
    features: ['Username/password auth', 'Passive/Active modes', 'Path specification', 'Port configuration']
  },
  {
    id: 'nfs',
    name: 'NFS',
    description: 'Network File System for Unix/Linux file sharing',
    icon: <HardDrive className="h-8 w-8" />,
    features: ['Mount point configuration', 'Version specification', 'Options support', 'Host-based access']
  },
  {
    id: 'webdav',
    name: 'WebDAV',
    description: 'Web-based Distributed Authoring and Versioning',
    icon: <Globe className="h-8 w-8" />,
    features: ['HTTP/HTTPS support', 'Username/password auth', 'Path specification', 'SSL/TLS encryption']
  },
  {
    id: 'local',
    name: 'Local Files',
    description: 'Direct access to local filesystem paths',
    icon: <Folder className="h-8 w-8" />,
    features: ['Base path configuration', 'No authentication', 'Fast access', 'Full permissions']
  }
]

export default function ProtocolSelectionStep({ onNext, onPrevious, canNext, canPrevious }: {
  onNext: () => void
  onPrevious: () => void
  canNext: boolean
  canPrevious: boolean
}) {
  const { setSelectedProtocol } = useConfiguration()
  const [selectedProtocol, setLocalSelectedProtocol] = useState<string | null>(null)

  const handleProtocolSelect = (protocolId: string) => {
    setLocalSelectedProtocol(protocolId)
    setSelectedProtocol(protocolId)
  }

  const handleNext = () => {
    if (selectedProtocol) {
      onNext()
    }
  }

  return (
    <div className="space-y-6">
      <div className="text-center">
        <h2 className="text-2xl font-bold text-gray-900 mb-2">Select Storage Protocol</h2>
        <p className="text-gray-600">
          Choose the protocol for your media storage. Each protocol has different capabilities and requirements.
        </p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {protocolOptions.map((protocol) => (
          <Card
            key={protocol.id}
            className={`cursor-pointer transition-all hover:shadow-md ${
              selectedProtocol === protocol.id
                ? 'ring-2 ring-blue-500 bg-blue-50'
                : 'hover:bg-gray-50'
            }`}
            onClick={() => handleProtocolSelect(protocol.id)}
          >
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className={`p-2 rounded-lg ${
                    selectedProtocol === protocol.id ? 'bg-blue-100 text-blue-600' : 'bg-gray-100 text-gray-600'
                  }`}>
                    {protocol.icon}
                  </div>
                  <div>
                    <CardTitle className="text-lg">{protocol.name}</CardTitle>
                    {selectedProtocol === protocol.id && (
                      <CheckCircle className="h-5 w-5 text-blue-600 mt-1" />
                    )}
                  </div>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <CardDescription className="mb-3">
                {protocol.description}
              </CardDescription>
              <ul className="text-sm text-gray-600 space-y-1">
                {protocol.features.map((feature, index) => (
                  <li key={index} className="flex items-center gap-2">
                    <div className="w-1.5 h-1.5 bg-gray-400 rounded-full" />
                    {feature}
                  </li>
                ))}
              </ul>
            </CardContent>
          </Card>
        ))}
      </div>

      {selectedProtocol && (
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <div className="flex items-start gap-3">
            <CheckCircle className="h-5 w-5 text-blue-600 mt-0.5" />
            <div>
              <h3 className="font-medium text-blue-900">
                {protocolOptions.find(p => p.id === selectedProtocol)?.name} Selected
              </h3>
              <p className="text-blue-700 text-sm mt-1">
                Click "Next" to configure your {protocolOptions.find(p => p.id === selectedProtocol)?.name.toLowerCase()} connection.
              </p>
            </div>
          </div>
        </div>
      )}

      <div className="flex justify-between pt-6">
        <Button
          variant="outline"
          onClick={onPrevious}
          disabled={!canPrevious}
        >
          Previous
        </Button>
        <Button
          onClick={handleNext}
          disabled={!selectedProtocol}
        >
          Next
        </Button>
      </div>
    </div>
  )
}