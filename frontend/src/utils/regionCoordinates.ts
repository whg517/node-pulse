/**
 * Map node region labels to approximate coordinates for dashboard world map.
 * Unknown regions get a deterministic spread from the region + node id string.
 */

export interface LatLng {
  lat: number
  lng: number
}

const REGION_PRESETS: Array<{ keys: string[]; lat: number; lng: number }> = [
  { keys: ['singapore', '新加坡', 'sgp', 'sg-'], lat: 1.3521, lng: 103.8198 },
  { keys: ['tokyo', 'japan', '日本', '东京', 'osaka', 'jp-'], lat: 35.6762, lng: 139.6503 },
  { keys: ['hong kong', '香港', 'hkg'], lat: 22.3193, lng: 114.1694 },
  { keys: ['taiwan', '台湾', '台北', 'taipei'], lat: 25.033, lng: 121.5654 },
  { keys: ['korea', '韩国', '首尔', 'seoul'], lat: 37.5665, lng: 126.978 },
  { keys: ['china', '中国', '北京', 'beijing', '上海', 'shanghai', '深圳', 'shenzhen', '广州', 'guangzhou'], lat: 31.2304, lng: 121.4737 },
  { keys: ['india', '印度', 'mumbai', 'bangalore', 'delhi'], lat: 28.6139, lng: 77.209 },
  { keys: ['australia', '悉尼', 'sydney', '墨尔本', 'melbourne', 'au-'], lat: -33.8688, lng: 151.2093 },
  { keys: ['london', '英国', 'uk-', 'ireland', 'dublin'], lat: 51.5074, lng: -0.1278 },
  { keys: ['frankfurt', 'germany', '德国', '柏林', 'berlin', 'munich', 'de-', 'eu-central'], lat: 50.1109, lng: 8.6821 },
  { keys: ['amsterdam', 'netherlands', '荷兰'], lat: 52.3676, lng: 4.9041 },
  { keys: ['paris', '法国', 'fr-'], lat: 48.8566, lng: 2.3522 },
  { keys: ['us-east', 'virginia', '纽约', 'new york', 'nyc', 'boston', 'washington'], lat: 40.7128, lng: -74.006 },
  { keys: ['us-west', 'california', 'oregon', '洛杉矶', 'la ', 'san francisco', '硅谷', 'seattle'], lat: 37.7749, lng: -122.4194 },
  { keys: ['texas', 'dallas', 'houston', 'us-south'], lat: 32.7767, lng: -96.797 },
  { keys: ['canada', '加拿大', 'toronto', 'montreal', '温哥华', 'vancouver'], lat: 43.6532, lng: -79.3832 },
  { keys: ['brazil', '巴西', 'são paulo', 'sao paulo'], lat: -23.5505, lng: -46.6333 },
  { keys: ['middle east', '中东', 'dubai', 'uae', 'tel aviv', '以色列'], lat: 25.2048, lng: 55.2708 },
  { keys: ['africa', '南非', 'johannesburg', 'nigeria', 'lagos'], lat: -26.2041, lng: 28.0473 },
]

function hashString(s: string): number {
  let h = 0
  for (let i = 0; i < s.length; i++) {
    h = Math.imul(31, h) + s.charCodeAt(i) | 0
  }
  return Math.abs(h)
}

/**
 * Base coordinates from a free-text region label.
 */
export function estimateRegionBaseCoordinates(region: string): LatLng {
  const r = (region || '').toLowerCase().trim()
  for (const preset of REGION_PRESETS) {
    if (preset.keys.some((k) => r.includes(k))) {
      return { lat: preset.lat, lng: preset.lng }
    }
  }
  const h = hashString(region || 'default')
  const lat = 10 + ((h % 3500) / 3500) * 45
  const lng = -35 + (((h >> 11) % 4000) / 4000) * 70
  return { lat, lng }
}

/**
 * Slight per-node offset so markers in the same region do not stack exactly.
 */
export function scatterAroundBase(base: LatLng, seed: string, spread = 1.8): LatLng {
  const h = hashString(seed)
  const dlat = ((h % 1000) / 1000 - 0.5) * spread
  const dlng = (((h >> 8) % 1000) / 1000 - 0.5) * spread
  return {
    lat: Math.max(-85, Math.min(85, base.lat + dlat)),
    lng: ((base.lng + dlng + 180) % 360) - 180,
  }
}
