import { createContext, useContext, useEffect, useState } from 'react'

const ColorModeContext = createContext()

export function ColorModeProvider({ children }) {
  const [colorMode, setColorModeState] = useState(() => {
    const saved = localStorage.getItem('chakra-ui-color-mode')
    return saved || 'light'
  })

  useEffect(() => {
    const root = document.documentElement
    root.classList.remove('light', 'dark')
    root.classList.add(colorMode)
    localStorage.setItem('chakra-ui-color-mode', colorMode)
  }, [colorMode])

  const toggleColorMode = () => {
    setColorModeState((prev) => (prev === 'light' ? 'dark' : 'light'))
  }

  const setColorMode = (mode) => {
    setColorModeState(mode)
  }

  return (
    <ColorModeContext.Provider value={{ colorMode, toggleColorMode, setColorMode }}>
      {children}
    </ColorModeContext.Provider>
  )
}

export function useColorMode() {
  const context = useContext(ColorModeContext)
  if (!context) {
    throw new Error('useColorMode must be used within ColorModeProvider')
  }
  return context
}
