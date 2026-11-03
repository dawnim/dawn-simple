import { request } from '@dawnim/admin'

export const fetchStatistics = () =>
  request.get<{
    user: number
    attach: number
    login: number
    operate: number
  }>({ url: '/dashboard/statistics' })
export const fetchLoginChart = () =>
  request.get<{ login_count: number[]; login_date: string[] }>({
    url: '/dashboard/daily'
  })
export const fetchLoginBarChart = () =>
  request.get<{ login_count: number[]; login_month: string[] }>({
    url: '/dashboard/monthly'
  })
