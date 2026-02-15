
export const getDistroIcon = (name) => {
  const n = name.toLowerCase()
  let iconName = 'UbuntuCoF.png' // 默认值

  if (n.includes('ubuntu')) iconName = 'UbuntuCoF.png'
  else if (n.includes('debian')) iconName = 'Debian.png'
  else if (n.includes('kali'))   iconName = 'Kali-drago.png'
  else if (n.includes('arch'))   iconName = 'Arch.png'
  else if (n.includes('fedora'))   iconName = 'Fedora.png'
  else if (n.includes('almalinux'))   iconName = 'AlmaLinux.png'
  else if (n.includes('opensuse'))   iconName = 'openSUSE.png'
  else if (n.includes('docker'))   iconName = 'Docker.png'

  // 关键：利用 Vite 的动态资源解析
  // 注意：import.meta.url 在被不同文件引用时可能路径不同，
  // 但我们这里是相对路径。如果这是放在 utils/icon.js，
  // 那么 ../assets/icons/ 是正确的 (utils 同级是 assets 的父级 src)
  // 结构: src/utils/icon.js -> src/assets/icons/
  return new URL(`../assets/icons/${iconName}`, import.meta.url).href
}
