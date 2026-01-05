import { http } from './http'

export type PointsSummaryRes = { total: number }

export type BackendPointsRecord = {
  pointsChange: number
  transactionType: number
  description: string
  transactionTime: string
}

export type PointsRecordsRes = {
  total: number
  hasMore: boolean
  list: BackendPointsRecord[]
}

/** GET /points/summary */
export async function apiPointsSummary(): Promise<PointsSummaryRes> {
  const res = await http.get('/points/summary')
  // 后端可能包一层 {code,message,data}，也可能直接返回 data
  const payload = res.data
  return (payload && typeof payload === 'object' && 'data' in payload) ? payload.data : payload
}

/** GET /points/records?limit=&offset= */
export async function apiPointsRecords(params?: { limit?: number; offset?: number }): Promise<PointsRecordsRes> {
  const res = await http.get('/points/records', { params })
  const payload = res.data
  return (payload && typeof payload === 'object' && 'data' in payload) ? payload.data : payload
}
