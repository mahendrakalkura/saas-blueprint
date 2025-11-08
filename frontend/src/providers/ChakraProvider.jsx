import { ChakraProvider as BaseChakraProvider, defaultSystem } from '@chakra-ui/react'
import { ColorModeProvider } from './ColorModeContext'

export function ChakraProvider({ children }) {
  return (
    <BaseChakraProvider value={defaultSystem}>
      <ColorModeProvider>
        {children}
      </ColorModeProvider>
    </BaseChakraProvider>
  )
}
