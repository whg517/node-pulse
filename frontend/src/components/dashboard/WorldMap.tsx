import { useState, useCallback } from 'react'
import { ComposableMap, Geographies, Geography, Marker, ZoomableGroup } from 'react-simple-maps'
import { useTranslation } from 'react-i18next'

export type HealthStatus = 'healthy' | 'warning' | 'critical' | 'offline'

export interface NodeLocation {
  id: string
  name: string
  lat: number
  lng: number
  region: string
  healthStatus: HealthStatus
  avgLatency: number
  packetLoss: number
}

export interface WorldMapProps {
  nodes: NodeLocation[]
  onNodeClick?: (nodeId: string) => void
  height?: string
  className?: string
  isLoading?: boolean
  refreshInterval?: number
}

const GEO_URL = 'https://cdn.jsdelivr.net/npm/world-atlas@2/countries-110m.json'

const statusColors: Record<HealthStatus, string> = {
  healthy: '#22c55e',
  warning: '#f59e0b',
  critical: '#ef4444',
  offline: '#6b7280',
}

export function WorldMap({
  nodes,
  onNodeClick,
  height = '480px',
  className = '',
  isLoading = false,
}: WorldMapProps) {
  const { t } = useTranslation()
  const [hoveredNode, setHoveredNode] = useState<string | null>(null)
  const [zoom, setZoom] = useState(1)
  const [position, setPosition] = useState({ coordinates: [0, 0] as [number, number], zoom: 1 })

  const handleMarkerClick = useCallback(
    (nodeId: string) => {
      onNodeClick?.(nodeId)
    },
    [onNodeClick],
  )

  const handleZoomIn = () => {
    if (position.zoom >= 8) return
    setPosition((pos) => ({ ...pos, zoom: Math.min(pos.zoom * 1.5, 8) }))
    setZoom((z) => Math.min(z * 1.5, 8))
  }

  const handleZoomOut = () => {
    if (position.zoom <= 1) return
    setPosition((pos) => ({ ...pos, zoom: Math.max(pos.zoom / 1.5, 1) }))
    setZoom((z) => Math.max(z / 1.5, 1))
  }

  const handleReset = () => {
    setPosition({ coordinates: [0, 0], zoom: 1 })
    setZoom(1)
  }

  const handleMoveEnd = (newPosition: { coordinates: [number, number]; zoom: number }) => {
    setPosition(newPosition)
    setZoom(newPosition.zoom)
  }

  if (isLoading) {
    return (
      <div className={`rounded-lg border bg-card ${className}`}>
        <div className="flex items-center justify-center py-20 text-muted-foreground">
          {t('common.loading')}
        </div>
      </div>
    )
  }

  return (
    <div className={`rounded-lg border bg-card ${className}`}>
      <div className="relative" style={{ height }}>
        <ComposableMap
          projectionConfig={{ scale: 160 }}
          style={{ width: '100%', height: '100%' }}
        >
          <ZoomableGroup
            zoom={position.zoom}
            center={position.coordinates}
            onMoveEnd={handleMoveEnd}
            minZoom={1}
            maxZoom={8}
            filterZoomEvent={(e: unknown) => {
              const evt = e as Record<string, unknown>
              if (evt.type === 'wheel') return false
              return !evt.ctrlKey && evt.button === 0
            }}
          >
            <Geographies geography={GEO_URL}>
              {({ geographies }) =>
                geographies.map((geo) => (
                  <Geography
                    key={geo.rsmKey}
                    geography={geo}
                    fill="var(--muted)"
                    stroke="var(--border)"
                    strokeWidth={0.3}
                    style={{
                      default: { outline: 'none' },
                      hover: { outline: 'none' },
                      pressed: { outline: 'none' },
                    }}
                  />
                ))
              }
            </Geographies>
            {nodes.map((node) => (
              <Marker
                key={node.id}
                coordinates={[node.lng, node.lat]}
                onClick={() => handleMarkerClick(node.id)}
                onMouseEnter={() => setHoveredNode(node.id)}
                onMouseLeave={() => setHoveredNode(null)}
                style={{ default: { cursor: onNodeClick ? 'pointer' : 'default', outline: 'none' }, hover: { outline: 'none' }, pressed: { outline: 'none' } }}
              >
                <circle
                  r={hoveredNode === node.id ? 6 : 4}
                  fill={statusColors[node.healthStatus]}
                  stroke="#fff"
                  strokeWidth={1}
                  opacity={0.9}
                />
                {hoveredNode === node.id && (
                  <text
                    textAnchor="middle"
                    y={-12}
                    style={{ fontSize: '9px', fontWeight: 600, fill: 'var(--foreground)' }}
                  >
                    {node.name}
                  </text>
                )}
              </Marker>
            ))}
          </ZoomableGroup>
        </ComposableMap>

        {/* Zoom controls */}
        <div className="absolute right-3 top-3 flex flex-col gap-1">
          <button
            onClick={handleZoomIn}
            disabled={zoom >= 8}
            className="flex size-8 items-center justify-center rounded-md border bg-background text-sm font-medium shadow-sm hover:bg-accent disabled:opacity-40"
            aria-label="Zoom in"
          >
            +
          </button>
          <button
            onClick={handleZoomOut}
            disabled={zoom <= 1}
            className="flex size-8 items-center justify-center rounded-md border bg-background text-sm font-medium shadow-sm hover:bg-accent disabled:opacity-40"
            aria-label="Zoom out"
          >
            −
          </button>
          <button
            onClick={handleReset}
            disabled={zoom === 1}
            className="flex size-8 items-center justify-center rounded-md border bg-background text-[10px] font-medium shadow-sm hover:bg-accent disabled:opacity-40"
            aria-label="Reset zoom"
          >
            ⊡
          </button>
        </div>
      </div>

      {/* Legend */}
      <div className="flex items-center justify-center gap-4 px-4 py-2 text-xs text-muted-foreground">
        {(['healthy', 'warning', 'critical', 'offline'] as HealthStatus[]).map((status) => (
          <span key={status} className="flex items-center gap-1">
            <span
              className="inline-block size-2.5 rounded-full"
              style={{ backgroundColor: statusColors[status] }}
            />
            {t(`status.${status}`)}
          </span>
        ))}
      </div>
    </div>
  )
}

export default WorldMap
