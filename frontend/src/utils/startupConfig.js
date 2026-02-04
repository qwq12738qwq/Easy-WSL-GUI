export const startupText = {
  zh: {
    permissionChecking: "正在检测 Windows 管理员权限...",
    permissionSuccess: "管理员权限已获取",
    permissionFail: "权限不足，请以管理员身份运行",
    wslChecking: "正在检查 WSL 功能是否开启...",
    wslSuccess: "WSL 功能已开启",
    wslFail: "WSL 未开启，请启用 Windows Subsystem for Linux 功能",
    retry: "重新检测",
    error: "检测失败，请稍后重试",
    timeout: "检测超时",
    checking: "正在检测..."
  },
  en: {
    permissionChecking: "Checking Windows Administrator Privileges...",
    permissionSuccess: "Administrator Privileges Granted",
    permissionFail: "Insufficient Privileges. Please run as Administrator.",
    wslChecking: "Checking WSL Status...",
    wslSuccess: "WSL Enabled",
    wslFail: "WSL Disabled. Please enable Windows Subsystem for Linux.",
    retry: "Retry",
    error: "Check Failed. Please try again later.",
    timeout: "Check Timeout",
    checking: "Checking..."
  }
};

export const currentLocale = 'zh'; // Can be switched dynamically
