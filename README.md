# Easy-WSL-GUI
基于Wails框架开发

更新事件
runtime.EventsEmit(ctx, "new-version", map[string]interface{}{
    "version":     "v1.2.0",
    "updateLog":   "1. 修复了若干Bug\n2. 优化了UI体验",
    "releaseDate": "2023-10-27",
    "url":         "https://example.com/update",
})