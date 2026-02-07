export const formatBytes = (usedBytes, totalBytes) => {
  // 3.4 边界值处理
  if (usedBytes === null || usedBytes === undefined || totalBytes === null || totalBytes === undefined) {
    return { text: 'N/A', percent: 0 }
  }
  
  const used = Number(usedBytes)
  const total = Number(totalBytes)
  
  if (isNaN(used) || isNaN(total) || total === 0) {
      return { text: '0 B / 0 B', percent: 0 }
  }

  const percent = Math.min(Math.round((used / total) * 100), 100)

  // 3.2 统一换算单位：≥1 GB 保留一位小数，<1 GB 使用 MB 并取整
  const formatSize = (bytes) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
    // changing logic to match specific requirement:
    // >= 1 GB (which is 1024*1024*1024 bytes) -> 1 decimal
    // < 1 GB -> MB (integer)
    
    const gb = k * k * k
    if (bytes >= gb) {
        return (bytes / gb).toFixed(1) + ' GB'
    } else {
        const mb = k * k
        // If it's less than 1 MB, maybe show KB? 
        // User says "<1 GB 使用 MB 并取整". Implies even small files show as 0 MB or 1 MB.
        // Let's strictly follow: < 1GB use MB and round to integer.
        // Wait, what if it's really small? 
        // "avoid showing long decimals".
        // Let's convert to MB.
        return Math.round(bytes / mb) + ' MB'
    }
  }

  // However, user example: "120.3 GB / 500 GB"
  // It seems we should format used and total separately but with consistent units?
  // "Unified conversion unit": If I have 500MB used of 2GB total.
  // Should it be "500 MB / 2.0 GB"?
  // Or "0.5 GB / 2.0 GB"?
  // The requirement says "≥1 GB keep 1 decimal, <1 GB use MB". 
  // This likely applies to the *number being displayed*.
  // So 120.3 GB is > 1GB.
  // 500 MB is < 1GB.
  
  return {
    text: `${formatSize(used)} / ${formatSize(total)}`,
    percent
  }
}
