import {
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  Button,
  Icon,
} from '@chakra-ui/react'
import { FiGlobe } from 'react-icons/fi'
import { useTranslation } from 'react-i18next'

const languages = [
  { code: 'en', name: 'English', flag: '🇺🇸' },
  { code: 'es', name: 'Español', flag: '🇪🇸' },
  { code: 'fr', name: 'Français', flag: '🇫🇷' },
]

export function LanguageSwitcher() {
  const { i18n } = useTranslation()

  const currentLanguage = languages.find((lang) => lang.code === i18n.language)

  const changeLanguage = (code) => {
    i18n.changeLanguage(code)
  }

  return (
    <Menu>
      <MenuButton
        as={Button}
        leftIcon={<Icon as={FiGlobe} />}
        variant="ghost"
        size="sm"
      >
        {currentLanguage?.flag} {currentLanguage?.name || 'Language'}
      </MenuButton>
      <MenuList>
        {languages.map((lang) => (
          <MenuItem
            key={lang.code}
            onClick={() => changeLanguage(lang.code)}
            bg={i18n.language === lang.code ? 'blue.50' : 'transparent'}
            fontWeight={i18n.language === lang.code ? 'bold' : 'normal'}
          >
            <span style={{ marginRight: '8px' }}>{lang.flag}</span>
            {lang.name}
          </MenuItem>
        ))}
      </MenuList>
    </Menu>
  )
}
