import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Button } from '../ui/Button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/Card'
import { useWizard } from '../../contexts/WizardContext'
import { useConfiguration } from '../../contexts/ConfigurationContext'
import { TauriService } from '../../services/tauri'
import { NetworkHost } from '../../types'
import {
  Search,
  Wifi,
  Server,
  Loader2,
  RefreshCw,
  AlertCircle,
  CheckCircle2,
  Monitor,
  Network
} from 'lucide-react'

export default function NetworkScanStep() {
  const { setCanNext } = useWizard()
  const { setSelectedHosts } = useConfiguration()
  const [selectedHosts, setSelectedHostsState] = useState<NetworkHost[]>([])
  const [isScanning, setIsScanning] = useState(false)

  const {
    data: hosts,
    isLoading,
    error,
    refetch,
    isError
  } = useQuery({
    queryKey: ['networkScan'],
    queryFn: TauriService.scanNetwork,
    enabled: false, // Start manually
    retry: 2,
    retryDelay: 1000,
  })

  useEffect(() => {
    // Can proceed if we have selected hosts or if user wants to skip
    setCanNext(selectedHosts.length > 0 || (hosts?.length === 0))
  }, [selectedHosts, hosts, setCanNext])

  useEffect(() => {
    // Update configuration context with selected hosts
    setSelectedHosts(selectedHosts)
  }, [selectedHosts, setSelectedHosts])

  const handleScan = async () => {
    setIsScanning(true)
    try {
      await refetch()
    } finally {
      setIsScanning(false)
    }
  }

  const handleHostToggle = (host: NetworkHost) => {
    setSelectedHostsState(prev =>
      prev.some(h => h.ip === host.ip)
        ? prev.filter(h => h.ip !== host.ip)
        : [...prev, host]
    )
  }

  const handleSelectAll = () => {
    if (hosts) {
      const smbHosts = hosts.filter(host =>
        host.open_ports.includes(445) || host.open_ports.includes(139)
      )
      setSelectedHostsState(smbHosts)
    }
  }

  const handleDeselectAll = () => {
    setSelectedHostsState([])
  }

  const smbHosts = hosts?.filter(host =>
    host.open_ports.includes(445) || host.open_ports.includes(139)
  ) || []

  return (
    <div className="space-y-6">
      <div className="text-center space-y-4">
        <div className="mx-auto w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center">
          <Network className="h-8 w-8 text-blue-600" />
        </div>
        <h2 className="text-xl font-bold text-gray-900">Network Discovery</h2>
        <p className="text-gray-600">
          Scan your local network to discover SMB-enabled devices
        </p>
      </div>

      {/* Scan Controls */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Search className="h-5 w-5" />
            Network Scanning
          </CardTitle>
          <CardDescription>
            Click "Start Scan" to discover SMB shares on your local network
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-4">
            <Button
              onClick={handleScan}
              disabled={isLoading || isScanning}
              className="flex items-center gap-2"
            >
              {(isLoading || isScanning) ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Search className="h-4 w-4" />
              )}
              {(isLoading || isScanning) ? 'Scanning...' : 'Start Scan'}
            </Button>

            {hosts && (
              <Button
                variant="outline"
                onClick={handleScan}
                disabled={isLoading || isScanning}
                className="flex items-center gap-2"
              >
                <RefreshCw className="h-4 w-4" />
                Rescan
              </Button>
            )}
          </div>

          {(isLoading || isScanning) && (
            <div className="mt-4 p-4 bg-blue-50 border border-blue-200 rounded-lg">
              <div className="flex items-center gap-2 text-blue-800">
                <Loader2 className="h-4 w-4 animate-spin" />
                <span>Scanning network for SMB devices...</span>
              </div>
              <p className="text-sm text-blue-600 mt-1">
                This may take a few moments depending on your network size
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Error Display */}
      {isError && (
        <Card className="border-red-200 bg-red-50">
          <CardContent className="pt-6">
            <div className="flex items-center gap-2 text-red-800">
              <AlertCircle className="h-5 w-5" />
              <span className="font-medium">Scan Failed</span>
            </div>
            <p className="text-red-700 mt-1">
              {(error as Error)?.message || 'Failed to scan network. Please check your network connection and try again.'}
            </p>
          </CardContent>
        </Card>
      )}

      {/* Results */}
      {hosts && hosts.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Server className="h-5 w-5" />
              Discovered Devices ({hosts.length})
            </CardTitle>
            <CardDescription>
              {smbHosts.length > 0
                ? `Found ${smbHosts.length} device(s) with SMB shares`
                : 'No SMB-enabled devices found'
              }
            </CardDescription>
          </CardHeader>
          <CardContent>
            {smbHosts.length > 0 ? (
              <>
                 <div className="flex gap-2 mb-4">
                   <Button
                     variant="outline"
                     size="sm"
                     onClick={handleSelectAll}
                     disabled={selectedHosts.length === smbHosts.length}
                   >
                     Select All
                   </Button>
                   <Button
                     variant="outline"
                     size="sm"
                     onClick={handleDeselectAll}
                     disabled={selectedHosts.length === 0}
                   >
                     Deselect All
                   </Button>
                 </div>

                <div className="grid grid-cols-1 gap-3">
                  {smbHosts.map((host) => (
                     <div
                       key={host.ip}
                       className={`p-4 border rounded-lg cursor-pointer transition-colors ${
                         selectedHosts.some(h => h.ip === host.ip)
                           ? 'border-blue-500 bg-blue-50'
                           : 'border-gray-200 hover:border-gray-300'
                       }`}
                       onClick={() => handleHostToggle(host)}
                     >
                       <div className="flex items-center justify-between">
                         <div className="flex items-center gap-3">
                           <div className={`w-4 h-4 rounded border-2 flex items-center justify-center ${
                             selectedHosts.some(h => h.ip === host.ip)
                               ? 'border-blue-500 bg-blue-500'
                               : 'border-gray-300'
                           }`}>
                             {selectedHosts.some(h => h.ip === host.ip) && (
                               <CheckCircle2 className="h-3 w-3 text-white" />
                             )}
                           </div>
                          <Monitor className="h-5 w-5 text-gray-500" />
                          <div>
                            <div className="font-medium">
                              {host.hostname || host.ip}
                            </div>
                            <div className="text-sm text-gray-500">
                              {host.hostname && host.ip} • {host.smb_shares.length} share(s)
                            </div>
                          </div>
                        </div>
                        <div className="text-sm text-gray-500">
                          Ports: {host.open_ports.filter(p => [139, 445].includes(p)).join(', ')}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>

                 {selectedHosts.length > 0 && (
                   <div className="mt-4 p-3 bg-green-50 border border-green-200 rounded-lg">
                     <div className="flex items-center gap-2 text-green-800">
                       <CheckCircle2 className="h-4 w-4" />
                       <span className="font-medium">
                         {selectedHosts.length} device(s) selected
                       </span>
                     </div>
                     <p className="text-sm text-green-700 mt-1">
                       Click "Next" to configure SMB connections for the selected devices
                     </p>
                   </div>
                 )}
              </>
            ) : (
              <div className="text-center py-8 text-gray-500">
                <Wifi className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                <p className="text-lg font-medium">No SMB devices found</p>
                <p className="text-sm">
                  Make sure SMB-enabled devices are powered on and accessible on your network
                </p>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {hosts && hosts.length === 0 && !isLoading && !isScanning && (
        <Card className="border-yellow-200 bg-yellow-50">
          <CardContent className="pt-6 text-center">
            <AlertCircle className="h-12 w-12 mx-auto mb-4 text-yellow-500" />
            <h3 className="text-lg font-medium text-yellow-800 mb-2">No Devices Found</h3>
            <p className="text-yellow-700 mb-4">
              No network devices were discovered. This could happen if:
            </p>
            <ul className="text-left text-yellow-700 text-sm space-y-1 max-w-md mx-auto">
              <li>• Devices are powered off or not accessible</li>
              <li>• Firewall is blocking discovery</li>
              <li>• Network configuration prevents scanning</li>
            </ul>
            <p className="text-yellow-600 text-sm mt-4">
              You can skip this step and configure SMB sources manually
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  )
}