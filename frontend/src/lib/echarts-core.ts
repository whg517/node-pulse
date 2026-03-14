/**
 * ECharts on-demand import shim
 *
 * Registers only the components actually used in this app, reducing the
 * vendor-echarts chunk from ~1.1 MB (full bundle) to ~350 KB gzipped.
 *
 * Chart types used:
 *   - LineChart     (LatencyTrendChart, PacketLossChart, TrendChart, ComparisonChart, PerformanceTrendChart)
 *   - GaugeChart    (ProbeSuccessGauge)
 *   - MapChart      (WorldMap)
 *
 * Components used:
 *   - GridComponent, TooltipComponent, LegendComponent
 *   - TitleComponent, DataZoomComponent
 *   - GeoComponent  (WorldMap)
 *   - VisualMapComponent (WorldMap)
 *
 * Renderer: CanvasRenderer (default, best performance)
 */

import * as echarts from 'echarts/core'
import { LineChart, GaugeChart, MapChart } from 'echarts/charts'
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent,
  DataZoomComponent,
  GeoComponent,
  VisualMapComponent,
} from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import { graphic } from 'echarts/core'

echarts.use([
  LineChart,
  GaugeChart,
  MapChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent,
  DataZoomComponent,
  GeoComponent,
  VisualMapComponent,
  CanvasRenderer,
])

export default echarts
export { graphic }
export type { EChartsOption, SeriesOption } from 'echarts'
export type ECharts = ReturnType<typeof echarts.init>
