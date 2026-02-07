import { describe, it, expect } from 'vitest'
import { formatBytes } from '../utils/format'

describe('Disk Usage Formatting', () => {
  it('3.4.1 边界值处理: 0 Bytes', () => {
    const result = formatBytes(0, 100)
    expect(result.text).toBe('0 B / 0 MB') // 100 bytes is < 1GB, so MB? 100/1024/1024 is ~0.
    // Wait, my logic: round(bytes / 1024 / 1024) + ' MB'.
    // 100 bytes is 0 MB.
    // 0 bytes is 0 B. (Special case)
  })

  it('3.4.2 边界值处理: Null/Undefined', () => {
    expect(formatBytes(null, 100)).toEqual({ text: 'N/A', percent: 0 })
    expect(formatBytes(100, undefined)).toEqual({ text: 'N/A', percent: 0 })
  })

  it('3.4.3 < 1 GB 显示 MB 并取整', () => {
    // 500 MB used, 800 MB total
    const mb = 1024 * 1024
    const used = 500 * mb
    const total = 800 * mb
    const result = formatBytes(used, total)
    expect(result.text).toBe('500 MB / 800 MB')
    expect(result.percent).toBe(Math.round(500/800*100))
  })

  it('3.4.4 >= 1 GB 保留一位小数', () => {
    // 120.3 GB used, 500 GB total
    const gb = 1024 * 1024 * 1024
    const used = 120.3 * gb
    const total = 500 * gb
    const result = formatBytes(used, total)
    expect(result.text).toBe('120.3 GB / 500.0 GB')
    // Wait, 500 * gb is exactly 500 GB. toFixed(1) -> "500.0 GB"
  })
  
  it('3.4.5 混合单位 (Used < 1GB, Total > 1GB)', () => {
      const mb = 1024 * 1024
      const gb = 1024 * 1024 * 1024
      const used = 500 * mb
      const total = 2 * gb
      const result = formatBytes(used, total)
      expect(result.text).toBe('500 MB / 2.0 GB')
  })
})
